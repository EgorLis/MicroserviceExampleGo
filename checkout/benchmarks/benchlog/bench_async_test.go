package benchlog

import (
	"io"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------- базовый sync-логгер (в /dev/null) ----------
var stdLogger = log.New(io.Discard, "", 0)

// ---------- общий воркер для async (экземпляр на каждый бенч) ----------
type asyncLogger struct {
	ch chan string
	wg sync.WaitGroup
	w  *log.Logger
}

func newAsyncLogger(buf int) *asyncLogger {
	al := &asyncLogger{
		ch: make(chan string, buf),
		w:  stdLogger, // пишет в io.Discard
	}
	al.wg.Add(1)
	go func() {
		defer al.wg.Done()
		for msg := range al.ch {
			// имитируем «дорогую» запись логгера
			_ = al.w.Output(2, msg)
		}
	}()
	return al
}

func (al *asyncLogger) Close() {
	close(al.ch)
	al.wg.Wait()
}

// ---------- 1) синхронное логирование ----------
func BenchmarkSyncLog(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stdLogger.Printf("hello %d", i)
	}
}

// ---------- 2) async: неблокирующее (дроп при заполнении буфера) ----------
func BenchmarkAsyncLog_NonBlocking(b *testing.B) {
	al := newAsyncLogger(10_000)
	b.Cleanup(func() { al.Close() })

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		select {
		case al.ch <- "hello":
		default:
			// дропаем — быстрый путь
		}
	}
}

// ---------- 3) async: блокирующее (гарантия доставки) ----------
func BenchmarkAsyncLog_Blocking(b *testing.B) {
	al := newAsyncLogger(10_000)
	b.Cleanup(func() { al.Close() })

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		al.ch <- "hello" // ждём, если буфер полон
	}
}

// ---------- 4) async + batching (склейка по N сообщений) ----------
type batchLogger struct {
	ch        chan string
	wg        sync.WaitGroup
	w         *log.Logger
	batchSize int
	flushTick time.Duration
}

func newBatchLogger(buf, batchSize int, flushTick time.Duration) *batchLogger {
	bl := &batchLogger{
		ch:        make(chan string, buf),
		w:         stdLogger,
		batchSize: batchSize,
		flushTick: flushTick,
	}
	bl.wg.Add(1)
	go func() {
		defer bl.wg.Done()
		t := time.NewTicker(bl.flushTick)
		defer t.Stop()

		batch := make([]string, 0, bl.batchSize)
		flush := func() {
			if len(batch) == 0 {
				return
			}
			_ = bl.w.Output(2, strings.Join(batch, "\n"))
			batch = batch[:0]
		}

		for {
			select {
			case msg, ok := <-bl.ch:
				if !ok {
					flush()
					return
				}
				batch = append(batch, msg)
				if len(batch) >= bl.batchSize {
					flush()
				}
			case <-t.C:
				flush()
			}
		}
	}()
	return bl
}

func (bl *batchLogger) Close() { close(bl.ch); bl.wg.Wait() }

func BenchmarkAsyncLog_Batching(b *testing.B) {
	bl := newBatchLogger(10_000, 100, 2*time.Millisecond)
	b.Cleanup(func() { bl.Close() })

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		select {
		case bl.ch <- "hello":
		default:
			// дроп, чтобы не блокироваться и не портить замер
		}
	}
}

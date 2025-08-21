package benchlog

import (
	"io"
	"log"
	"os"
	"testing"

	"log/slog"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	msg      = "payment created"
	f1k, f1v = "payment_id", "pay_bc342cbc-8da0-4016-80e7-3967557df853"
	f2k, f2v = "merchant_id", "m_129"
	f3k, f3v = "order_id", "o_456"
	f4k, f4v = "amount", "100.00"
	f5k, f5v = "currency", "USD"

	discard = io.Discard
)

func BenchmarkStdLog_Printf(b *testing.B) {
	l := log.New(discard, "", 0) // без времени
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Printf("%s %s=%s %s=%s %s=%s %s=%s %s=%s",
			msg, f1k, f1v, f2k, f2v, f3k, f3v, f4k, f4v, f5k, f5v)
	}
}

func BenchmarkSlog_Text(b *testing.B) {
	h := slog.NewTextHandler(discard, &slog.HandlerOptions{
		// уберём время через ReplaceAttr
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	l := slog.New(h)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info(msg,
			f1k, f1v, f2k, f2v, f3k, f3v, f4k, f4v, f5k, f5v)
	}
}

func BenchmarkSlog_JSON(b *testing.B) {
	h := slog.NewJSONHandler(discard, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	l := slog.New(h)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info(msg,
			f1k, f1v, f2k, f2v, f3k, f3v, f4k, f4v, f5k, f5v)
	}
}

func BenchmarkZap_JSON(b *testing.B) {
	encCfg := zapcore.EncoderConfig{
		MessageKey: "msg", LevelKey: "level",
		// время и пр. убираем
		TimeKey: "", NameKey: "", CallerKey: "", StacktraceKey: "",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapcore.AddSync(discard),
		zapcore.InfoLevel,
	)
	l := zap.New(core) // без caller/stacktrace
	defer l.Sync()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info(msg,
			zap.String(f1k, f1v),
			zap.String(f2k, f2v),
			zap.String(f3k, f3v),
			zap.String(f4k, f4v),
			zap.String(f5k, f5v),
		)
	}
}

func BenchmarkZerolog_JSON(b *testing.B) {
	// zerolog по умолчанию без timestamp, пока не вызвали .Timestamp()
	logger := zerolog.New(discard).Level(zerolog.InfoLevel)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().
			Str(f1k, f1v).
			Str(f2k, f2v).
			Str(f3k, f3v).
			Str(f4k, f4v).
			Str(f5k, f5v).
			Msg(msg)
	}
}

func BenchmarkLogrus_JSON(b *testing.B) {
	l := logrus.New()
	l.SetOutput(discard)
	l.SetFormatter(&logrus.JSONFormatter{DisableTimestamp: true})
	l.SetLevel(logrus.InfoLevel)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{
			f1k: f1v, f2k: f2v, f3k: f3v, f4k: f4v, f5k: f5v,
		}).Info(msg)
	}
}

// sanity-check чтобы исключить влияние переменных окружения
func init() { _ = os.Setenv("TZ", "UTC") }

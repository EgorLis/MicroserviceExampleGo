package v1

type healthResponse struct {
	Status string `json:"status"`
}
type versionResponse struct {
	Version string `json:"version"`
}

type statsResponse struct {
	Processed int `json:"processed"`
	Autorized int `json:"autorized"`
	Declined  int `json:"declined"`
}

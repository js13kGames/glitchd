package metrics

import "runtime"

type RuntimeInfo struct {
	Arch     string `json:"arch"`
	Os       string `json:"os"`
	MaxProcs int    `json:"maxProcs"`
	Version  string `json:"version"`
}

func newRuntimeInfo() RuntimeInfo {
	return RuntimeInfo{
		Arch:     runtime.GOARCH,
		Os:       runtime.GOOS,
		MaxProcs: runtime.GOMAXPROCS(-1),
		Version:  runtime.Version(),
	}
}

package app

import "net/http"

type apiFunc func(http.ResponseWriter, *http.Request) error

type InputData struct {
	Area1Text string `json:"src"`
	Area2Text string `json:"config"`
}

type SimulateRequest struct {
	Code   string `json:"src"`
	Config string `json:"config"`
}

type SimulateResponse struct {
	Message    string   `json:"message"`
	StatusCode int      `json:"status_code"`
	Output     any      `json:"output,omitempty"`
	MemI       []uint32 `json:"mem_i"`
	MemD       []byte   `json:"mem_d"`
	DebugAsm   []string `json:"debug_asm"`
	CGErrors   []string `json:"cg_errors"`
	Error      string   `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error  string   `json:"error"`
	Status int      `json:"status"`
	Errors []string `json:"errors"`
}

type apiError struct {
	Err    string   `json:"err"`
	Status int      `json:"status"`
	Errors []string `json:"errors"`
}

func (e apiError) Error() string {
	return e.Err
}

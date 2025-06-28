package app

import (
	"net/http"
)

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
	Output   any      `json:"output,omitempty"`
	MemI     []string `json:"mem_i"`
	MemD     []string `json:"mem_d"`
	Ast      string   `json:"ast"`
	DebugAsm []string `json:"debug_asm"`
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

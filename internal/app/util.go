package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if e, ok := err.(apiError); ok {
				_ = writeJson(w, e.Status, e)
				return
			}
			_ = writeJson(w, http.StatusInternalServerError, apiError{Err: "internal server", Status: http.StatusInternalServerError})
		}
	}
}

func writeJson(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
func formatDebugAsm(a []string) string {
	return strings.Join(a, "\n")
}

func formatMemD(memD []byte) string {
	var lines []string

	for i, val := range memD {
		if i%4 == 0 {
			lines = append(lines, "_____\n")
		}
		line := fmt.Sprintf("[0x%X|%d]: 0x%02X|%d", i, i, val, val)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
func formatMemI(memI []uint32) string {
	var lines []string

	for i, val := range memI {
		line := fmt.Sprintf("[0x%X|%d]: 0x%08X", i, i, val)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

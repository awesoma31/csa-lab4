package app

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func writeJson(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func formatMemD(memD []byte) []string {
	var lines []string

	for i, val := range memD {
		if i%4 == 0 {
			lines = append(lines, "_____")
		}
		line := fmt.Sprintf("[0x%X|%d]: 0x%02X|%d", i, i, val, val)
		lines = append(lines, line)
	}

	return lines
}
func formatMemI(memI []uint32) []string {
	var lines []string

	for i, val := range memI {
		line := fmt.Sprintf("[0x%X|%d]: 0x%08X", i, i, val)
		lines = append(lines, line)
	}

	return lines
}

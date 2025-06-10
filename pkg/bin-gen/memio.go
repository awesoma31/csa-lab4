package bingen

import (
	"encoding/binary"
	"fmt"
	"os"
)

func SaveInstructionMemory(path string, instrMem []uint32) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create instr bin: %w", err)
	}
	defer f.Close()

	for _, word := range instrMem {
		if err := binary.Write(f, binary.LittleEndian, word); err != nil {
			return fmt.Errorf("write instr word: %w", err)
		}
	}
	return nil
}

func SaveDataMemory(path string, dataMem []byte) error {
	return os.WriteFile(path, dataMem, 0o644)
}

func LoadInstructionMemory(path string) ([]uint32, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw)%4 != 0 {
		return nil, fmt.Errorf("instruction image size (%d) not aligned to 4 bytes", len(raw))
	}
	words := make([]uint32, len(raw)/4)
	for i := range words {
		words[i] = binary.LittleEndian.Uint32(raw[i*4:])
	}
	return words, nil
}

func LoadDataMemory(path string) ([]byte, error) { return os.ReadFile(path) }

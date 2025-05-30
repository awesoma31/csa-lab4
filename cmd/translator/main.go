package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
	"github.com/sanity-io/litter"
)

func main() {
	flags := &flags{}
	flags.parseFlags()

	sourceBytes, err := os.ReadFile(flags.InPath)
	if err != nil {
		fmt.Printf("couldn't resolve %s: %v\n", flags.InPath, err)
		os.Exit(1)
	}

	program, parseErrors := parser.Parse(string(sourceBytes))
	if len(parseErrors) > 0 {
		fmt.Println("Ошибки парсера:")
		for _, err := range parseErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	fmt.Println("-------------------AST----------------------")
	litter.Dump(program)

	cg := codegen.NewCodeGenerator()
	instructionMemory, dataMemory, debugAssembly, cgErrors := cg.Generate(program)
	if len(cgErrors) > 0 {
		for _, e := range cgErrors {
			fmt.Println(e)
		}
		os.Exit(1)
	}

	fmt.Println("-------------------debugAssembly----------------------")
	for _, val := range debugAssembly {
		fmt.Println(val)
	}

	fmt.Println("-------------------instructionMemory----------------------")
	for i, instr := range instructionMemory {
		// fmt.Println(instr)
		fmt.Println(fmt.Sprintf("[0x%X]:", i), instr)
	}

	fmt.Println("-------------------dataMemory----------------------")
	for i, val := range dataMemory {
		fmt.Println(fmt.Sprintf("[0x%X]:", i), val)
	}
}

type flags struct {
	InPath     string
	OutDirPath string
}

func (f *flags) parseFlags() {
	flag.StringVar(&f.InPath, "in", "", "файл исходной программы (*.lang)")
	flag.StringVar(&f.OutDirPath, "out", "out", "каталог с результатами компиляции")
	flag.Parse()

	if f.InPath == "" {
		fmt.Println("usage: translator -in=source.lang [-out dir]")
		os.Exit(1)
	}
}

// writeBinaryFile записывает слайс uint32 в бинарный файл.
// Использует LittleEndian для записи, что является распространенным выбором для большинства систем.
func writeBinaryFile(filename string, data []uint32) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	// Гарантируем закрытие файла при выходе из функции
	defer file.Close()

	for _, word := range data {
		// Записываем каждое 32-битное слово (4 байта)
		// Используем binary.Write и указываем порядок байтов (LittleEndian)
		err := binary.Write(file, binary.LittleEndian, word)
		if err != nil {
			return err
		}
	}
	return nil
}

// getPrettyJson форматирует JSON байты в удобочитаемую строку с отступами.
func getPrettyJson(in []byte) (string, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(in), "", " ")
	if err != nil {
		return "", err
	}
	return prettyJson.String(), nil
}

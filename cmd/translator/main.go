package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
)

func main() {
	// var codeFile string
	// if len(os.Args) > 1 {
	// 	codeFile = os.Args[1]
	// }
	// sourceBytes, err := os.ReadFile(codeFile)
	// if err != nil {
	// 	sourceBytes, _ = os.ReadFile("examples/00.lang")
	// }
	// source := string(sourceBytes)
	// astTree := parser.Parse(source)
	//
	// astJson, _ := json.Marshal(astTree)
	// var prettyJson bytes.Buffer
	// err = json.Indent(&prettyJson, []byte(astJson), "", " ")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// prettyJsonAst, _ := getPrettyJson(astJson)
	// // fmt.Println(string(astJson))
	// fmt.Println(prettyJsonAst)
	// litter.Dump(astTree)
	// fmt.Println(astTree)
	// fmt.Println(string(prettyJson.String()))
	// }

	source := `
        let a = 10;
        let b = "hello_world";
        let c = a + 5;
        if (c > 10) {
            let d = 20;
            c = c + d;
        } else {
            c = c - 5;
        }
        // print(c); // Assuming 'print' is a built-in function or external call
        // let s = "test_string";
        // func my_func(x, y) {
        //     let z = x + y;
        //     return z;
        // }
        // let result = my_func(c, 100);
    `

	// 1. Парсинг AST
	program, parseErrors := parser.Parse(source)
	if len(parseErrors) > 0 {
		fmt.Println("Parser Errors:")
		for _, err := range parseErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	// 2. Кодогенерация
	cg := codegen.NewCodeGenerator()
	instructionMemory, dataMemory, debugAssembly, cgErrors := cg.Generate(program)

	if len(cgErrors) > 0 {
		fmt.Println("Code Generation Errors:")
		for _, err := range cgErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	// 3. Вывод результатов
	outputDir := "output"
	os.MkdirAll(outputDir, os.ModePerm) // Создаем директорию, если ее нет

	// Сохраняем отладочный ассемблер
	debugAsmPath := filepath.Join(outputDir, "output.asm")
	err := os.WriteFile(debugAsmPath, []byte(strings.Join(debugAssembly, "\n")), 0644)
	if err != nil {
		fmt.Println("Error writing debug assembly:", err)
	} else {
		fmt.Printf("Debug assembly saved to %s\n", debugAsmPath)
	}

	// Сохраняем бинарные файлы памяти инструкций
	instrBinPath := filepath.Join(outputDir, "instruction_memory.bin")
	err = writeBinaryFile(instrBinPath, instructionMemory)
	if err != nil {
		fmt.Println("Error writing instruction memory binary:", err)
	} else {
		fmt.Printf("Instruction memory binary saved to %s\n", instrBinPath)
	}

	// Сохраняем бинарные файлы памяти данных
	dataBinPath := filepath.Join(outputDir, "data_memory.bin")
	err = writeBinaryFile(dataBinPath, dataMemory)
	if err != nil {
		fmt.Println("Error writing data memory binary:", err)
	} else {
		fmt.Printf("Data memory binary saved to %s\n", dataBinPath)
	}
}

// writeBinaryFile helper function (copied from previous response for completeness)
func writeBinaryFile(filename string, data []uint32) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Используйте LittleEndian или BigEndian в зависимости от вашей ISA
	// Для большинства современных систем LittleEndian
	writer := binary.NewWriter(file)
	for _, word := range data {
		// Записываем 32-битное слово (4 байта)
		err := writer.WriteUint32(word, binary.LittleEndian) // Предположим Little-Endian
		if err != nil {
			return err
		}
	}
	return nil
}

func getPrettyJson(in []byte) (string, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(in), "", " ")
	if err != nil {
		return "", err
	}
	return prettyJson.String(), nil
}

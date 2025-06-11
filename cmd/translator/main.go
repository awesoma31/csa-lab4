package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator"
)

func main() {
	flags := &flags{}
	flags.parseFlags()
	in := flags.InPath
	out := flags.OutDirPath
	dbg := flags.Debug
	flag.Parse()

	if in == "" {
		log.Fatal("usage: translator -in prog.lang [-out dir] [-debug]")
	}

	if _, _, err := translator.Run(translator.Options{
		SrcPath: in, OutDir: out,
		Debug:  dbg,
		LogDir: "logs",
	}); err != nil {
		log.Fatal(err)
	}
}

type flags struct {
	InPath     string
	OutDirPath string
	Debug      bool
}

func (f *flags) parseFlags() {
	flag.StringVar(&f.InPath, "in", "", "source file path")
	flag.StringVar(&f.OutDirPath, "o", "bin", "directory to save bin files ")
	flag.BoolVar(&f.Debug, "debug", false, "print dumps to stdout")
	flag.Parse()

	if f.InPath == "" {
		fmt.Println("usage: translator -in=source-path [-out dir] [-debug]")
		os.Exit(1)
	}
}

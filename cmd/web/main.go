package main

import (
	"log"

	"github.com/awesoma31/csa-lab4/config"
	"github.com/awesoma31/csa-lab4/internal/app"
)

func main() {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	app.Run(cfg)
}

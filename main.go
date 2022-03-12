package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/moviedb/scraper/pkg/config"
	"github.com/moviedb/scraper/pkg/server"
)

func main() {

	env := flag.String("env", ".env", ".env path")
	flag.Parse()

	cfg, err := config.LoadConfig(env)
	if err != nil {
		log.Fatalf("unable to load config: %+v", err)
	}

	srv := server.NewServer(
		cfg,
	)

	srv.Execute()
	fmt.Println("End process")
}

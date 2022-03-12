package server

import (
	"fmt"

	"github.com/moviedb/scraper/pkg/config"
	"github.com/moviedb/scraper/pkg/streaming"
)

type Server struct {
	Config  config.Config
	Scraper streaming.Scraper
}

func NewServer(
	cfg config.Config,
) Server {
	srv := Server{
		Config: cfg,
	}

	return srv
}

func (s Server) Execute() {

	fmt.Println("Init process")

	var scraper streaming.Scraper

	scraper = streaming.Netflix{
		Config: s.Config,
	}
	scraper.ExecuteProcess()

	scraper = streaming.Hbo{
		Config: s.Config,
	}
	scraper.ExecuteProcess()

	fmt.Println("End process")
}

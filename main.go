package main

import (
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/nitishm/go-rejson"

	scraper "github.com/xegea/scraper/streaming"
)

var conn redis.Conn

func main() {

	log.Println("Init process")

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file")
	}

	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	country := os.Getenv("COUNTRY")

	conn, err = redis.Dial("tcp", host,
		redis.DialPassword(password),
	)
	if err != nil {
		log.Fatalf("Failed to communicate to redis-server @ %v", err)
	}
	defer conn.Close()

	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	// scraper.ProcessNetflixGenres(109012, country)
	// scraper.ExecuteNetflixProcess(rh, 0, country)

	scraper.ExecuteHboProcess(rh, country)
	//scraper.Whats_on_netflix()
}

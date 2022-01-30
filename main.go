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

	path, err := os.Getwd()
	if err != nil {
		log.Println("Error getting current path ", err)
	}
	path = path + "/.env"

	err = godotenv.Load(path)
	if err != nil {
		log.Printf("Error loading %s file", path)
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

	//scraper.ProcessNetflixGenres(109012, country)
	scraper.ExecuteNetflixProcess(rh, 0, country)

	//scraper.ExecuteHboProcess(rh, country)
	//scraper.Whats_on_netflix()
}

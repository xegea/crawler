package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/nitishm/go-rejson"

	scraper "github.com/xegea/scraper/streaming"
)

var conn redis.Conn

func main() {

	logFile, err := os.OpenFile("current.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	envPath := flag.String("env", ".env", ".env path")
	flag.Parse()

	err = godotenv.Load(*envPath)
	if err != nil {
		log.Printf("Error loading %s file", *envPath)
	}

	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	country := os.Getenv("COUNTRY")

	useRedis := false
	useRedisEnv := os.Getenv("USE_REDIS")
	if useRedisEnv == "TRUE" {
		useRedis = true
	}

	conn, err = redis.Dial("tcp", host,
		redis.DialPassword(password),
	)
	if err != nil {
		log.Fatalf("Failed to communicate to redis-server @ %v", err)
	}
	defer conn.Close()

	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	fmt.Println("Init process")

	//scraper.ProcessNetflixGenres(109012, country)
	scraper.ExecuteNetflixProcess(rh, 0, country, useRedis)

	//scraper.ExecuteHboProcess(rh, country)
	//scraper.Whats_on_netflix()

	fmt.Println("End process")
}

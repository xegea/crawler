package main

import (
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/nitishm/go-rejson"

	scraper "github.com/xegea/scraper/streaming"
)

func main() {

	log.Println("Init process")

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file")
	}

	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")

	conn, err := redis.Dial("tcp", host,
		redis.DialPassword(password),
	)
	if err != nil {
		log.Fatalf("Failed to communicate to redis-server @ %v", err)
	}

	// defer func() {
	// 	_, err = conn.Do("FLUSHALL")
	// 	err = conn.Close()
	// 	if err != nil {
	// 		log.Fatalf("Failed to communicate to redis-server @ %v", err)
	// 	}
	// }()
	defer conn.Close()

	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	log.Println(os.Getenv("PORT"))

	scraper.ProcessGenres(109012)
	scraper.ExecuteProcess(rh, 0)
	//scraper.Whats_on_netflix()

	log.Println("End process")
}

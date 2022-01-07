package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
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

	conn, err = redis.Dial("tcp", host,
		redis.DialPassword(password),
	)
	if err != nil {
		log.Fatalf("Failed to communicate to redis-server @ %v", err)
	}
	defer conn.Close()

	handleRequests()

	log.Println("End process")
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", homeHandler)

	port := ":" + os.Getenv("PORT")
	if os.Getenv("PORT") == "" {
		port = ":8080"
	}

	fmt.Println("Listen on port", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	w.WriteHeader(200)

	scraper.ProcessGenres(109012)
	scraper.ExecuteProcess(rh, 0)
	//scraper.Whats_on_netflix()
}

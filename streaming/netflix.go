package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nitishm/go-rejson"
)

type NetflixContent struct {
	Context         string               `json:"@context"`
	Type            string               `json:"@type"`
	Name            string               `json:"name"`
	ItemListElement []NetflixListElement `json:"itemListElement"`
}

type NetflixListItem struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type NetflixListElement struct {
	Type     string          `json:"@type"`
	Position int             `json:"position"`
	Item     NetflixListItem `json:"item"`
}

type NetflixContentDetail struct {
	Context       string `json:"@context"`
	Type          string `json:"@type"`
	URL           string `json:"url"`
	ContentRating string `json:"contentRating"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Genre         string `json:"genre"`
	Image         string `json:"image"`
	DateCreated   string `json:"dateCreated"`
	Trailer       []struct {
		Type         string    `json:"@type"`
		Name         string    `json:"name"`
		Description  string    `json:"description"`
		ThumbnailURL string    `json:"thumbnailUrl"`
		Duration     string    `json:"duration"`
		ContentURL   string    `json:"contentUrl"`
		UploadDate   time.Time `json:"uploadDate"`
	} `json:"trailer"`
	Actors []struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"actors"`
	Creator  []interface{} `json:"creator"`
	Director []struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"director"`
	NumberOfSeasons int    `json:"numberOfSeasons"`
	StartDate       string `json:"startDate"`
}

func ExecuteNetflixProcess(rh *rejson.Handler, initialGenre int, country string, useRedis bool) {

	genresUrl := resolveGenresUrl(country)

	for _, v := range getGenres() {

		i := 0
		fmt.Sscan(v, &i)
		if i < initialGenre {
			continue
		}

		fmt.Println(genresUrl + fmt.Sprint(v))

		b, err := httpGet(genresUrl + fmt.Sprint(v))
		if err != nil {
			fmt.Printf("Failed to http get %s\n", genresUrl+fmt.Sprint(v))
			continue
		}

		var netflixContent *NetflixContent
		if err := json.Unmarshal(extractJson(b), &netflixContent); err != nil {
			fmt.Printf("Failed to Unmarshall %s\n", genresUrl+fmt.Sprint(v))
			continue
		}

		buildNetflixContent(netflixContent, rh, country, useRedis)
	}
}

func ProcessNetflixGenres(genresMax int, country string) {

	genres := getGenres()
	genresUrl := resolveGenresUrl(country)

	initialGenre, err := strconv.Atoi(genres[len(genres)-1])
	if err != nil {
		log.Fatal(err)
	}

	for i := initialGenre + 1; i <= genresMax; i++ {

		req, _ := http.NewRequest("GET", genresUrl+fmt.Sprint(i), nil)
		client := http.Client{}

		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			appendToFile(os.Getenv("GENRESFILE"), fmt.Sprint(i))
			fmt.Print("*")
		}

		fmt.Println(i)
	}
}

func buildNetflixContent(nc *NetflixContent, rh *rejson.Handler, country string, useRedis bool) {

	for i, v := range nc.ItemListElement {

		redisKey := buildNetflixRedisKey(v.Item.URL)

		if useRedis {
			value, err := rh.JSONGet(redisKey, ".")
			if err != nil {
				fmt.Printf("Failed to JSONGet %s\n", redisKey)
				continue
			}
			if value != nil {
				fmt.Printf("%s --> found\n", v.Item.URL)
				continue
			}
		} else {
			value, err := httpGet("http://87.106.124.190/movie/" + redisKey)
			if err != nil {
				fmt.Printf("Failed to http get %s - error: %s\n", "http://87.106.124.190/movie/"+redisKey, err)
			}
			if value != nil {
				//fmt.Printf("%s --> found\n", v.Item.URL)
				continue
			}
		}

		b, err := httpGet(v.Item.URL)
		if err != nil {
			fmt.Printf("Failed to http get %s - error: %s\n", v.Item.URL, err)
			continue
		}

		var detail *NetflixContentDetail
		if err := json.Unmarshal(extractJson(b), &detail); err != nil {
			fmt.Printf("Failed to Unmarshall %s\n", v.Item.URL)
			time.Sleep(3 * time.Second)
			continue
		}

		var movie Movie
		movie.Title = make(map[string]string)
		movie.Title[country] = detail.Name
		movie.Url = detail.URL
		movie.ContentRating = detail.ContentRating
		movie.Type = detail.Type
		movie.Description = make(map[string]string)
		movie.Description[country] = detail.Description
		movie.Genre = detail.Genre
		movie.Image = detail.Image
		movie.ReleaseDate = unixTimestamp(detail.DateCreated)

		for _, act := range detail.Actors {
			movie.Actors = append(movie.Actors, act.Name)
		}

		for _, dir := range detail.Director {
			movie.Director = append(movie.Director, dir.Name)
		}

		for _, tr := range detail.Trailer {
			var trailer Trailer
			trailer.Url = tr.ContentURL
			trailer.Name = make(map[string]string)
			trailer.Name[country] = tr.Name
			trailer.Description = make(map[string]string)
			trailer.Description[country] = tr.Description
			trailer.ThumbnailUrl = tr.ThumbnailURL
			movie.Trailer = append(movie.Trailer, trailer)
		}

		if useRedis {
			_, err = rh.JSONSet(redisKey, ".", movie)
			if err != nil {
				fmt.Printf("Failed to JSONSet %s\n", redisKey)
				continue
			}
		} else {
			json_data, err := json.Marshal(movie)
			if err != nil {
				fmt.Printf("Failed to Marshall movie")
				continue
			}

			_, err = http.Post("http://87.106.124.190/movie", "application/json", bytes.NewBuffer(json_data))
			if err != nil {
				fmt.Printf("Failed to http post %s\n", redisKey)
				continue
			}
		}

		fmt.Printf("%d: %s --> %s\n", i, movie.Url, movie.Title[country])

		time.Sleep(3 * time.Second)
	}
}

func resolveGenresUrl(country string) string {

	switch country {
	case "ES":
		{
			return "https://www.netflix.com/es/browse/genre/"
		}
	case "US":
		{
			return "https://www.netflix.com/browse/genre/"
		}
	case "GB":
		{
			return "https://www.netflix.com/gb/browse/genre/"
		}
	case "DE":
		{
			return "https://www.netflix.com/de-de/browse/genre/"
		}
	}

	return ""
}

func getGenres() []string {

	genresFile := os.Getenv("GENRESFILE")

	b, err := ioutil.ReadFile(genresFile)
	if err != nil {
		fmt.Printf("Failed to read %s\n", genresFile)
	}

	genres := strings.Split(string(b), ",")
	return genres
}

func buildNetflixRedisKey(movieUrl string) string {

	var redisKey string
	redisKey = strings.Replace(movieUrl, "https://www.netflix.com", "netflix", 1)
	redisKey = strings.Replace(redisKey, "/es/title/", ":es-es:", 1)
	redisKey = strings.Replace(redisKey, "/es-es/title/", ":es-es:", 1)
	redisKey = strings.Replace(redisKey, "/en-us/title/", ":en-us:", 1)
	redisKey = strings.Replace(redisKey, "/de-de/title/", ":de-de:", 1)
	redisKey = strings.Replace(redisKey, "/de/title/", ":de-de:", 1)
	redisKey = strings.Replace(redisKey, "/gb/title/", ":en-gb:", 1)
	redisKey = strings.Replace(redisKey, "/title/", ":en-us:", 1)

	if strings.Contains(redisKey, "es-en") {
		log.Fatalf("Incorrect key %s", redisKey)
	}

	if strings.Contains(redisKey, "de-en") {
		log.Fatalf("Incorrect key %s", redisKey)
	}

	return redisKey
}

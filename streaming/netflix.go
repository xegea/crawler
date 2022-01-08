package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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

type Content struct {
	Title         string
	Url           string
	ContentRating string
	Type          string
	Description   string
	Genre         string
	Image         string
	ReleaseDate   int64
	Director      []string
	Actors        []string
	Trailer       []Trailer
}

type Trailer struct {
	Name         string
	Description  string
	Url          string
	ThumbnailUrl string
}

const genresFile string = "streaming/netflixgenres.txt"

var genresUrl string

func ExecuteProcess(rh *rejson.Handler, initialGenre int, country string) {

	genresUrl = resolveGenresUrl(country)

	for _, v := range getGenres() {

		i := 0
		fmt.Sscan(v, &i)
		if i < initialGenre {
			continue
		}

		fmt.Println(genresUrl + fmt.Sprint(v))

		b, err := httpGet(genresUrl + fmt.Sprint(v))
		if err != nil {
			log.Fatalf("Failed to http get %s", genresUrl+fmt.Sprint(v))
		}

		var netflixContent *NetflixContent
		if err := json.Unmarshal(extractJson(b), &netflixContent); err != nil {
			log.Fatalf("Failed to Unmarshall %s", genresUrl+fmt.Sprint(v))
		}

		buildContent(netflixContent, rh)
	}
}

func ProcessGenres(genresMax int) {

	genres := getGenres()

	initialGenre, err := strconv.Atoi(genres[len(genres)-1])
	if err != nil {
		log.Fatal(err)
	}

	for i := initialGenre + 1; i <= genresMax; i++ {

		req, err := http.NewRequest("GET", genresUrl+fmt.Sprint(i), nil)
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
			appendToFile(genresFile, fmt.Sprint(i))
			fmt.Print("*")
		}

		fmt.Println(i)
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
			return "https://www.netflix.com/es-US/browse/genre/"
		}
	case "GB":
		{
			return "https://www.netflix.com/en-GB/browse/genre/"
		}
	case "DE":
		{
			return "https://www.netflix.com/de-DE/browse/genre/"
		}
	}

	return ""
}

func getGenres() []string {
	b, err := ioutil.ReadFile(genresFile)
	if err != nil {
		log.Fatalf("Failed to read %s", genresFile)
	}

	genres := strings.Split(string(b), "\n")
	return genres
}

func buildContent(nc *NetflixContent, rh *rejson.Handler) {

	for i, v := range nc.ItemListElement {

		redisKey := buildRedisKey(v.Item.URL)
		redisValue, err := rh.JSONGet(redisKey, ".")
		if err != nil {
			log.Printf("Failed to JSONGet %s", redisKey)
		}
		if redisValue != nil {
			//fmt.Printf("%s --> found\n", v.Item.URL)
			continue
		}

		b, err := httpGet(v.Item.URL)
		if err != nil {
			log.Printf("Failed to http get %s", v.Item.URL)
			continue
		}

		var detail *NetflixContentDetail
		if err := json.Unmarshal(extractJson(b), &detail); err != nil {
			log.Printf("Failed to Unmarshall %s", v.Item.URL)
			continue
		}

		var content Content
		content.Title = detail.Name
		content.Url = detail.URL
		content.ContentRating = detail.ContentRating
		content.Type = detail.Type
		content.Description = detail.Description
		content.Genre = detail.Genre
		content.Image = detail.Image
		content.ReleaseDate = parseDate(detail.DateCreated)

		for _, dir := range detail.Actors {
			content.Actors = append(content.Actors, dir.Name)
		}

		for _, dir := range detail.Director {
			content.Director = append(content.Director, dir.Name)
		}

		for _, tr := range detail.Trailer {
			var trailer Trailer
			trailer.Url = tr.ContentURL
			trailer.Name = tr.Name
			trailer.Description = tr.Description
			trailer.ThumbnailUrl = tr.ThumbnailURL
			content.Trailer = append(content.Trailer, trailer)
		}

		_, err = rh.JSONSet(redisKey, ".", content)
		if err != nil {
			log.Printf("Failed to JSONSet %s", redisKey)
			continue
		}

		fmt.Printf("%d: %s --> %s\n", i, content.Url, content.Title)

		time.Sleep(3 * time.Second)
	}
}

func httpGet(url string) ([]byte, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Host = "www.netflix.com"

	req.Header = http.Header{
		"Content-Type": []string{"application/json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func appendToFile(file string, s string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("\n%s", s)); err != nil {
		log.Println(err)
	}
}

func extractJson(b []byte) []byte {
	left := `<script type="application/ld+json">`
	right := `</script>`

	rx := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(left) + `(.*?)` + regexp.QuoteMeta(right))
	matches := rx.FindAllStringSubmatch(string(b), -1)

	var json []byte
	for _, v := range matches {
		return []byte(v[1])
	}

	return json
}

func parseDate(date string) int64 {

	var year, day int
	var month time.Month

	fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
}

func buildRedisKey(contentUrl string) string {

	var redisKey string
	redisKey = strings.Replace(contentUrl, "https://www.netflix.com", "hbo", 1)
	redisKey = strings.Replace(redisKey, "/es/title/", ":es-es:", 1)
	redisKey = strings.Replace(redisKey, "/es-es/title/", ":es-es:", 1)
	redisKey = strings.Replace(redisKey, "/en-us/title/", ":en-us:", 1)
	redisKey = strings.Replace(redisKey, "/title/", ":en-us:", 1)
	redisKey = strings.Replace(redisKey, "/de-de/title/", ":de-de:", 1)
	redisKey = strings.Replace(redisKey, "/de/title/", ":de-de:", 1)

	return redisKey
}

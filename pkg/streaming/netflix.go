package streaming

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/moviedb/scraper/pkg/config"
)

//go:embed netflixgenres.txt
var netflixGenres embed.FS

type Netflix struct {
	Config config.Config
}

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

func (n Netflix) ExecuteProcess() error {

	config := n.Config
	genresUrls := resolveGenresUrl(config.Country)
	genres, err := getNetflixGenres()
	if err != nil {
		return err
	}

	for _, v := range genres {

		for _, url := range genresUrls {
			fmt.Println(url + fmt.Sprint(v))

			b, err := httpGet(url+fmt.Sprint(v), "")
			if err != nil {
				fmt.Printf("Failed to http get %s\n", url+fmt.Sprint(v))
				continue
			}

			var netflixContent *NetflixContent
			if err := json.Unmarshal(extractJson(b), &netflixContent); err != nil {
				fmt.Printf("Failed to Unmarshall %s\n", url+fmt.Sprint(v))
				continue
			}

			buildNetflixContent(netflixContent, config)
		}
	}

	return nil
}

func buildNetflixContent(nc *NetflixContent, config config.Config) {

	for i, v := range nc.ItemListElement {

		key := buildNetflixRedisKey(v.Item.URL)

		if config.ProcessMode == "NewOnly" {
			value, err := httpGet(config.ApiUrl+"/movie/"+key, config.ApiKey)
			if err != nil {
				fmt.Printf("Failed to http get %s - error: %s\n", config.ApiUrl+"/movie/"+key, err)
			}
			if value != nil && err == nil {
				//fmt.Printf("%s --> found\n", v.Item.URL)
				continue
			}
		}

		b, err := httpGet(v.Item.URL, config.ApiKey)
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
		movie.Title = detail.Name
		movie.Url = detail.URL
		movie.ContentRating = detail.ContentRating
		movie.Type = detail.Type
		movie.Description = detail.Description
		movie.Genre = detail.Genre
		movie.Image = detail.Image
		movie.ReleaseDate = unixTimestamp(detail.DateCreated)
		movie.Updated = time.Now().Unix()

		for _, act := range detail.Actors {
			movie.Actors = append(movie.Actors, act.Name)
		}

		for _, dir := range detail.Director {
			movie.Director = append(movie.Director, dir.Name)
		}

		for _, tr := range detail.Trailer {
			var trailer Trailer
			trailer.Url = tr.ContentURL
			trailer.Name = tr.Name
			trailer.Description = tr.Description
			trailer.ThumbnailUrl = tr.ThumbnailURL
			movie.Trailer = append(movie.Trailer, trailer)
		}

		json_data, err := json.Marshal(movie)
		if err != nil {
			fmt.Printf("Failed to Marshall movie")
			continue
		}

		err = httpPost(config.ApiUrl+"/movie", bytes.NewBuffer(json_data), config.ApiKey)
		if err != nil {
			fmt.Printf("Failed to http post %s\n", key)
			continue
		}

		fmt.Printf("%d: %s --> %s\n", i, movie.Url, movie.Title)

		time.Sleep(3 * time.Second)
	}
}

func getNetflixGenres() ([]string, error) {
	b, err := netflixGenres.ReadFile("netflixgenres.txt")
	if err != nil {
		fmt.Printf("Failed to read %s\n", "netflixgenres.txt")
		return nil, err
	}

	return strings.Split(string(b), ","), nil
}

func resolveGenresUrl(country string) []string {

	switch country {
	case "ES":
		{
			return []string{"https://www.netflix.com/es/browse/genre/", "https://www.netflix.com/es-en/browse/genre/"}
		}
	case "US":
		{
			return []string{"https://www.netflix.com/browse/genre/"}
		}
	case "GB":
		{
			return []string{"https://www.netflix.com/gb/browse/genre/"}
		}
	case "DE":
		{
			return []string{"https://www.netflix.com/de-de/browse/genre/"}
		}
	}

	return nil
}

func buildNetflixRedisKey(movieUrl string) string {

	var key string
	key = strings.Replace(movieUrl, "https://www.netflix.com", "netflix", 1)
	key = strings.Replace(key, "/es/title/", ":es-es:", 1)
	key = strings.Replace(key, "/es-es/title/", ":es-es:", 1)
	key = strings.Replace(key, "/es-en/title/", ":es-en:", 1)
	key = strings.Replace(key, "/en-us/title/", ":en-us:", 1)
	key = strings.Replace(key, "/de-de/title/", ":de-de:", 1)
	key = strings.Replace(key, "/de/title/", ":de-de:", 1)
	key = strings.Replace(key, "/gb/title/", ":en-gb:", 1)
	key = strings.Replace(key, "/title/", ":en-us:", 1)

	if strings.Contains(key, "de-en") {
		log.Fatalf("Incorrect key %s", key)
	}

	return key
}

// func ProcessNetflixGenres(genresMax int, country string) {

// 	genres := getGenres()
// 	genresUrl := resolveGenresUrl(country)

// 	initialGenre, err := strconv.Atoi(genres[len(genres)-1])
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for i := initialGenre + 1; i <= genresMax; i++ {

// 		req, _ := http.NewRequest("GET", genresUrl+fmt.Sprint(i), nil)
// 		client := http.Client{}

// 		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
// 			return http.ErrUseLastResponse
// 		}

// 		resp, err := client.Do(req)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer resp.Body.Close()

// 		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
// 			appendToFile(os.Getenv("GENRESFILE"), fmt.Sprint(i))
// 			fmt.Print("*")
// 		}

// 		fmt.Println(i)
// 	}
// }

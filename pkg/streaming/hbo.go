package streaming

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/moviedb/scraper/pkg/config"
)

type Hbo struct {
	Config config.Config
}

type HboContentDetail struct {
	Context       string   `json:"@context"`
	Type          string   `json:"@type"`
	URL           string   `json:"url"`
	Name          string   `json:"name"`
	Image         string   `json:"image"`
	Genre         []string `json:"genre"`
	ContentRating []string `json:"contentRating"`
	Description   string   `json:"description"`
	Actor         []struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"actor"`
}

func (h Hbo) ExecuteProcess() error {

	config := h.Config
	var urlList []string
	for _, v := range []string{"movies", "series"} {
		url := resolveContentUrl(v, config.Country)

		b, err := httpGet(url, config.ApiKey)
		if err != nil {
			log.Printf("Failed to http get %s", url)
		}

		m := regexp.MustCompile(`urn:hbo:page:([A-Za-z0-9]+):type:`)
		content := m.FindAllSubmatch(b, 10000)
		for _, j := range content {
			urlList = append(urlList, resolveDetailUrl(string(j[1]), h.Config.Country, v))
		}
	}

	buildHboContent(urlList, config)

	return nil
}

func buildHboContent(urlList []string, config config.Config) {
	for i, url := range urlList {

		key := buildHboRedisKey(url)
		value, err := httpGet(config.ApiUrl+"/movie/"+key, config.ApiKey)
		if err != nil {
			fmt.Printf("Failed to http get %s - error: %s\n", config.ApiUrl+"/movie/"+key, err)
		}
		if value != nil {
			//fmt.Printf("%s --> found\n", v.Item.URL)
			continue
		}

		b, err := httpGet(url, config.ApiKey)
		if err != nil {
			fmt.Printf("Failed to http get %s - error: %s\n", url, err)
			continue
		}

		var detail *HboContentDetail
		if err := json.Unmarshal(extractJson(b), &detail); err != nil {
			log.Printf("Failed to Unmarshall %s", url)
			continue
		}

		var movie Movie
		movie.Title = detail.Name
		movie.Url = detail.URL
		movie.ContentRating = strings.Join(detail.ContentRating, ",")
		movie.Type = detail.Type
		movie.Description = detail.Description
		movie.Genre = strings.Join(detail.Genre, ",")
		movie.Image = detail.Image
		// movie.ReleaseDate = parseDate(detail.DateCreated)

		for _, dir := range detail.Actor {
			movie.Actors = append(movie.Actors, dir.Name)
		}

		// for _, dir := range detail.Director {
		// 	movie.Director = append(movie.Director, dir.Name)
		// }

		// for _, tr := range detail.Trailer {
		// 	var trailer Trailer
		// 	trailer.Url = tr.ContentURL
		// 	trailer.Name = tr.Name
		// 	trailer.Description = tr.Description
		// 	trailer.ThumbnailUrl = tr.ThumbnailURL
		// 	movie.Trailer = append(movie.Trailer, trailer)
		// }

		json_data, err := json.Marshal(movie)
		if err != nil {
			fmt.Printf("Failed to Marshall movie")
			continue
		}

		err = httpPost(config.ApiUrl+"/movie/"+key, bytes.NewBuffer(json_data), config.ApiKey)
		if err != nil {
			fmt.Printf("Failed to http post %s\n", key)
			continue
		}

		fmt.Printf("%d: %s --> %s\n", i, movie.Url, movie.Title)

		time.Sleep(3 * time.Second)
	}
}

func resolveDetailUrl(id string, country string, contentType string) string {

	if contentType == "movies" {
		contentType = "feature"
	}
	switch country {
	case "ES":
		{
			return fmt.Sprintf("https://www.hbomax.com/es/es/%s/urn:hbo:%s:%s", contentType, contentType, id)
		}
	case "US":
		{
			return fmt.Sprintf("https://www.hbomax.com/%s/urn:hbo:%s:%s", contentType, contentType, id)
		}
	}

	return ""
}

func resolveContentUrl(contentType string, country string) string {

	var culture string
	switch country {
	case "ES":
		{
			culture = "es-es"
		}
	case "US":
		{
			culture = "en-us"
		}
	}

	return fmt.Sprintf("https://comet-emea.api.hbo.com/express-content/urn:hbo:page:%s?device-code=desktop&product-code=hboMax&api-version=v9.0&country-code=%s&signed-in=false&profile-type=adult&brand=HBO%%20MAX&navigation-channels=HBO%%20MAX%%20SUBSCRIPTION%%7CHBO%%20MAX%%20FREE&upsell-channels=HBO%%20MAX%%20SUBSCRIPTION%%7CHBO%%20MAX%%20FREE&playback-channels=HBO%%20MAX%%20FREE&client-version=hadron_50.60&language=%s", contentType, country, culture)
}

func buildHboRedisKey(movieUrl string) string {

	var redisKey string
	redisKey = strings.Replace(movieUrl, "https://www.hbomax.com/es/es/", "hbo:es-es", 1)
	redisKey = strings.Replace(redisKey, "https://www.hbomax.com/", "hbo:en-us", 1)
	redisKey = strings.Replace(redisKey, "feature/urn:hbo:feature", "", 1)
	redisKey = strings.Replace(redisKey, "series/urn:hbo:series", "", 1)

	return redisKey
}

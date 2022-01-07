package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WhatsOnNetflixList struct {
	Title          string `json:"title"`
	Type           string `json:"type"`
	Titlereleased  string `json:"titlereleased"`
	ImageLandscape string `json:"image_landscape"`
	ImagePortrait  string `json:"image_portrait"`
	Rating         string `json:"rating"`
	Quality        string `json:"quality"`
	Actors         string `json:"actors"`
	Director       string `json:"director"`
	Category       string `json:"category"`
	Imdb           string `json:"imdb"`
	Runtime        string `json:"runtime"`
	Netflixid      string `json:"netflixid"`
	DateReleased   string `json:"date_released"`
	Description    string `json:"description"`
	Language       string `json:"language"`
}

func Whats_on_netflix() (*[]WhatsOnNetflixList, error) {

	const url = "https://www.whats-on-netflix.com/wp-content/plugins/whats-on-netflix/json/movie.json?_=1638813604000"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	var result []WhatsOnNetflixList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil

}

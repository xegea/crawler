package scraper

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

func httpGet(url string) ([]byte, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Content-Type": []string{"application/json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error Url: %s - Status Code: %d", req.URL, resp.StatusCode)
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

func unixTimestamp(date string) int64 {

	var year, day int
	var month time.Month

	fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
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

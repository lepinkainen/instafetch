package worker

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"io/ioutil"
	"net/http"
)

// GetPage returns a byte array of the given url's contents
// It also does a feeble attempt at impersonating some browser headers
func GetPage(url string) ([]byte, error) {

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error generating new request: %v", err)
	}

	// At least make a decent attempt at faking a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US")

	// Fetch the page
	res, err := client.Do(req)
	if err != nil {
		log.Errorln("HTTP Error: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	// Read the whole response to memory
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Error reading response from I/O: %v", err)
		return nil, err
	}

	if body[0] == '<' {
		return nil, errors.New("Received HTML content")
	}

	return body, nil
}

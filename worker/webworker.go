package worker

import (
	"errors"

	log "github.com/Sirupsen/logrus"

	"io/ioutil"
	"net/http"
)

func GetPage(url string) ([]byte, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error generating new request: %v", err)
	}

	// At least make a decent attempt at faking a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/604.3.5 (KHTML, like Gecko) Version/11.0.1 Safari/604.3.5")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-us")

	// Fetch the page
	res, err := client.Do(req)
	if err != nil {
		log.Errorln("HTTP Error", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	// Read the whole response to memory
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorln("Error reading response from I/O", err.Error())
		return nil, err
	}

	if body[0] == '<' {
		return body, errors.New("Received HTML content")
	}

	return body, nil
}

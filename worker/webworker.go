package worker

import (
	log "github.com/Sirupsen/logrus"

	"io/ioutil"
	"net/http"
)

func GetPage(url string) ([]byte, error) {

	// Fetch the page
	res, err := http.Get(url)
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

	return body, nil
}

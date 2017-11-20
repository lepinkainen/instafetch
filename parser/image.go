package parser

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/lepinkainen/instafetch/worker"
)

func getDirectImageURL(response MediaObject) string {
	return response.Graphql.ShortcodeMedia.DisplayURL
}

// GetImageURL returns image URL from shortcode
func getImageURL(shortcode string, urls chan<- string) {
	myLogger := log.WithField("module", "image")
	var url = fmt.Sprintf(mediaURL, shortcode)

	var response MediaObject

	data, err := worker.GetPage(url)
	if err != nil {
		myLogger.Errorln("Error fetching page", err.Error())
	}

	// unmarshal the JSON to the interface
	err = json.Unmarshal(data, &response)
	if err != nil {
		myLogger.Errorln("Error unmashaling JSON", err.Error())
		fmt.Println(string(data))
	}

	urls <- getDirectImageURL(response)

	myLogger.Debugf("Got image from shortcode %s", shortcode)
}

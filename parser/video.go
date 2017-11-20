package parser

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/lepinkainen/instafetch/worker"
)

var (
	mediaURL = "https://www.instagram.com/p/%s/?__a=1" // completed with shortcode
)

func getDirectVideoURL(response MediaObject) string {
	return response.Graphql.ShortcodeMedia.VideoURL
}

// GetVideoURL parses a video page and returns the direct video URL
func getVideoURL(shortcode string, urls chan<- string) {
	myLogger := log.WithField("module", "video")
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

	urls <- getDirectVideoURL(response)

	myLogger.Debugf("Got video from shortcode %s", shortcode)
}

package parser

import (
	"encoding/json"
	"fmt"
	"time"

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
func getVideoURL(baseItem DownloadItem, items chan<- DownloadItem) {
	myLogger := log.WithField("module", "video")
	var url = fmt.Sprintf(mediaURL, baseItem.Shortcode)

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

	item := DownloadItem(baseItem)
	item.URL = getDirectImageURL(response)
	item.Created = time.Unix(int64(response.TakenAtTimestamp), 0) // save created as go Time

	items <- item

	myLogger.Debugf("Got video from shortcode %s", baseItem.Shortcode)
}

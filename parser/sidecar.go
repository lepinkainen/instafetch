package parser

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lepinkainen/instafetch/worker"
)

// GetCarouselURLs parses a video page and returns the direct video URL
func getSidecarURLs(baseItem DownloadItem, items chan<- DownloadItem) {
	myLogger := log.WithField("module", "sidecar")
	var url = fmt.Sprintf(mediaURL, baseItem.Shortcode)

	var response mediaObject

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

	for _, image := range response.graphql.edgeSidecarToChildren.Edgess {
		item := DownloadItem(baseItem)
		item.URL = image.node.DisplayURL
		item.Created = time.Unix(int64(response.graphql.TakenAtTimestamp), 0) // save created as go Time

		items <- item
	}

	myLogger.Debugf("Got carousel images from shortcode %s", baseItem.Shortcode)
}

// Parses the main media stream
package parser

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/lepinkainen/instafetch/worker"
)

// GetCarouselURLs parses a video page and returns the direct video URL
func getSidecarURLs(shortcode string, urls chan<- string) {
	myLogger := log.WithField("module", "sidecar")
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

	for _, image := range response.Graphql.EdgeSidecarToChildren.Edgess {
		urls <- image.Node.DisplayURL
	}

	myLogger.Debugf("Got carousel images from shortcode %s", shortcode)
}

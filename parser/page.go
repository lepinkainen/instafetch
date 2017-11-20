package parser

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/lepinkainen/instafetch/worker"
)

var (
	// ID and EndCursor
	nextPageURL = "https://www.instagram.com/graphql/query/?query_id=17888483320059182&id=%s&first=100&after=%s"

	// QueryID: 17852405266163336
	// 17863787143139595
	// 17875800862117404
)

type Nextpage struct {
	Data   `json:"data"`
	Status string `json:"status"`
}

type Data struct {
	NextPageUser `json:"user"`
}

type NextPageUser struct {
	EdgeOwnerToTimelineMedia `json:"edge_owner_to_timeline_media"`
}

type EdgeOwnerToTimelineMedia struct {
	Count    int     `json:"count"`
	Edgess   []Edges `json:"edges"`
	PageInfo `json:"page_info"`
}

type Node struct {
	Typename           string `json:"__typename"`
	CommentsDisabled   bool   `json:"comments_disabled"`
	DisplayURL         string `json:"display_url"`
	EdgeMediaToCaption `json:"edge_media_to_caption"`
	ID                 string `json:"id"`
	IsVideo            bool   `json:"is_video"`
	Shortcode          string `json:"shortcode"`
}

type EdgeMediaToCaption struct {
	Edgess []Edges `json:"edges"`
}

// fetches all urls from a page and returns the cursor for the next page
func parseNextPage(id string, endCursor string, urls chan<- string) string {
	myLogger := log.WithField("module", "page")

	myLogger.Debug("-- Parsing next page")

	var url = fmt.Sprintf(nextPageURL, id, endCursor)

	// interface to hold the instagram JSON
	var response Nextpage

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

	for _, image := range response.Data.Edgess {
		switch shortcode := image.Typename; shortcode {
		case "GraphVideo":
			getVideoURL(image.Shortcode, urls)
		case "GraphSidecar":
			getSidecarURLs(image.Shortcode, urls)
		case "GraphImage":
			urls <- image.DisplayURL
		default:
			myLogger.Errorf("Unknown media type: '%v'", image.Typename)
		}
	}

	// return info about next page for looping through all pages
	if response.Data.PageInfo.HasNextPage {
		return response.Data.PageInfo.EndCursor
	}

	return ""
}

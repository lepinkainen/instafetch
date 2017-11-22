package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lepinkainen/instafetch/worker"
)

var (
	// ID and EndCursor
	nextPageURL = "https://www.instagram.com/graphql/query/?query_id=17888483320059182&id=%s&first=100&after=%s"
	//nextPageURL = "https://www.instagram.com/graphql/query/?query_id=17852405266163336&id=%s&first=100&after=%s"

	// QueryID: 17852405266163336
	// 17863787143139595
	// 17875800862117404
	// 17888483320059182
)

type Nextpage struct {
	Data    `json:"data"`
	Status  string `json:"status"`
	Message string `json:"message"`
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

type EdgeMediaToCaption struct {
	Edgess []Edges `json:"edges"`
}

func getPageImageItem(edge Edges) DownloadItem {
	return DownloadItem{
		URL: edge.DisplayURL,
	}
}

// fetches all urls from a page and returns the cursor for the next page
func parseNextPage(baseItem DownloadItem, id string, endCursor string, items chan<- DownloadItem) (string, error) {
	myLogger := log.WithField("module", "page")

	myLogger.Debug("-- Parsing next page")

	// generate url for the page
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

	if response.Status == "fail" {
		return "", errors.New(response.Message)
	}

	var wgSubWorkers sync.WaitGroup

	for _, image := range response.Data.Edgess {
		item := DownloadItem(baseItem)
		item.Shortcode = image.Shortcode

		switch shortcode := image.Typename; shortcode {
		case "GraphVideo":
			go func(item DownloadItem, items chan<- DownloadItem) {
				wgSubWorkers.Add(1)
				defer wgSubWorkers.Done()

				getVideoURL(item, items)
			}(item, items)
		case "GraphSidecar":
			go func(item DownloadItem, items chan<- DownloadItem) {
				wgSubWorkers.Add(1)
				defer wgSubWorkers.Done()

				getSidecarURLs(item, items)
			}(item, items)
		case "GraphImage":
			item.Created = time.Unix(int64(image.Node.TakenAtTimestamp), 0)
			item.URL = image.DisplayURL
			items <- item
		default:
			myLogger.Errorf("Unknown media type: '%v'", image.Typename)
		}
	}
	wgSubWorkers.Wait()

	// return info about next page for looping through all pages
	if response.HasNextPage {
		return response.EndCursor, nil
	}
	return "", nil
}

package parser

import (
	"fmt"

	"github.com/lepinkainen/instafetch/worker"
	"github.com/tidwall/gjson"
)

// Return a stream page as a gjson.Result
func getNextPage(id int64, cursor string) (gjson.Result, error) {
	var url = fmt.Sprintf(nextPageURL, id, cursor)

	bytes, err := worker.GetPage(url)
	if err != nil {
		fmt.Errorf("Error when fetching media page: %v", err)
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(bytes), nil
}

// Fetch a single media by shortcode
func getPageJSON(shortcode string) (gjson.Result, error) {
	var url = fmt.Sprintf(mediaURL, shortcode)

	bytes, err := worker.GetPage(url)
	if err != nil {
		fmt.Errorf("Error when fetching media page: %v", err)
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(bytes), nil
}

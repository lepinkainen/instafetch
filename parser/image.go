package parser

func getDirectImageURL(response mediaObject) string {
	return response.graphql.shortcodeMedia.DisplayURL
}

/*
// Not used at the moment

// GetImageURL returns image URL from shortcode
func getImageURL(shortcode string, items chan<- DownloadItem) {
	myLogger := log.WithField("module", "image")
	var url = fmt.Sprintf(mediaURL, shortcode)

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

	item := DownloadItem{
		URL: getDirectImageURL(response),
	}

	items <- item

	myLogger.Debugf("Got image from shortcode %s", shortcode)
}
*/

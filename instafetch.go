package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var userName = flag.String("username", "", "Username to back up")

// InstagramAPI holds all of the data returned by the Instagram media query
// generated with https://mholt.github.io/json-to-go/
type InstagramAPI struct {
	Items []struct {
		Code            string      `json:"code"`
		AltMediaURL     interface{} `json:"alt_media_url"`
		UserHasLiked    bool        `json:"user_has_liked"`
		ID              string      `json:"id"`
		CanViewComments bool        `json:"can_view_comments"`
		Location        interface{} `json:"location"`
		Caption         struct {
			Text string `json:"text"`
			From struct {
				FullName       string `json:"full_name"`
				Username       string `json:"username"`
				ProfilePicture string `json:"profile_picture"`
				ID             string `json:"id"`
			} `json:"from"`
			ID          string `json:"id"`
			CreatedTime string `json:"created_time"`
		} `json:"caption"`
		CreatedTime string `json:"created_time"`
		Images      struct {
			Thumbnail struct {
				Height int    `json:"height"`
				Width  int    `json:"width"`
				URL    string `json:"url"`
			} `json:"thumbnail"`
			LowResolution struct {
				Height int    `json:"height"`
				Width  int    `json:"width"`
				URL    string `json:"url"`
			} `json:"low_resolution"`
			StandardResolution struct {
				Height int    `json:"height"`
				Width  int    `json:"width"`
				URL    string `json:"url"`
			} `json:"standard_resolution"`
		} `json:"images"`
		Likes struct {
			Data []struct {
				FullName       string `json:"full_name"`
				Username       string `json:"username"`
				ProfilePicture string `json:"profile_picture"`
				ID             string `json:"id"`
			} `json:"data"`
			Count int `json:"count"`
		} `json:"likes"`
		CanDeleteComments bool `json:"can_delete_comments"`
		Comments          struct {
			Data  []interface{} `json:"data"`
			Count int           `json:"count"`
		} `json:"comments"`
		User struct {
			FullName       string `json:"full_name"`
			Username       string `json:"username"`
			ProfilePicture string `json:"profile_picture"`
			ID             string `json:"id"`
		} `json:"user"`
		Type   string `json:"type"`
		Link   string `json:"link"`
		Videos struct {
			StandardResolution struct {
				Height int    `json:"height"`
				Width  int    `json:"width"`
				URL    string `json:"url"`
			} `json:"standard_resolution"`
			LowBandwidth struct {
				Height int    `json:"height"`
				Width  int    `json:"width"`
				URL    string `json:"url"`
			} `json:"low_bandwidth"`
			LowResolution struct {
				Height int    `json:"height"`
				Width  int    `json:"width"`
				URL    string `json:"url"`
			} `json:"low_resolution"`
		} `json:"videos,omitempty"`
	} `json:"items"`
	MoreAvailable bool   `json:"more_available"`
	Status        string `json:"status"`
}

// DownloadItem contains all data needed to download a file
type DownloadItem struct {
	URL    string
	userID string
}

// parsePage returns images for the given user
func parsePage(userID string, maxID string) InstagramAPI {
	var url = fmt.Sprintf("https://www.instagram.com/%s/media/", userID)

	if maxID != "" {
		url = fmt.Sprintf("%s?max_id=%s", url, maxID)
	}

	res, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	// interface to hold the instagram json
	var response InstagramAPI

	// unmarshal the json to the interface
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err.Error())
	}

	return response
}

func getPages(userName string) []InstagramAPI {
	var responses []InstagramAPI

	// get the first page
	response := parsePage(userName, "")
	responses = append(responses, response)

	for {
		// Get last ID on this page
		lastID := response.Items[len(response.Items)-1].ID
		// fetch next page
		response = parsePage(userName, lastID)
		// store current page
		responses = append(responses, response)
		// no more pages, stop
		if !response.MoreAvailable {
			break
		}
	}

	log.Println("Pages parsed, count: ", len(responses))
	return responses
}

func parseMediaURLs(responses []InstagramAPI, userName string) []DownloadItem {

	var items []DownloadItem
	var url string

	for _, response := range responses {
		for _, item := range response.Items {
			switch item.Type {
			case "image":
				url = item.Images.StandardResolution.URL
				// Fix up the URL to return the full res image
				// TODO: imageURL = imageURL.replaceAll("\\?ig_cache_key.+$", "");
				url = strings.Replace(url, "s640x640/", "", -1)
				url = strings.Replace(url, "scontent.cdninstagram.com/hphotos-", "igcdn-photos-d-a.akamaihd.net/hphotos-ak-", -1)
			case "video":
				url = item.Videos.StandardResolution.URL
			default:
				fmt.Println("Unknown type: ", item.Type)
			}

			// TODO: Add the url to a queue for workers to download
			item := DownloadItem{}
			item.URL = url
			item.userID = userName
			items = append(items, item)
		}
	}
	return items
}

func downloadFile(item DownloadItem, outputFolder string) {
	var filename string

	os.MkdirAll(path.Join(outputFolder, item.userID), 0766)

	url, err := url.Parse(item.URL)
	if err != nil {
		panic(err.Error())
	}
	tokens := strings.Split(url.Path, "/")
	filename = tokens[len(tokens)-1]

	filename = path.Join(outputFolder, item.userID, filename)

	// Create output file
	// from: https://groups.google.com/d/msg/golang-nuts/Ayx-BMNdMFo/IVTRVqMECw8J
	out, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		log.Println("Already downloaded: ", filename)
		return
	}
	if err != nil {
		panic(err.Error())
	}
	defer out.Close()

	// download the file
	resp, err := http.Get(item.URL)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()

	// copy file to disk
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err.Error())
	}
	log.Println("Downloaded: ", filename)
}

func main() {
	flag.Parse()

	if *userName == "" {
		fmt.Println("username must be defined")
		return
	}

	responses := getPages(*userName)
	items := parseMediaURLs(responses, *userName)

	for _, item := range items {
		downloadFile(item, "output")
	}
}

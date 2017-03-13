package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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

// getImages returns images from the given user
func getImages(userID string, maxID string) InstagramAPI {
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

func main() {
	var url string
	userName := "lepinkainen"

	response := getImages(userName, "")

	if response.MoreAvailable {
		lastID := response.Items[len(response.Items)-1].ID
		fmt.Println("Last ID: ", lastID)
		//response = getImages(userName, lastID)
	}

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
		fmt.Println(url)
	}

}

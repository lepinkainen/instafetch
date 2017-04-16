// instafetch is a tool for quicky backing up an instagram account

package main

import (
	log "github.com/Sirupsen/logrus"

	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var userName = flag.String("username", "", "Username to back up")
var update = flag.Bool("update", false, "Update all existing downloads")
var latest = flag.Bool("latest", false, "Only fetch the first page of each target")
var debug = flag.Bool("debug", false, "Enable debug logging")

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
	URL     string
	userID  string
	created time.Time
}

// parsePage returns images for the given user
func parsePage(userID string, maxID string) InstagramAPI {
	var url = fmt.Sprintf("https://www.instagram.com/%s/media/", userID)

	// interface to hold the instagram JSON
	var response InstagramAPI

	if maxID != "" {
		url = fmt.Sprintf("%s?max_id=%s", url, maxID)
	}

	// Fetch the page
	res, err := http.Get(url)
	if err != nil {
		log.Errorln("HTTP Error", err.Error())
		return response
	}
	defer res.Body.Close()

	// Read the whole response to memory
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorln("Error reading response from I/O", err.Error())
		return response
	}

	// unmarshal the JSON to the interface
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Errorln("Error unmashaling JSON", err.Error())
		return response
	}

	return response
}

func getPages(userName string, c chan<- InstagramAPI) {
	var pageCount = 1

	// get the first page and send it forward immediately
	response := parsePage(userName, "")

	if len(response.Items) == 0 {
		log.Errorf("User page is private for %s", userName)
		return
	}

	c <- response

	if *latest {
		log.Printf("Only fetching latest page for %s", userName)
		return
	}

	for {
		// Get last ID on this page
		lastID := response.Items[len(response.Items)-1].ID
		// fetch next page
		response = parsePage(userName, lastID)

		pageCount = pageCount + 1

		log.Printf("Got page %d for %s", pageCount, userName)

		// An error during parsePage will return an empty interface
		if len(response.Items) == 0 {
			continue
		} else {
			// send found page response to channel
			c <- response
		}

		// no more pages, stop
		if !response.MoreAvailable {
			break
		}
	}

	log.Printf("Parsed %d pages for %s", pageCount, userName)
}

// Takes an InstagramAPI response and parses images it finds to DownloadItems
func parseMediaURLs(in <-chan InstagramAPI, out chan<- DownloadItem) {
	//defer close(out)

	var url string
	var userName string

	for response := range in {
		userName = response.Items[0].User.Username
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
				log.Warningln("Unknown type: ", item.Type)
			}

			// CreatedTime is an unix timestamp in string format
			i, err := strconv.ParseInt(item.CreatedTime, 10, 64)
			if err != nil {
				log.Panicln("Could not parse CreatedTime")
			}

			item := DownloadItem{}
			item.URL = url
			item.userID = userName
			item.created = time.Unix(i, 0) // save created as go Time
			out <- item
		}
	}
}

// download files received from the channel
func downloadFiles(c <-chan DownloadItem, outputFolder string) {
	// TODO: Worker-pattern for a fixed amount of simultaneous downloads
	for item := range c {
		log.Debugln("Downloading ", item)
		downloadFile(item, outputFolder)
	}
}

// download a single file defined by DownloadItem to outputFolder
func downloadFile(item DownloadItem, outputFolder string) {
	var filename string

	os.MkdirAll(path.Join(outputFolder, item.userID), 0766)

	url, err := url.Parse(item.URL)
	if err != nil {
		panic(err.Error())
	}
	tokens := strings.Split(url.Path, "/")
	// Grab the actual filename from the path
	filename = tokens[len(tokens)-1]
	// Prepend the date the image was added to instagram and username to the file for additional metadata
	created := item.created.UTC().Format("2006-01-02")
	filename = fmt.Sprintf("%s_%s_%s", created, item.userID, filename)
	filename = path.Join(outputFolder, item.userID, filename)

	// Create output file and check for its existence at the same time - no race conditions
	// from: https://groups.google.com/d/msg/golang-nuts/Ayx-BMNdMFo/IVTRVqMECw8J
	out, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		if *debug {
			log.Println("Already downloaded: ", filename)
		}
		return
	}
	if err != nil {
		log.Errorln("Error when opening file for saving: ", err.Error())
		return
	}
	defer out.Close()

	// download the file
	resp, err := http.Get(item.URL)
	if err != nil {
		log.Errorln("Error when downloading: ", err.Error())
		return
	}
	defer resp.Body.Close()

	// copy file to disk
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Panicln("Could not write file to disk", err.Error())
	}
	log.Printf("Downloaded: %s", filename)
}

func init() {
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true

	log.SetFormatter(formatter)
}

func main() {
	flag.Parse()

	// check for required variables
	if *userName == "" && *update == false {
		fmt.Println("Usage: ")
		flag.PrintDefaults()
		return
	}

	var accounts []string

	if *update {
		fmt.Println("Updating all existing sets:")
		// multiple accounts
		// loop through directories in output and assume each is an userID
		files, _ := ioutil.ReadDir("./output")
		for _, f := range files {
			if f.IsDir() {
				fmt.Println("  ", f.Name())
				accounts = append(accounts, f.Name())
			}
		}
		fmt.Println()
	} else {
		// Single account
		accounts = append(accounts, *userName)
	}

	// Buffered channels for pages and files to throttle a bit
	// Maximum number of page responses to buffer
	pages := make(chan InstagramAPI, 10)
	// Resolved and parsed actual DownloadItems
	files := make(chan DownloadItem, 200)

	var wgPages sync.WaitGroup

	// launch a goroutine for each account to fetch the pages
	// the buffered pages-channel will limit hammering on the server
	// combined with the single-threaded download operation
	for _, userName := range accounts {
		userName := userName // copy the variable to a new memory location
		go func() {
			// Use WaitGroup to track when all page goroutines are ready
			wgPages.Add(1)
			getPages(userName, pages)
			wgPages.Done()
		}()
	}

	// Wait for pages to be downloaded and close the channel after done
	// this will end the range loop and the related goroutines will finish
	go func() {
		wgPages.Wait()
		close(pages)
		log.Println("Page channel closed")
	}()

	var wgParse sync.WaitGroup
	wgParse.Add(1)
	// grab instagram api responses and produce a list to files to download
	go func() {
		parseMediaURLs(pages, files)
		wgParse.Done()
	}()

	go func() {
		wgParse.Wait()
		close(files)
		log.Println("File channel closed")
	}()

	downloadFiles(files, "output")
}

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

var (
	instagramURL = "https://www.instagram.com/%s/media/"
	// command line args
	userName = flag.String("username", "", "Username to back up")
	update   = flag.Bool("update", false, "Update all existing downloads")
	latest   = flag.Bool("latest", false, "Only fetch the first page of each target")
	debug    = flag.Bool("debug", false, "Enable debug logging")
	cron     = flag.Bool("cron", false, "Silent run for running from cron (most useful with --latest)")
)

// API Structs autogenerated with https://github.com/mohae/json2go/tree/master/cmd/json2go

// InstagramAPI holds all of the data returned by the Instagram media query
type InstagramAPI struct {
	Items         []Items `json:"items"`
	MoreAvailable bool    `json:"more_available"`
	Status        string  `json:"status"`
}

// Items is a struct for the media responses
type Items struct {
	Caption       `json:"caption"`
	Code          string `json:"code"`
	CreatedTime   string `json:"created_time"`
	ID            string `json:"id"`
	Images        `json:"images"`
	Videos        `json:"videos"`
	CarouselMedia `json:"carousel_media"`
	Link          string `json:"link"`
	Type          string `json:"type"`
	User          `json:"user"`
}

// Caption describes the media caption
type Caption struct {
	CreatedTime string `json:"created_time"`
	ID          string `json:"id"`
	Text        string `json:"text"`
}

// CarouselMedia is an image/video carousel of multiple images
type CarouselMedia []struct {
	Images `json:"images"`
	Videos `json:"videos"`
	Type   string `json:"type"`
}

// Images holds the image-type media info
type Images struct {
	StandardResolution `json:"standard_resolution"`
}

// Videos holds the video-type media info
type Videos struct {
	StandardResolution `json:"standard_resolution"`
}

// User describes the user who posted the media
type User struct {
	FullName       string `json:"full_name"`
	ID             string `json:"id"`
	ProfilePicture string `json:"profile_picture"`
	Username       string `json:"username"`
}

// StandardResolution holds the standard resolution media from Instagram
type StandardResolution struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

// DownloadItem contains all data needed to download a file
type DownloadItem struct {
	URL     string
	userID  string
	created time.Time
}

// parsePage returns images for the given user
func parsePage(userID string, maxID string) InstagramAPI {
	var url = fmt.Sprintf(instagramURL, userID)

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

// getPages retrieves all pages for given user and returns them
// to the given channel as InstagramAPI objects
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
		if !*cron {
			log.Printf("Only fetching latest page for %s", userName)
		}
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

		if !*cron {
			log.Printf("Got page %d for %s\n", pageCount, userName)
		}
		// no more pages, stop
		if !response.MoreAvailable {
			break
		}
	}

	if !*cron {
		log.Printf("Parsed %d pages for %s", pageCount, userName)
	}
}

// Takes an InstagramAPI response and parses images it finds to DownloadItems
func parseMediaURLs(in <-chan InstagramAPI, out chan<- DownloadItem) {
	var url string
	var userName string

	for response := range in {
		userName = response.Items[0].User.Username
		for _, item := range response.Items {

			// CreatedTime is an unix timestamp in string format
			i, err := strconv.ParseInt(item.CreatedTime, 10, 64)
			if err != nil {
				log.Panicln("Could not parse CreatedTime")
			}

			switch item.Type {
			case "image":
				url = item.Images.StandardResolution.URL
				// Fix up the URL to return the full res image
				url = strings.Replace(url, "s640x640/", "", -1)
			case "video":
				url = item.Videos.StandardResolution.URL
			case "carousel":
				log.Println("got carousel")
				for _, subItem := range item.CarouselMedia {
					switch subItem.Type {
					case "image":
						url = subItem.Images.StandardResolution.URL
						url = strings.Replace(url, "s640x640/", "", -1)
					case "video":
						url = subItem.Videos.StandardResolution.URL
					default:
						log.Warningf("Unknown subtype: %s for user %s ", item.Type, userName)
					}

					item := DownloadItem{}
					item.URL = url
					item.userID = userName
					item.created = time.Unix(i, 0) // save created as go Time
					out <- item
				}
				// the whole carousel has been sent for downloading, continue to the next main item
				continue
			default:
				log.Warningf("Unknown type: %s for user %s ", item.Type, userName)
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
		log.Debugln("Already downloaded: ", filename)
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

	// streams file to disk
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Panicln("Could not write file to disk", err.Error())
	}
	if !*cron {
		log.Printf("Downloaded: %s", filename)
	}
}

func init() {
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true

	log.SetFormatter(formatter)

	log.SetOutput(os.Stdout)

	if *cron {
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	flag.Parse()

	// check for required variables
	if *userName == "" && *update == false {
		fmt.Println("Usage: ")
		flag.PrintDefaults()
		return
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	cwd := path.Dir(ex)

	if *debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("Working directory %s", cwd)
	}

	outDir := path.Join(cwd, "output")

	var accounts []string

	if *update {
		if !*cron {
			fmt.Println("Updating all existing sets:")
		}
		// multiple accounts
		// loop through directories in output and assume each is an userID
		files, _ := ioutil.ReadDir(outDir)
		for _, f := range files {
			if f.IsDir() {
				if !*cron {
					fmt.Println("  ", f.Name())
				}
				accounts = append(accounts, f.Name())
			}
		}
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
		log.Debugln("Page channel closed")
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
		log.Debugln("File channel closed")
	}()

	downloadFiles(files, outDir)
}

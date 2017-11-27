// instafetch is a tool for quicky backing up an instagram account

// https://github.com/rarcega/instagram-scraper/commit/7ae2b3b2b80f7292a3a7bf036822ad6b23b7a9dd

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lepinkainen/instafetch/parser"

	"flag"
	"os"
)

var (
	// command line args
	userName            = flag.String("username", "", "Username to back up")
	update              = flag.Bool("update", false, "Update all existing downloads")
	latest              = flag.Bool("latest", false, "Only fetch the first page of each target")
	debug               = flag.Bool("debug", false, "Enable debug logging")
	cron                = flag.Bool("cron", false, "Silent run for running from cron (most useful with --latest)")
	rateLimitSleep      = 60
	downloadWorkerCount = 3
	pageWorkerCount     = 1 // anything above 1 here tends to trigger instagram's flood protection
)

// downloadFile gets a single file defined by DownloadItem to outputFolder
func downloadFile(item parser.DownloadItem, outputFolder string) error {
	var filename string

	// create download dir for the account
	os.MkdirAll(path.Join(outputFolder, item.UserID), 0766)

	url, err := url.Parse(item.URL)
	if err != nil {
		panic(err.Error())
	}
	tokens := strings.Split(url.Path, "/")
	// Grab the actual filename from the path
	filename = tokens[len(tokens)-1]
	// Prepend the date the image was added to instagram and username to the file for additional metadata
	// example output: 2017-11-05_alexandrabring_22860351_504365496598712_7456505757811343360_n.jpg
	created := item.Created.UTC().Format("2006-01-02")
	filename = fmt.Sprintf("%s_%s_%s", created, item.UserID, filename)
	filename = path.Join(outputFolder, item.UserID, filename)

	// Create output file and check for its existence at the same time - no race conditions
	// from: https://groups.google.com/d/msg/golang-nuts/Ayx-BMNdMFo/IVTRVqMECw8J
	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		log.Debugln("Already downloaded: ", filename)
		return nil
	}
	if err != nil {
		log.Errorln("Error when opening file for saving: ", err.Error())
		return err
	}
	defer out.Close()

	// download the file
	resp, err := http.Get(item.URL)
	if err != nil {
		log.Errorln("Error when downloading: ", err.Error())
		return err
	}
	defer resp.Body.Close()

	// streams file to disk
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("Could not write file to disk %s", err.Error())
		return err
	}
	if !*cron {
		log.Printf("Downloaded: %s", filename)
	}
	return nil
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

func downloadWorker(id int, outDir string, jobs <-chan parser.DownloadItem) {
	log.Debugf("DownloadWorker %d started", id)
	for job := range jobs {
		downloadFile(job, outDir)
	}

	log.Debugf("DownloadWorker %d stopped", id)
}

func parseWorker(id int, settings parser.Settings, jobs <-chan string, items chan<- parser.DownloadItem) {
	log.Debugf("ParseWorker %d started", id)
	for job := range jobs {
		log.Debugf("Parsing data for %s", job)
		err := parser.MediaURLs(job, settings, items)
		if err != nil {
			// rate limiting activated, no sense in attempting to continue
			if err.Error() == "rate limited" {
				log.Errorf("Rate limiting detected, pausing for %d seconds!", rateLimitSleep)
				time.Sleep(time.Second * time.Duration(rateLimitSleep))
			}
		}
	}
	log.Debugf("ParseWorker %d stopped", id)
}

func main() {
	flag.Parse()

	// check for required variables
	if *userName == "" && !*update {
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

	var wgDownloads sync.WaitGroup
	var wgParsing sync.WaitGroup

	// channel for urls, buffered
	users := make(chan string)
	items := make(chan parser.DownloadItem, 10)

	// start workers for downloads
	for w := 1; w <= downloadWorkerCount; w++ {
		wgDownloads.Add(1)
		go func(w int) {
			defer wgDownloads.Done()

			downloadWorker(w, outDir, items)
		}(w)
	}

	settings := parser.Settings{
		Silent:     *cron,
		LatestOnly: *latest,
	}

	// workers for page scraping
	for w := 1; w <= pageWorkerCount; w++ {
		wgParsing.Add(1)
		go func(w int) {
			defer wgParsing.Done()

			parseWorker(w, settings, users, items)
		}(w)
	}

	// Add work for parsers, which in turn will add work to the downloaders
	if *update {
		if !*cron {
			fmt.Println("Updating all existing sets")
		}

		// multiple accounts
		// loop through directories in output and assume each is an userID
		files, _ := ioutil.ReadDir(outDir)
		for _, f := range files {
			if f.IsDir() {
				users <- f.Name()
			}
		}
	} else {
		// Single account
		users <- *userName
	}
	log.Debug("Task queue full")
	// all users have been added, close the channel
	close(users)

	// Wait for pages to be downloaded and close the download worker input channel after done
	// this will end the range loop and the related goroutines will finish
	go func() {
		wgParsing.Wait()
		if !*cron {
			log.Info("All pages parsed, waiting for downloads to finish")
		}
		// All pages have been parsed, so we can close the job input channel
		close(items)
	}()

	// Wait for downloads to complete
	wgDownloads.Wait()
	if !*cron {
		log.Info("Downloads done")
	}
}

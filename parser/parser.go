package parser

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/lepinkainen/instafetch/worker"
)

// Settings defines the options for the downloaders
type Settings struct {
	LatestOnly bool
	Silent     bool
}

// User represents a single instagram user with their media nodes
type User struct {
	ID        int64
	Username  string
	FullName  string
	Followers int64
	Follows   int64
	Nodes     []Node
}

// Node is a single media node, image or video
type Node struct {
	Timestamp time.Time
	MediaType string
	URL       string
	IsVideo   bool
	Likes     int64
	Shortcode string
	ViewCount int64
}

var (
	userPageURL = "https://www.instagram.com/%s/?__a=1"                                                          // username
	mediaURL    = "https://www.instagram.com/p/%s/?__a=1"                                                        // completed with shortcode
	nextPageURL = "https://www.instagram.com/graphql/query/?query_id=17888483320059182&id=%d&first=100&after=%s" // userid + cursor
)

// Fetch the user's root page
func getRootPage(username string) (gjson.Result, error) {
	var url = fmt.Sprintf(userPageURL, username)

	bytes, err := worker.GetPage(url)
	if err != nil {
		log.Errorf("Error when fetching media page: %v", err)
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(bytes), nil
}

// ParseUser returns an User struct with all of the user's media
func ParseUser(username string, settings Settings) (User, error) {
	myLog := log.WithField("username", username)
	myLog.Infof("Starting parse")
	page, err := getRootPage(username)
	if err != nil {
		myLog.Errorf("Unable to get root page for user %s: %v", username, err)
		return User{}, err
	}

	userRoot := page.Get("graphql.user")

	user, hasNext, cursor, _ := parseFirstPage(userRoot)

	myLog.Infof("First page parsed")
	// User has more than one page of content
	if hasNext && !settings.LatestOnly {
		user, _ = parseStream(user, cursor)
	}
	myLog.Infof("Parsing done")
	return user, nil
}

// Recursively parse all of an users pages
func parseStream(user User, cursor string) (User, error) {
	myLog := log.WithField("username", user.Username)

	root, _ := getNextPage(user.ID, cursor)
	mediaRoot := root.Get("data.user.edge_owner_to_timeline_media")

	hasNext := mediaRoot.Get("page_info.has_next_page").Bool()
	newCursor := mediaRoot.Get("page_info.end_cursor").Str

	pageMedia := mediaRoot.Get("edges")

	nodes := parseEdges(pageMedia)

	user.Nodes = append(user.Nodes, nodes...)

	// recurse downwards if there are more pages
	if hasNext {
		myLog.Infof("Parsing next page (%4d nodes parsed)", len(user.Nodes))

		var err error
		user, err = parseStream(user, newCursor)
		if err != nil {
			myLog.Errorf("Error parsing stream: %v", err)
			return user, err
		}
	}
	return user, nil
}

// Parse the "edges" structure in the result json
func parseEdges(pageMedia gjson.Result) []Node {
	result := []Node{}

	for _, node := range pageMedia.Array() {

		typeName := node.Get("node.__typename").Str
		shortCode := node.Get("node.shortcode").Str

		switch typeName {
		// This needlessly fetches the whole separate json response for a single image
		case "GraphImage_full":
			res, _ := parseGraphImage(shortCode)
			result = append(result, res)
		case "GraphImage":
			res, _ := parseGraphImageNode(node)
			result = append(result, res)
		case "GraphVideo":
			res, _ := parseGraphVideo(shortCode)
			result = append(result, res)
		case "GraphSidecar":
			results, _ := parseGraphSidecar(shortCode)
			result = append(result, results...)
		default:
			log.Errorf("Uknown sidecar type '%v' for shortcode '%s'", typeName, shortCode)
		}
	}

	return result
}

func parseFirstPage(userRoot gjson.Result) (User, bool, string, error) {

	user := User{}
	user.Username = userRoot.Get("username").Str
	user.ID = userRoot.Get("id").Int()
	user.FullName = userRoot.Get("full_name").Str
	user.Followers = userRoot.Get("edge_followed_by.count").Int()
	user.Follows = userRoot.Get("edge_follow.count").Int()

	mediaRoot := userRoot.Get("edge_owner_to_timeline_media")

	pageMedia := mediaRoot.Get("edges")

	nodes := parseEdges(pageMedia)
	user.Nodes = append(user.Nodes, nodes...)

	hasNext := mediaRoot.Get("page_info.has_next_page").Bool()
	cursor := mediaRoot.Get("page_info.end_cursor").Str

	return user, hasNext, cursor, nil
}

// Fetch and Parse a GraphVideo node
func parseGraphVideo(shortCode string) (Node, error) {
	root, err := getPageJSON(shortCode)
	if err != nil {
		log.Errorf("Error fetching video page %s, %v", shortCode, err)
		return Node{}, err
	}
	node := root.Get("graphql.shortcode_media")

	// We actually need to fetch the subpage here
	result := Node{}
	result.URL = node.Get("video_url").Str
	result.ViewCount = node.Get("video_view_count").Int()

	result.MediaType = node.Get("__typename").Str
	result.Shortcode = node.Get("shortcode").Str
	result.IsVideo = node.Get("is_video").Bool()
	result.Timestamp = time.Unix(node.Get("taken_at_timestamp").Int(), 0)
	result.Likes = node.Get("edge_media_preview_like.count").Int()

	return result, nil
}

// Featch and Parse a GraphImage node
func parseGraphImage(shortCode string) (Node, error) {
	root, err := getPageJSON(shortCode)
	if err != nil {
		log.Errorf("Error fetching image page %s, %v", shortCode, err)
		return Node{}, err
	}
	node := root.Get("graphql.shortcode_media")

	result := Node{}
	result.URL = node.Get("display_url").Str

	result.MediaType = node.Get("__typename").Str
	result.Shortcode = node.Get("shortcode").Str
	result.IsVideo = node.Get("is_video").Bool()
	result.Timestamp = time.Unix(node.Get("taken_at_timestamp").Int(), 0)
	result.Likes = node.Get("edge_media_preview_like.count").Int()

	return result, nil
}

func parseGraphImageNode(node gjson.Result) (Node, error) {
	node = node.Get("node")

	result := Node{}
	result.URL = node.Get("display_url").Str

	result.MediaType = node.Get("__typename").Str
	result.Shortcode = node.Get("shortcode").Str
	result.IsVideo = node.Get("is_video").Bool()
	result.Timestamp = time.Unix(node.Get("taken_at_timestamp").Int(), 0)
	result.Likes = node.Get("edge_media_preview_like.count").Int()

	return result, nil
}

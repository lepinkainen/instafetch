package main

import (
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	"github.com/lepinkainen/instafetch/worker"
)

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

// Fetch the user's root page
func getRootPage(username string) (gjson.Result, error) {
	var url = fmt.Sprintf(userPageURL, username)

	bytes, err := worker.GetPage(url)
	if err != nil {
		fmt.Errorf("Error when fetching media page: %v", err)
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(bytes), nil
}

// ParseUser returns an User struct with all of the user's media
func ParseUser(username string) (User, error) {

	page, err := getRootPage(username)
	if err != nil {
		fmt.Errorf("Unable to get root page for user %s: %v", username, err)
		return User{}, err
	}

	userRoot := page.Get("graphql.user")

	user, hasNext, cursor, _ := parseFirstPage(userRoot)

	// User has more than one page of content
	if hasNext {
		user, _ = parseStream(user, cursor)
	}
	return user, nil
}

// Recursively parse all of an users pages
func parseStream(user User, cursor string) (User, error) {
	root, _ := getNextPage(user.ID, cursor)
	mediaRoot := root.Get("data.user.edge_owner_to_timeline_media")

	hasNext := mediaRoot.Get("page_info.has_next_page").Bool()
	newCursor := mediaRoot.Get("page_info.end_cursor").Str

	pageMedia := mediaRoot.Get("edges")

	nodes := parseEdges(pageMedia)

	user.Nodes = append(user.Nodes, nodes...)

	// recurse downwards if there are more pages
	if hasNext {
		var err error
		user, err = parseStream(user, newCursor)
		if err != nil {
			fmt.Errorf("ERror parsing stream: %v", err)
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
		case "GraphImage":
			res, _ := parseGraphImage(shortCode)
			result = append(result, res)
		case "GraphVideo":
			res, _ := parseGraphVideo(shortCode)
			result = append(result, res)
		case "GraphSidecar":
			results, _ := parseGraphSidecar(shortCode)
			result = append(result, results...)
		default:
			fmt.Println("unknown media type")
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
		fmt.Errorf("Error fetching video page %s, %v", shortCode, err)
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
		fmt.Errorf("Error fetching image page %s, %v", shortCode, err)
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

// Parse a GraphSidecar image
func parseSidecarImage(node gjson.Result) (Node, error) {

	result := Node{}
	result.URL = node.Get("display_url").Str
	result.MediaType = node.Get("__typename").Str
	result.Shortcode = node.Get("shortcode").Str
	result.IsVideo = node.Get("is_video").Bool()

	return result, nil
}

// Fetch and parse a GraphSidecar node
func parseGraphSidecar(shortCode string) ([]Node, error) {
	root, err := getPageJSON(shortCode)
	if err != nil {
		fmt.Errorf("Error fetching sidecar page %s, %v", shortCode, err)
		return []Node{}, err
	}
	node := root.Get("graphql.shortcode_media")

	result := []Node{}

	nodes := node.Get("edge_sidecar_to_children.edges")

	// Go through the nodes
	for _, node := range nodes.Array() {
		typeName := node.Get("node.__typename").Str
		shortCode := node.Get("node.shortcode").Str

		switch typeName {
		case "GraphImage":
			res, _ := parseSidecarImage(node.Get("node"))
			result = append(result, res)
		default:
			fmt.Errorf("Uknown sidecar type '%v' for shortcode '%s'", typeName, shortCode)
		}
	}

	return result, nil
}

func main() {
	user, _ := ParseUser("lepinkainen")

	fmt.Printf("Media count: %d", len(user.Nodes))
}

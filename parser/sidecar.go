package parser

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// Parse a GraphSidecar image
func parseSidecarImage(node gjson.Result) (Node, error) {

	result := Node{}
	result.URL = node.Get("display_url").Str
	result.MediaType = node.Get("__typename").Str
	result.Shortcode = node.Get("shortcode").Str
	result.IsVideo = node.Get("is_video").Bool()

	return result, nil
}

func parseSidecarVideo(node gjson.Result) (Node, error) {
	result := Node{}
	result.URL = node.Get("video_url").Str
	result.MediaType = node.Get("__typename").Str
	result.Shortcode = node.Get("shortcode").Str
	result.IsVideo = node.Get("is_video").Bool()
	result.ViewCount = node.Get("video_view_count").Int()

	return result, nil
}

// Fetch and parse a GraphSidecar node
func parseGraphSidecar(shortCode string) ([]Node, error) {
	root, err := getPageJSON(shortCode)
	if err != nil {
		log.Errorf("Error fetching sidecar page %s, %v", shortCode, err)
		return []Node{}, err
	}
	rootNode := root.Get("graphql.shortcode_media")

	result := []Node{}

	nodes := rootNode.Get("edge_sidecar_to_children.edges")

	timestamp := time.Unix(rootNode.Get("taken_at_timestamp").Int(), 0)

	// Go through the nodes
	for _, node := range nodes.Array() {
		typeName := node.Get("node.__typename").Str
		shortCode := node.Get("node.shortcode").Str

		switch typeName {
		case "GraphImage":
			res, _ := parseSidecarImage(node.Get("node"))
			// sidecars don't have separate timestamps, use the root's time
			res.Timestamp = timestamp
			result = append(result, res)
		case "GraphVideo":
			res, _ := parseSidecarVideo(node.Get("node"))
			// sidecars don't have separate timestamps, use the root's time
			res.Timestamp = timestamp
			result = append(result, res)
		default:
			log.Errorf("Uknown sidecar type '%v' for shortcode '%s'", typeName, shortCode)
		}
	}

	return result, nil
}

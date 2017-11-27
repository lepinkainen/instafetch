// Collection of common elements from Instagram JSON API

package parser

import "time"

// DownloadItem contains all data needed to download a file
type DownloadItem struct {
	URL       string
	UserID    string // Username
	ID        string // Numeric ID of user
	Shortcode string // Shortcode for the media
	Created   time.Time
}

// User defines the user who has posted stuff
type user struct {
	Biography              string      `json:"biography"`
	BlockedByViewer        bool        `json:"blocked_by_viewer"`
	ConnectedFbPage        interface{} `json:"connected_fb_page"`
	CountryBlock           bool        `json:"country_block"`
	ExternalURL            interface{} `json:"external_url"`
	ExternalURLLinkshimmed interface{} `json:"external_url_linkshimmed"`
	FollowedByViewer       bool        `json:"followed_by_viewer"`
	FollowsViewer          bool        `json:"follows_viewer"`
	FullName               string      `json:"full_name"`
	HasBlockedViewer       bool        `json:"has_blocked_viewer"`
	HasRequestedViewer     bool        `json:"has_requested_viewer"`
	ID                     string      `json:"id"`
	IsPrivate              bool        `json:"is_private"`
	IsVerified             bool        `json:"is_verified"`
	media                  `json:"media"`
	ProfilePicURL          string `json:"profile_pic_url"`
	ProfilePicURLHd        string `json:"profile_pic_url_hd"`
	RequestedByViewer      bool   `json:"requested_by_viewer"`
	Username               string `json:"username"`
}

// MediaObject defines the root element of shortcode replies
type mediaObject struct {
	graphql `json:"graphql"`
}

// Graphql response element, directly under MediaObject
type graphql struct {
	shortcodeMedia `json:"shortcode_media"`
}

// ShortcodeMedia - all media retrieved via direct shortcode link
type shortcodeMedia struct {
	Typename                   string             `json:"__typename"`
	CaptionIsEdited            bool               `json:"caption_is_edited"`
	CommentsDisabled           bool               `json:"comments_disabled"`
	DisplayURL                 string             `json:"display_url"`
	GatingInfo                 interface{}        `json:"gating_info"`
	ID                         string             `json:"id"`
	IsAd                       bool               `json:"is_ad"`
	IsVideo                    bool               `json:"is_video"`
	Location                   interface{}        `json:"location"`
	MediaPreview               string             `json:"media_preview"`
	Shortcode                  string             `json:"shortcode"`
	ShouldLogClientEvent       bool               `json:"should_log_client_event"`
	TakenAtTimestamp           int                `json:"taken_at_timestamp"`
	TrackingToken              string             `json:"tracking_token"`
	VideoURL                   string             `json:"video_url"`
	VideoViewCount             int                `json:"video_view_count"`
	ViewerHasLiked             bool               `json:"viewer_has_liked"`
	ViewerHasSaved             bool               `json:"viewer_has_saved"`
	ViewerHasSavedToCollection bool               `json:"viewer_has_saved_to_collection"`
	DisplayResourcess          []displayResources `json:"display_resources"`
	edgeSidecarToChildren      `json:"edge_sidecar_to_children"`
}

// PageInfo tells us if there is a new page after this one
type pageInfo struct {
	EndCursor   string `json:"end_cursor"`
	HasNextPage bool   `json:"has_next_page"`
}

type edges struct {
	node `json:"node"`
}

type displayResources struct {
	ConfigHeight int    `json:"config_height"`
	ConfigWidth  int    `json:"config_width"`
	Src          string `json:"src"`
}

type edgeSidecarToChildren struct {
	Edgess []edges `json:"edges"`
}

type media struct {
	Count    int     `json:"count"`
	Nodess   []nodes `json:"nodes"`
	pageInfo `json:"page_info"`
}

type nodes struct {
	Typename         string      `json:"__typename"`
	Caption          string      `json:"caption"`
	Code             string      `json:"code"`
	CommentsDisabled bool        `json:"comments_disabled"`
	Date             int         `json:"date"`
	DisplaySrc       string      `json:"display_src"`
	GatingInfo       interface{} `json:"gating_info"`
	ID               string      `json:"id"`
	IsVideo          bool        `json:"is_video"`
	MediaPreview     string      `json:"media_preview"`
	VideoViews       int         `json:"video_views"`
}

type node struct {
	Typename         string `json:"__typename"`
	CommentsDisabled bool   `json:"comments_disabled"`
	DisplayURL       string `json:"display_url"`
	//	edgeMediaToCaption `json:"edge_media_to_caption"`
	ID               string `json:"id"`
	IsVideo          bool   `json:"is_video"`
	Shortcode        string `json:"shortcode"`
	TakenAtTimestamp int    `json:"taken_at_timestamp"`
	ThumbnailSrc     string `json:"thumbnail_src"`
}

/*
type edgeMediaToCaption struct {
	Edgess []edges `json:"edges"`
}
*/
type edgeOwnerToTimelineMedia struct {
	Count    int     `json:"count"`
	Edgess   []edges `json:"edges"`
	pageInfo `json:"page_info"`
}

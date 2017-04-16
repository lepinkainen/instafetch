// https://www.reddit.com/user/theshrike/upvoted.json?feed=cf8e82696e1019150103ea28407f8464fc9d7baa&user=theshrike
package main

// RedditUpvoted holds the last upvoted posts by the user
type RedditUpvoted struct {
	Kind string `json:"kind"`
	Data struct {
		Modhash  string `json:"modhash"`
		Children []struct {
			Kind string `json:"kind"`
			Data struct {
				ContestMode bool        `json:"contest_mode"`
				BannedBy    interface{} `json:"banned_by"`
				MediaEmbed  struct {
				} `json:"media_embed"`
				Subreddit        string        `json:"subreddit"`
				SelftextHTML     interface{}   `json:"selftext_html"`
				Selftext         string        `json:"selftext"`
				Likes            bool          `json:"likes"`
				SuggestedSort    interface{}   `json:"suggested_sort"`
				UserReports      []interface{} `json:"user_reports"`
				SecureMedia      interface{}   `json:"secure_media"`
				LinkFlairText    interface{}   `json:"link_flair_text"`
				ID               string        `json:"id"`
				Gilded           int           `json:"gilded"`
				SecureMediaEmbed struct {
				} `json:"secure_media_embed"`
				Clicked               bool          `json:"clicked"`
				Score                 int           `json:"score"`
				ReportReasons         interface{}   `json:"report_reasons"`
				Author                string        `json:"author"`
				Saved                 bool          `json:"saved"`
				ModReports            []interface{} `json:"mod_reports"`
				Name                  string        `json:"name"`
				SubredditNamePrefixed string        `json:"subreddit_name_prefixed"`
				ApprovedBy            interface{}   `json:"approved_by"`
				Over18                bool          `json:"over_18"`
				Domain                string        `json:"domain"`
				Hidden                bool          `json:"hidden"`
				Preview               struct {
					Images []struct {
						Source struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"source"`
						Resolutions []struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"resolutions"`
						Variants struct {
							Obfuscated struct {
								Source struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"source"`
								Resolutions []struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"resolutions"`
							} `json:"obfuscated"`
							Nsfw struct {
								Source struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"source"`
								Resolutions []struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"resolutions"`
							} `json:"nsfw"`
						} `json:"variants"`
						ID string `json:"id"`
					} `json:"images"`
					Enabled bool `json:"enabled"`
				} `json:"preview"`
				Thumbnail           string      `json:"thumbnail"`
				SubredditID         string      `json:"subreddit_id"`
				Edited              bool        `json:"edited"`
				LinkFlairCSSClass   interface{} `json:"link_flair_css_class"`
				AuthorFlairCSSClass interface{} `json:"author_flair_css_class"`
				Downs               int         `json:"downs"`
				BrandSafe           bool        `json:"brand_safe"`
				Archived            bool        `json:"archived"`
				RemovalReason       interface{} `json:"removal_reason"`
				PostHint            string      `json:"post_hint"`
				IsSelf              bool        `json:"is_self"`
				HideScore           bool        `json:"hide_score"`
				Spoiler             bool        `json:"spoiler"`
				Permalink           string      `json:"permalink"`
				NumReports          interface{} `json:"num_reports"`
				Locked              bool        `json:"locked"`
				Stickied            bool        `json:"stickied"`
				Created             float64     `json:"created"`
				URL                 string      `json:"url"`
				AuthorFlairText     interface{} `json:"author_flair_text"`
				Quarantine          bool        `json:"quarantine"`
				Title               string      `json:"title"`
				CreatedUtc          float64     `json:"created_utc"`
				Distinguished       interface{} `json:"distinguished"`
				Media               interface{} `json:"media"`
				NumComments         int         `json:"num_comments"`
				Visited             bool        `json:"visited"`
				SubredditType       string      `json:"subreddit_type"`
				Ups                 int         `json:"ups"`
			} `json:"data"`
		} `json:"children"`
		After  string      `json:"after"`
		Before interface{} `json:"before"`
	} `json:"data"`
}

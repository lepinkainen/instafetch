package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	mediaJSON = `{
   "items":[
      {
         "id":"1442206613258829026_22792833",
         "code":"BQDv0YMAvzi",
         "user":{
            "id":"22792833",
            "full_name":"Test User",
            "profile_picture":"https://scontent-ams3-1.cdninstagram.com/t51.2885-19/s150x150/12716678_1180466221963874_1859835217_a.jpg",
            "username":"instagram"
         },
         "images":{
            "thumbnail":{
               "width":150,
               "height":150,
               "url":"https://scontent-ams3-1.cdninstagram.com/t51.2885-15/s150x150/e35/16230488_110094772841683_7263292011141136384_n.jpg"
            },
            "low_resolution":{
               "width":320,
               "height":320,
               "url":"https://scontent-ams3-1.cdninstagram.com/t51.2885-15/s320x320/e35/16230488_110094772841683_7263292011141136384_n.jpg"
            },
            "standard_resolution":{
               "width":640,
               "height":640,
               "url":"https://scontent-ams3-1.cdninstagram.com/t51.2885-15/s640x640/sh0.08/e35/16230488_110094772841683_7263292011141136384_n.jpg"
            }
         },
         "created_time":"1486144447",
         "caption":{
            "id":"17860387969107590",
            "text":"Corn-fed chicken with garlic sauce and cous-cous #food",
            "created_time":"1486144447",
            "from":{
               "id":"22792833",
               "full_name":"Riku Lindblad",
               "profile_picture":"https://scontent-ams3-1.cdninstagram.com/t51.2885-19/s150x150/12716678_1180466221963874_1859835217_a.jpg",
               "username":"lepinkainen"
            }
         },
         "user_has_liked":false,
         "likes":{
            "data":[
               {
                  "id":"1957659109",
                  "full_name":"\u041b\u0435\u0441\u0443\u043d\u044c\u043a\u0430 \u041f\u043e\u0442\u0430\u0440",
                  "profile_picture":"https://scontent-ams3-1.cdninstagram.com/t51.2885-19/s150x150/11264259_157036187965920_938536761_a.jpg",
                  "username":"olesyapotar"
               }
            ],
            "count":133
         },
         "comments":{
            "data":[
               {
                  "id":"17871360793005105",
                  "text":"Great picture",
                  "created_time":"1486402892",
                  "from":{
                     "id":"3303766116",
                     "full_name":"",
                     "profile_picture":"https://scontent-ams3-1.cdninstagram.com/t51.2885-19/s150x150/13395172_1029373177132334_2000321521_a.jpg",
                     "username":"mo_aralius"
                  }
               },
               {
                  "id":"17877831475002493",
                  "text":"\ud83d\udc95",
                  "created_time":"1491546899",
                  "from":{
                     "id":"2324583565",
                     "full_name":"Fay",
                     "profile_picture":"https://scontent-ams3-1.cdninstagram.com/t51.2885-19/s150x150/17077187_1211916322257040_3918842029043351552_a.jpg",
                     "username":"fay_a_smith"
                  }
               }
            ],
            "count":2
         },
         "can_view_comments":true,
         "can_delete_comments":true,
         "type":"image",
         "link":"https://www.instagram.com/p/BQDv0YMAvzi/",
         "location":{
            "name":"Vanajanlinna"
         },
         "alt_media_url":null
      }
   ],
   "more_available":false,
   "status":"ok"
}`
)

func TestJSONParsing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mediaJSON)
	}))
	defer ts.Close()

	// generate fake URL for parsePage
	instagramURL = ts.URL + "/%s"
	result := parsePage("instagram", "")

	if result.MoreAvailable != false {
		t.Error("Invalid MoreAvailable response")
	}

	itemCount := len(result.Items)
	if itemCount != 1 {
		t.Errorf("Found too many items: %d", itemCount)
	}

	if result.Items[0].User.Username != "instagram" {
		t.Errorf("Wrong username in response")
	}
}

func TestGetPages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mediaJSON)
	}))
	defer ts.Close()

	instagramURL = ts.URL + "/%s"
	c := make(chan InstagramAPI)
	defer close(c)

	go getPages("lepinkainen", c)
	result := <-c

	if result.MoreAvailable != false {
		t.Error("Invalid MoreAvailable response")
	}

	itemCount := len(result.Items)
	if itemCount != 1 {
		t.Errorf("Found too many items: %d", itemCount)
	}

	if result.Items[0].User.Username != "instagram" {
		t.Errorf("Wrong username in response")
	}
}

func createTestItem(url, createdTime, mediaType string) []Items {
	// Slice to hold the media items
	items := []Items{}
	// a single media item
	item := Items{}
	item.Images.URL = url
	item.CreatedTime = createdTime
	item.Type = mediaType

	// append the test item to the slice
	items = append(items, item)

	return items
}

func TestParseMediaURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mediaJSON)
	}))
	defer ts.Close()

	in := make(chan InstagramAPI)
	out := make(chan DownloadItem)

	go parseMediaURLs(in, out)
	apiresponse := InstagramAPI{}

	// A valid test item
	apiresponse.Items = createTestItem("https://scontent-ams3-1.cdninstagram.com/t51.2885-15/s640x640/sh0.08/e35/16230488_110094772841683_7263292011141136384_n.jpg",
		"1486144447",
		"image")
	apiresponse.MoreAvailable = false

	in <- apiresponse
	resultItem := <-out

	if resultItem.URL != "https://scontent-ams3-1.cdninstagram.com/t51.2885-15/sh0.08/e35/16230488_110094772841683_7263292011141136384_n.jpg" {
		t.Errorf("Large image URL detection failed")
	}

	// TODO: check	resultItem.created

	/*
		// TODO: Invalid test item
		// How to handle errors thrown by the method
		apiresponse.Items = createTestItem("https://scontent-ams3-1.cdninstagram.com/t51.2885-15/s640x640/sh0.08/e35/16230488_110094772841683_7263292011141136384_n.jpg",
			"1486144447",
			"")

		in <- apiresponse
		resultItem = <-out
	*/

}

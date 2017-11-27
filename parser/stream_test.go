package parser

import "testing"

func Test_getNextPageInfo(t *testing.T) {
	type args struct {
		response instagramAPI
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"Has next page", args{instagramAPI{user: user{ID: "12345", media: media{pageInfo: pageInfo{
			HasNextPage: true,
			EndCursor:   "thisistheend",
		}}}}},
			"12345",
			"thisistheend"},
		{"Last page", args{instagramAPI{user: user{ID: "12345", media: media{pageInfo: pageInfo{
			HasNextPage: false,
			EndCursor:   "thisistheend",
		}}}}},
			"",
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getNextPageInfo(tt.args.response)
			if got != tt.want {
				t.Errorf("getNextPageInfo() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getNextPageInfo() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

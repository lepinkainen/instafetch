package parser

import "testing"

func Test_getDirectImageURL(t *testing.T) {
	type args struct {
		response mediaObject
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Basic test",
			args{response: mediaObject{graphql: graphql{shortcodeMedia{DisplayURL: "http://httpbin.org/"}}}},
			"http://httpbin.org/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDirectImageURL(tt.args.response); got != tt.want {
				t.Errorf("getDirectImageURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

package main

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	//"io/ioutil"
	"code.google.com/p/go.net/html"
	"io"
	//	"log"
	"net/http"
	"strings"
)

func init() {
	http.HandleFunc("/", handler)
}

// func handler(w http.ResponseWriter, r *http.Request){
// 	fmt.Fprint(w, "Hello, world!")
// }

// var SEARCH_URL = "http://images.google.co.jp/images?q=C%23&hl=ja"
// var SEARCH_URL = "http://images.google.co.jp/images?q=golang&hl=ja"
var SEARCH_URL = "http://www.google.co.jp/search?q=golang&hl=ja"

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.Get(SEARCH_URL)
	defer resp.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/html charset=UTF-8")
	fmt.Fprintf(w, "<!DOCTYPE HTML><html lang=\"ja\"><head><meta charset=\"UTF-8\"></head><body>")

	ParseGoogleSearch(w, resp.Body)
	fmt.Fprintln(w, "</body></html>")

	// fmt.Fprintf(w, "HTTP GET returned status %v", resp)
}

func ParseGoogleSearch(w http.ResponseWriter, r io.Reader) {
	var (
		classrflag bool = false
		key, val   []byte
		url        string
		text       string
	)

	t := html.NewTokenizer(r)

	for {
		tokenType := t.Next()

		switch tokenType {
		case html.ErrorToken:
			return
		case html.StartTagToken, html.EndTagToken: //<a href="http://~">,</a>, <h3 class="r">
			tagname, _ := t.TagName() // a
			if string(tagname) == "h3" {
				key, val, _ = t.TagAttr() // href, http://~, class, r
				if string(key) == "class" {
					if string(val) == "r" {
						classrflag = true
					}
				}
			}

			if classrflag {
				key, val, _ = t.TagAttr()

				if string(tagname) == "a" {
					if string(key) == "href" {
						aval := strings.Split(string(val), "&")
						url = aval[0][7:]
					}
				}

				if tokenType == html.EndTagToken {
					if string(tagname) == "a" {
						fmt.Fprintf(w, "<a href=\"%s\">%v</a>", url, text)
						fmt.Fprintln(w, "<br>")
						text = ""
						classrflag = false
					}
				}
			}

		case html.TextToken:
			if classrflag {
				text += string(t.Text())
			}
		case html.SelfClosingTagToken:
		}
	}

}

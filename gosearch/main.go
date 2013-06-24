package main

import (
	"appengine"
	"appengine/urlfetch"
	// "fmt"
	//"io/ioutil"
	"code.google.com/p/go.net/html"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

type Result struct {
	Text string
	Url  string
}

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

	w.Header().Add("Content-type", "text/html charset=utf-8")
	results := ParseGoogleSearch(w, resp.Body)

	t := template.Must(template.ParseFiles("gosearch/tmpl/main.tmpl"))
	err = t.Execute(w, results)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	// fmt.Fprintf(w, "HTTP GET returned status %v", resp)
}

//Google Web検索
func ParseGoogleSearch(w http.ResponseWriter, r io.Reader) []Result {
	var (
		classrflag bool = false
		key, val   []byte
		result     Result
		results    []Result
	)

	t := html.NewTokenizer(r)

	for {
		tokenType := t.Next()

		switch tokenType {
		case html.ErrorToken:
			//fmt.Fprintln(w, "return")
			return results
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
						result.Url = aval[0][7:]
					}
				}

				if tokenType == html.EndTagToken {
					if string(tagname) == "a" {
						tmp := Result{result.Text, result.Url}
						if len(results) != 0 {
							results = append(results, tmp)
						} else {
							results = []Result{tmp}
						}
						result.Text = ""
						classrflag = false
					}
				}
			}

		case html.TextToken:
			if classrflag {
				result.Text += string(t.Text())
			}
		case html.SelfClosingTagToken:
		}
	}

	return results

}


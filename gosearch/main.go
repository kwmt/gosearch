package main

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	//"io/ioutil"
	"code.google.com/p/go.net/html"
	"code.google.com/p/mahonia"
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
	http.HandleFunc("/", index)
	http.HandleFunc("/search", search)
}

var SEARCH_URL = "http://www.google.co.jp/search?hl=ja&q="

func index(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("gosearch/tmpl/main.tmpl"))
	err := t.Execute(w, nil)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
}
func search(w http.ResponseWriter, r *http.Request) {
	search_string := r.FormValue("search_string")

	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.Get(SEARCH_URL + search_string)
	defer resp.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "text/html; charset=utf-8")
	results := ParseGoogleSearch(w, resp.Body)

	t := template.Must(template.ParseFiles("gosearch/tmpl/main.tmpl"))
	err = t.Execute(w, results)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
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
						rt := mahonia.NewDecoder("Shift_JIS").ConvertString(result.Text)
						tmp := Result{rt, result.Url}
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
				//result.Text = append(result.Text, t.Text()...)
			}
		case html.SelfClosingTagToken:
		}
	}

	return results

}

//Google 画像検索(未使用)
//  http://godoc.org/code.google.com/p/go.net/html
// にのっているサンプルにParse部分を追加
func ParseGoogleImageSearch(w http.ResponseWriter, r io.Reader) {
	doc, err := html.Parse(r)
	if err != nil {
		log.Fatal(err)
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					str := a.Val
					if strings.Contains(str, "imgurl") {
						strs := strings.Split(str, "&")
						imageurl := strings.Split(strs[0], "=")
						img := imageurl[1]
						fmt.Fprintf(w, "<html><body><ul><li><a href=%v><img src=%v></a></li></ul></body></html>", img, img)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

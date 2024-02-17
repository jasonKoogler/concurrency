package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type crawler struct {
	wg     *sync.WaitGroup
	respch chan *Response
	// resultch chan *ResultMap
}

func NewCrawler(respchBuf int) *crawler {
	return &crawler{
		wg:     new(sync.WaitGroup),
		respch: make(chan *Response, respchBuf),
		// resultch: make(chan *ResultMap, resultchBuf),
	}
}

// Response is the response of a website visit
// to be sent through the channnel for further processing.
type Response struct {
	Url  string
	Body string
	Err  error
}

// NewResponse is a helper function for creating a new
// Response object.
func NewResponse(url string, body []byte, err error) *Response {
	return &Response{
		Url:  url,
		Body: string(body),
		Err:  err,
	}
}

// ResultMap is an alias for a basic map[string]string, it will be
// populated be the GrabBody func, and is meant to contain the pertinent
// data from the crawling operation
// ie [url][body]
type ResultMap map[string]string

// GrabBody takes a slice of urls and spawns goroutines
// to fetch the body from the response, assigns the result to
// the ResultMap and finally sends it through the result channel
func (c *crawler) GrabBody(urls []string) *ResultMap {
	for _, u := range urls {
		c.wg.Add(1)
		go fetch(u, c.respch, c.wg)
	}
	c.wg.Wait()
	close(c.respch)

	var result = make(ResultMap)

	for msg := range c.respch {
		if msg.Err != nil {
			log.Print(msg.Err)
			continue
		}
		result[msg.Url] = msg.Body
	}

	return &result
}

// Fetch takes a url and a response channel and a WaitGroup, makes an
// HTTP GET request and collates the resp.Body and error,
// if there is one and sends it into the channel
func fetch(url string, respch chan *Response, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		respch <- NewResponse(url, nil, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Status code other than 200:", resp.StatusCode)
		return
	}

	b, err := io.ReadAll(resp.Body)
	respch <- NewResponse(url, b, err)
}

func main() {
	var urls = []string{"https://books.toscrape.com", "https://quotes.toscrape.com"}

	c := NewCrawler(len(urls))

	// r := <-c.GrabBody(urls)
	r := c.GrabBody(urls)

	fmt.Println(r)
}

package proxy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ProxyResult struct {
	Ok    bool
	Key   string
	Error error
	Data  []byte
}

// NewHLSProxy creates a new server
func NewHLSProxy() *HLSProxy {
	return &HLSProxy{logger: log.New(os.Stdout, "", 0)}
}

// HLSProxy handles file cache
type HLSProxy struct {
	cachePath string
	addr      string
	cache     *Cache
	client    *http.Client
	logger    *log.Logger
}

// Setup setup cache proxy
func (s *HLSProxy) Setup(addr, cachePath string) {

	Info.Println("Init cache")
	cache, err := CreateCache(cachePath)
	if err != nil {
		Error.Fatalf("Could not init cache: '%s'", err.Error())
		return
	}
	s.client = &http.Client{
		Timeout: time.Second * 20,
	}
	s.addr = addr
	s.cache = cache
	s.logger.Println("Setup Cache")
}

// RewriteHLS rewrite hls segment urls to proxy urls
func (s *HLSProxy) RewriteHLS(fullURL string) *ProxyResult {
	response, err := s.client.Get(fullURL)
	if err != nil {
		return &ProxyResult{false, "", err, nil}
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return &ProxyResult{false, "", err, nil}
	}

	// parse and replace urls
	body, _ = ReplaceHLSUrls(body, fmt.Sprintf("%s/cache?file=", s.addr))
	return &ProxyResult{true, "", nil, body}
}

// Clear clear proxy cache
func (s *HLSProxy) Clear() *ProxyResult {
	err := s.cache.clear()
	return &ProxyResult{
		Error: err,
	}
}

// Has check if cache has item or cache item
func (s *HLSProxy) Has(fullURL string) *ProxyResult {
	if key, ok := s.cache.has(fullURL); ok {
		// TODO: validate cache value is correct
		return &ProxyResult{true, key, nil, nil}
	}
	// Debug.Printf("cache does not contain %s", fullURL)
	response, err := s.client.Get(fullURL)
	if err != nil {
		return &ProxyResult{false, "", err, nil}
	}
	if response.StatusCode != 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return &ProxyResult{false, "", err, nil}
		}
		response.Body.Close()
		log.Println("retrieving cache source failed", fullURL, body)
		return &ProxyResult{false, "", errors.New("unsuccessful response"), nil}
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return &ProxyResult{false, "", err, nil}
	}

	if strings.Contains(fullURL, ".m3u8") {
		// parse and replace urls
		body, _ = ReplaceHLSUrls(body, fmt.Sprintf("%s/cache?r=1&file=", s.addr))
	}

	// TODO: write check file as well
	// i.e write content length to fullURL+size
	// use the value to check if cash item is valid by comparing size of item to size in file
	key, err := s.cache.put(fullURL, body)

	// Do not fail. Even if the put failed, the end user would be sad if he
	// gets an error, even if the proxy alone works.
	if err != nil {
		Error.Printf("Could not write into cache: %s", err)
		return &ProxyResult{false, "", err, nil}
	}

	return &ProxyResult{true, key, nil, nil}
}

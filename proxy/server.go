package proxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// DefaultServer is a global proxy server
var DefaultServer = NewServer()

// NewServer creates a new server
func NewServer() *Server {
	return &Server{logger: log.New(os.Stdout, "", 0)}
}

// Server defines a proxy cache server
type Server struct {
	addr      string
	cachePath string
	logger    *log.Logger
	timeout   time.Duration
	server    *http.Server
	cache     *Cache
	client    *http.Client
}

// Setup setup the server with addr and cache path
func (s *Server) Setup(addr, cachePath string) {

	Info.Println("Init cache")
	cache, err := CreateCache(cachePath)
	if err != nil {
		Error.Fatalf("Could not init cache: '%s'", err.Error())
		return
	}
	s.cache = cache

	s.client = &http.Client{
		Timeout: time.Second * 30,
	}
	if addr == "" {
		addr = ":2017"
	}
	s.addr = addr

	s.server = &http.Server{
		Addr:         addr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		Handler:      http.HandlerFunc(s.handleGet),
	}
}

// Start the server
func (s *Server) Start() {
	go func() {
		s.logger.Printf("Listening on http://%s\n", s.addr)

		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Fatal(err)
		}
	}()
}

// Shutdown the server
func (s *Server) Shutdown() {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	s.logger.Printf("\nShutdown with timeout: %s\n", s.timeout)

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Printf("Error: %v\n", err)
	} else {
		s.logger.Println("Server stopped")
	}
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	fullURL := strings.TrimLeft(r.URL.Path+"?"+r.URL.RawQuery, "/")

	Info.Printf("Requested '%s'\n", fullURL)

	// Only pass request to target host when cache does not has an entry for the
	// given URL.
	if s.cache.has(fullURL) {
		content, err := s.cache.get(fullURL)

		if err != nil {
			s.handleError(err, w)
		} else {
			w.Write(content)
		}
	} else {
		Debug.Printf("cache does not contain %s", fullURL)
		response, err := s.client.Get(fullURL)
		if err != nil {
			s.handleError(err, w)
			return
		}

		body, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			s.handleError(err, w)
			return
		}

		if strings.Contains(fullURL, ".m3u8") {
			// parse and replace urls
			body, _ = replaceHLSUrls(body, fmt.Sprintf("http://%s/", s.addr))
		}

		err = s.cache.put(fullURL, body)

		// Do not fail. Even if the put failed, the end user would be sad if he
		// gets an error, even if the proxy alone works.
		if err != nil {
			Error.Printf("Could not write into cache: %s", err)
		}

		w.Write(body)
	}
}

func (s *Server) handleError(err error, w http.ResponseWriter) {
	Error.Println(err.Error())
	w.WriteHeader(500)
	fmt.Fprintf(w, err.Error())
}

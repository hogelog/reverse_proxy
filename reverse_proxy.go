package main

import (
	"log"
	"bytes"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type CloserBuffer struct {
	*bytes.Buffer
}
func (buffer *CloserBuffer) Close() (err error) {
	return err
}

type CacheResponse struct {
	Response *http.Response
	Body []byte
}
func (cache *CacheResponse) response() (res *http.Response, err error) {
	clone := *cache.Response
	clone.Body = &CloserBuffer{bytes.NewBuffer(cache.Body)}
	return &clone, err
}

type Proxy struct {
	DefaultTransport http.RoundTripper
	cache map[string]*CacheResponse
}
func (proxy *Proxy) RoundTrip(req *http.Request) (res *http.Response, err error) {
	if proxy.cache == nil {
		proxy.cache = make(map[string]*CacheResponse)
	}
	log.Printf("Request: %v", req.URL)
	key := req.Method + req.URL.Path + req.URL.RawQuery

	cache := proxy.cache[key]
	if cache == nil {
		log.Printf("Cache: MISS")
		connRes, connErr := proxy.DefaultTransport.RoundTrip(req)
		if connErr != nil {
			return connRes, connErr
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(connRes.Body)
		cache = &CacheResponse{Response: connRes, Body: buf.Bytes()}
		proxy.cache[key] = cache
	} else {
		log.Printf("Cache: HIT")
	}
	return cache.response()
}

func connectReverseProxy(connectUrl, listenAddr string) {
	log.Printf("Listen: %v, Connect: %v", listenAddr, connectUrl)
	target, _ := url.Parse(connectUrl)

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}
	transport := &Proxy{DefaultTransport: http.DefaultTransport}
	reverseProxy := &httputil.ReverseProxy{Director: director, Transport: transport}

	s := &http.Server{
		Addr:    listenAddr,
		Handler: reverseProxy,
	}
	log.Fatal(s.ListenAndServe())
}

func main() {
  connectReverseProxy("http://localhost:3000", ":3002")
}

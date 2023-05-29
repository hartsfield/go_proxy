package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	insecure := newServerConf(httpPort, http.HandlerFunc(forwardHTTP))
	secure := newServerConf(tlsPort, http.HandlerFunc(forwardTLS))

	ctx, cancel := context.WithCancel(context.Background())
	globalHalt = cancel

	go startHTTPServer(insecure)
	go startTLSServer(secure)

	<-ctx.Done()
}

// makeProxy takes var #SERVICE *service{} and creates a *http.ReverseProxy
// using the properties of #SERVICE
func makeProxy(s *service) *service {
	u, err := url.Parse("http://localhost:" + s.Port + "/")
	if err != nil {
		log.Println(err)
	}
	s.ReverseProxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", u.Host)
			req.URL.Host = u.Host
			req.URL.Scheme = "http"
		},
		FlushInterval: -1,
	}
	return s
}

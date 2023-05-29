package main

import (
	"log"
	"net/http"
	"time"
)

// newServerConf returns a new server configuration, and is used when
// instantiating the servers that intercept HTTP/S traffic.
func newServerConf(port string, hf http.HandlerFunc) *http.Server {
	return &http.Server{
		Addr:              ":" + port,
		Handler:           hf,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Second,
	}
}

// startHTTPServer is used to start the HTTP server
func startHTTPServer(s *http.Server) {
	err := s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
	// if the server encounters an error, this function will be called to
	// halt the server.
	globalHalt()
}

// startTLSServer is used to start the TLS server
func startTLSServer(s *http.Server) {
	err := s.ListenAndServeTLS(certs.Fullchain, certs.Privkey)
	if err != nil {
		log.Println(err)
	}
	// if the server encounters an error, this function will be called to
	// halt the server.
	globalHalt()
}

// forwardTLS is the handler used for requests on the secure port (TLS/HTTPS).
// forwardTLS will check if a host exists and has TLS enabled, if both are
// true, it serves the website, if the host exists, but doesn't have TLS
// enabled, it forwarss it to the HTTP server, otherwise it sends the client to
// the 'not found' page.
func forwardTLS(w http.ResponseWriter, r *http.Request) {
	if host, ok := proxyMap[r.Host]; ok {
		if proxyMap[r.Host].TLSEnabled {
			log.Println(r.RemoteAddr, r.Host, r.URL.String())
			host.ReverseProxy.ServeHTTP(w, r)
			return
		}
		forwardHTTP(w, r)
		return
	}
	notFound(w, r)
}

// forwardHTTP checks the host name of HTTP traffic, if TLS is enabled, it
// re-writes the address and forwards the client to the the https website,
// other wise it forwards it to the appropriate service
func forwardHTTP(w http.ResponseWriter, r *http.Request) {
	if host, ok := proxyMap[r.Host]; ok {
		if proxyMap[r.Host].TLSEnabled {
			target := "https://" + r.Host + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
			return
		}
		log.Println(r.RemoteAddr, r.Host, r.URL.String())
		host.ReverseProxy.ServeHTTP(w, r)
		return
	}
	notFound(w, r)
}

// notFound is used If the user tries to visit a host that can't be found.
func notFound(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, r.Host, r.URL.String())
	w.Write([]byte("coming soon"))
}

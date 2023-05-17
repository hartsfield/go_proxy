package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// This program runs on port 8080 for http traffic and 8443 for https. This is
// so that the program won't need to run with administrative privileges.
//
// When running this on a new server, use iptables to redirect traffic from
// port 443 to port 8443, and from port 80 to port 8080, the following commands
// should achieve this on most Linux systems:
//
// sudo iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
// sudo iptables -t nat -A PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443
//
//
// IMPORTANT:
// NOTE: You need to run those iptables commands again after reboots.
// NOTE: When renewing certs, reboot, and make sure this program is not running.
// NOTE: After renewing certs, mv them to ~/tlsCerts and chown -R USER ~/tlsCerts/*
// NOTE: Make sure these files have the correct permissions, you likely copied
// them from root.

type tlsCerts struct {
	Privkey   string
	Fullchain string
}

type service struct {
	Port         string
	TLSEnabled   bool
	AlertsOn     bool
	DomainName   string
	ReverseProxy *httputil.ReverseProxy
}

var (
	globalHalt context.CancelFunc
	certs      tlsCerts = tlsCerts{
		Privkey:   os.Getenv("privkey"),
		Fullchain: os.Getenv("fullchain"),
	}
	httpPort          string      = ":8080"
	tlsPort           string      = ":8443"
	proxyMap                      = make(map[string]*service)
	genericServerConf http.Server = http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Second,
	}
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	file, err := os.Open("prox.config")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sc := strings.Split(scanner.Text(), ":")
		s := &service{
			Port:       sc[0],
			DomainName: sc[3],
		}
		if sc[1] == "true" {
			s.TLSEnabled = true
		}
		if sc[2] == "true" {
			s.AlertsOn = true
		}

		proxyMap[s.DomainName] = makeProxy(s)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	if certs.Fullchain == "" || certs.Privkey == "" {
		certs.Fullchain = "~/tlsCerts/fullchain.pem"
		certs.Privkey = "~/tlsCerts/privkey.pem"
	}
}

func main() {
	insecure := &genericServerConf
	insecure.Addr = httpPort
	insecure.Handler = http.HandlerFunc(forwardHTTP)

	secure := &genericServerConf
	secure.Addr = tlsPort
	secure.Handler = http.HandlerFunc(forwardTLS)

	ctx, cancel := context.WithCancel(context.Background())
	globalHalt = cancel

	// go startHTTPServer(insecure)
	go startTLSServer(secure)

	<-ctx.Done()
}

func startHTTPServer(s *http.Server) {
	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
	globalHalt()
}

func startTLSServer(s *http.Server) {
	err := s.ListenAndServeTLS(certs.Fullchain, certs.Privkey)
	if err != nil {
		log.Println(err)
	}
	globalHalt()
}

func forwardTLS(w http.ResponseWriter, r *http.Request) {
	if host, ok := proxyMap[r.Host]; ok {
		if proxyMap[r.Host].TLSEnabled {
			host.ReverseProxy.ServeHTTP(w, r)
			return
		}
		forwardHTTP(w, r)
		return
	}
	notFound(w, r)
}

func forwardHTTP(w http.ResponseWriter, r *http.Request) {
	if host, ok := proxyMap[r.Host]; ok {
		if proxyMap[r.Host].TLSEnabled {
			tlsRedirect(w, r)
			return
		}
		host.ReverseProxy.ServeHTTP(w, r)
		return
	}
	notFound(w, r)
}

// tlsRedirect is used to re-write the host name and redirect the user to
// the secure website via https.
func tlsRedirect(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

// notFound is used If the user tries to visit a host that can't be found.
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("coming soon"))
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

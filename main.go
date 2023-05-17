package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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
// NOTE: Make sure these files have the correct permissions, you likely
// copied them from root.

type tlsCerts struct {
	Privkey   string
	Fullchain string
}

type service struct {
	DomainName   string
	Port         string
	ReverseProxy *httputil.ReverseProxy
	TLSEnabled   bool
}

var (
	certs tlsCerts = tlsCerts{
		Privkey:   os.Getenv("privkey"),
		Fullchain: os.Getenv("fullchain"),
	}
	httpPort string = ":8080"
	tlsPort  string = ":8443"
	proxyMap        = make(map[string]*service)
	services        = []*service{
		{
			DomainName: "mysterygift.org",
			Port:       "8050",
			TLSEnabled: true,
		},
		{
			DomainName: "btstrmr.xyz",
			Port:       "5555",
			TLSEnabled: true,
		},
		{
			DomainName: "tagmachine.xyz",
			Port:       "9001",
			TLSEnabled: true,
		},
		{
			DomainName: "telesoft.network",
			Port:       "9002",
			TLSEnabled: true,
		},
		{
			DomainName: "sbvrt.telesoft.network",
			Port:       "9669",
			TLSEnabled: true,
		},
		{
			DomainName: "particlestore.telesoft.network",
			Port:       "8667",
			TLSEnabled: true,
		},
		{
			DomainName: "tsconsulting.telesoft.network",
			Port:       "9047",
			TLSEnabled: true,
		},
		{
			DomainName: "generic.telesoft.network",
			Port:       "9677",
			TLSEnabled: true,
		},
		{
			DomainName: "anglewood.telesoft.network",
			Port:       "4420",
			TLSEnabled: true,
		},
	}
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UTC().UnixNano())

	for _, service := range services {
		proxyMap[service.DomainName] = makeProxy(service)
	}

	if certs.Fullchain == "" || certs.Privkey == "" {
		certs.Fullchain = "~/tlsCerts/fullchain.pem"
		certs.Privkey = "~/tlsCerts/privkey.pem"
	}
}

func main() {
	insecure := &http.Server{
		Addr:              httpPort,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Second,
		Handler:           http.HandlerFunc(secureEntryPoint),
	}
	secure := &http.Server{
		Addr:              tlsPort,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Second,
		Handler:           http.HandlerFunc(secureEntryPoint),
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		err := insecure.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
		cancelCtx()
	}()
	go func() {
		err := secure.ListenAndServeTLS(certs.Fullchain, certs.Privkey)
		if err != nil {
			log.Println(err)
		}
		cancelCtx()
	}()

	<-ctx.Done()
}

// secureEntryPoint is used to re-write the host name and redirect the user to
// the secure website via https.
func secureEntryPoint(w http.ResponseWriter, r *http.Request) {
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
	u, err := url.Parse("localhost:" + s.Port)
	if err != nil {
		log.Println(err)
	}
	s.ReverseProxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", u.Host)
			req.URL.Host = u.Host
			if s.TLSEnabled {
				req.URL.Scheme = "https"
			} else {
				req.URL.Scheme = "http"
			}
		},
		FlushInterval: -1,
	}
	return s
}

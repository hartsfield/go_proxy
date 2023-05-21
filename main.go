package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// PROX
//
// prox80=http_port
// prox443=https_port
// proxConf=/path/to/configuraton
// privkey=/path/to/privkey
// fullchain=/path/to/fullchain
//
// ex.
//
// prox80=8080 prox443=8443 proxConf=prox.config privkey=~/tlsCerts/privkey.pem fullchain=~/tlsCerts/fullchain.pem prox
//
// It is advised to forward traffic on port :80 (HTTP) and :443 (HTTPS/TLS) to
// higher ports which dont require administrator privileges. By default prox
// runs on port :8080 for HTTP, and port :8443 for HTTPS traffic.You can change
// these defaults by run with different environment variables.
//
// When restarting the server, you can use iptables to redirect traffic from
// port :443 to port :8443, and from port :80 to port :8080, or whatever your
// desired prts may be. The following commands should achieve this on most
// Linux systems:
//
// sudo iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
// sudo iptables -t nat -A PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443
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
	certs      *tlsCerts = &tlsCerts{
		Privkey:   os.Getenv("privkey"),
		Fullchain: os.Getenv("fullchain"),
	}
	httpPort string              = os.Getenv("prox80")
	tlsPort  string              = os.Getenv("prox443")
	confPath string              = os.Getenv("proxConf")
	proxyMap map[string]*service = make(map[string]*service)
)

func main() {
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

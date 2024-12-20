package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

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
func main() {
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: 9213,
		IP:   net.ParseIP("127.0.0.1"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	go func() {
		for {
			_, _, err := ser.ReadFromUDP(p)
			if string(p)[:6] == "reload" {
				scan()
			}

			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()

	insecure := newServerConf(httpPort, http.HandlerFunc(forwardHTTP))
	secure := newServerConf(tlsPort, http.HandlerFunc(forwardTLS))

	ctx, cancel := context.WithCancel(context.Background())
	globalHalt = cancel

	go startHTTPServer(insecure)
	go startTLSServer(secure)

	fmt.Println()
	fmt.Println("Started services:")
	fmt.Println()
	for s, v := range pc.Services {
		if !strings.Contains(s, "www.") {
			fmt.Println("  ->", v.App.Port, s)
		}
	}
	fmt.Println()

	<-ctx.Done()
}

type MyRoundTripper struct{}

func (t *MyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header["X-Forwarded-For"] = []string{req.RemoteAddr}
	return http.DefaultTransport.RoundTrip(req)
}

// makeProxy takes var #SERVICE *service{} and creates a *http.ReverseProxy
// using the properties of #SERVICE
func makeProxy(s *serviceConf) *serviceConf {
	u, err := url.Parse("http://localhost:" + s.App.Port + "/")
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
		FlushInterval: 0,
		// FlushInterval: -1,
		Transport: &MyRoundTripper{},
	}
	return s
}

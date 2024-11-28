package main

import (
	"bufio"
	"context"
	"log"
	"net/http/httputil"
	"os"
	"strings"
)

// tlsCerts are used for the tls server
type tlsCerts struct {
	Privkey   string
	Fullchain string
}

// service is a type of application running on a port
type service struct {
	// Port is the port on which the application is runnning
	Port         string
	DomainName   string
	TLSEnabled   bool
	AlertsOn     bool
	ReverseProxy *httputil.ReverseProxy
}

var (
	// globalHalt is used to safely shutdown the server int he event of an
	// error
	globalHalt context.CancelFunc
	// certs is used for the TLS server
	certs *tlsCerts = &tlsCerts{
		Privkey:   os.Getenv("privkey"),
		Fullchain: os.Getenv("fullchain"),
	}
	// httpPort is the port your server recieves http traffic on (port 80
	// not recommended)
	httpPort string = os.Getenv("prox80")
	// tlsPort is the port your server recieves https traffic on (port 443
	// not recommended)
	tlsPort string = os.Getenv("prox443")
	// confPath is the path to this programs configuraton file
	confPath string = os.Getenv("proxConf")
	// proxyMap is a map of host names to services running on the server.
	proxyMap map[string]*service = make(map[string]*service)
	// f        *os.File
)

// init sets flags that tell log to log the date and line number. Init also
// reads the configuration file
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	file, err := os.Open(confPath)
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
		proxyMap["www."+s.DomainName] = makeProxy(s)
	}

	if err := scanner.Err(); err != nil {
		log.Panicln(err)
	}

	if certs.Fullchain == "" || certs.Privkey == "" {
		certs.Fullchain = "~/tlsCerts/fullchain.pem"
		certs.Privkey = "~/tlsCerts/privkey.pem"
	}
}

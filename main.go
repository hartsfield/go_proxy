package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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
// You may need to run these commands again after restarts.

// proxyMap is a hashmap of our website hostnames to their proxy.
var proxyMap = make(map[string]*httputil.ReverseProxy)

// If you want to modify this proxy for a different set of websites, follow the
// pattern below. origin, director, proxyMap
func init() {
	//////// MysteryGift.org running on port 8050
	originMysteryGift, _ := url.Parse("http://localhost:8050/")
	directorMysteryGift := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originMysteryGift.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originMysteryGift.Host
	}

	// add to proxyMap
	proxyMap["mysterygift.org"] =
		&httputil.ReverseProxy{Director: directorMysteryGift}

	//////// TagMachine.TeleSoft.network running on port 9001
	originTagMachine, _ := url.Parse("http://localhost:9001/")
	directorTagMachine := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originTagMachine.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originTagMachine.Host
	}

	///////// TeleSoft.network running on port 9002
	originTeleSoft, _ := url.Parse("http://localhost:9002/")
	directorTeleSoft := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originTeleSoft.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originTeleSoft.Host
	}

	// add to proxyMap
	proxyMap["telesoft.network"] =
		&httputil.ReverseProxy{Director: directorTeleSoft}
	proxyMap["tagmachine.telesoft.network"] =
		&httputil.ReverseProxy{Director: directorTagMachine}

	///////// tonedef.TeleSoft.network running on port 9003
	originToneDef, _ := url.Parse("http://localhost:9003/")
	directorToneDef := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originToneDef.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originToneDef.Host
	}

	// add to proxyMap
	proxyMap["tonedef.telesoft.network"] =
		&httputil.ReverseProxy{Director: directorToneDef}

	originTSC, _ := url.Parse("http://localhost:9047/")
	directorTSC := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originTSC.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originTSC.Host
	}

	// add to proxyMap
	proxyMap["tsconsulting.telesoft.network"] =
		&httputil.ReverseProxy{Director: directorTSC}

	originAngle, _ := url.Parse("http://localhost:4420/")
	directorAngle := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originAngle.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originAngle.Host
	}

	// add to proxyMap
	proxyMap["anglewood.telesoft.network"] =
		&httputil.ReverseProxy{Director: directorAngle}

	originBtstrmr, _ := url.Parse("http://localhost:5555/")
	directorBtstrmr := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originBtstrmr.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originBtstrmr.Host
	}

	// add to proxyMap
	proxyMap["btstrmr.xyz"] =
		&httputil.ReverseProxy{Director: directorBtstrmr}

}

func main() {
	// run insecureEntryPoint() when users visit the server
	http.HandleFunc("/", insecureEntryPoint)

	// Start a TLS (HTTPS) server, with links to files generated by
	// letsencrypt.
	// NOTE: When renewing certs, make sure this program is not running,
	// and you have reset iptables so that it doesn't redirect traffic
	// NOTE: Make sure these files have the correct permissions, you likely
	// copied them from root.
	go http.ListenAndServeTLS(":8443", os.Getenv("fullchain"), os.Getenv("privkey"), nil)

	// start an http server
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(upgradeToTLS)))
}

// insecureEntryPoint is used when we cant upgrade to TLS
func insecureEntryPoint(w http.ResponseWriter, r *http.Request) {
	// check the host name and make sure it exists in our proxyMap. then
	// redirect the user to the appropriate port
	if host, ok := proxyMap[r.Host]; ok {
		host.ServeHTTP(w, r)
		return
	}
	// else redirect the user to the not found page
	notFound(w, r)
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

// upgradeToTLS checks the host, if we have certs for the host we upgrade their
// connection to TLS secured (https) using secureEntryPoint(). Otherwise we
// use insecureEntryPoint() to attempt to send them to the insecured page.
func upgradeToTLS(w http.ResponseWriter, r *http.Request) {
	switch r.Host {
	case "mysterygift.org":
		secureEntryPoint(w, r)
	case "telesoft.network":
		secureEntryPoint(w, r)
	case "tagmachine.telesoft.network":
		secureEntryPoint(w, r)
	case "tonedef.telesoft.network":
		secureEntryPoint(w, r)
	case "tsconsulting.telesoft.network":
		secureEntryPoint(w, r)
	default:
		insecureEntryPoint(w, r)
	}
}

// notFound is used If the user tries to visit a host that can't be found.
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("coming soon"))
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var proxyMap = make(map[string]*httputil.ReverseProxy)

func init() {
	originMysteryGift, _ := url.Parse("https://localhost:8050/")
	directorMysteryGift := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originMysteryGift.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originMysteryGift.Host
	}

	proxyMap["mysterygift.org"] =
		&httputil.ReverseProxy{Director: directorMysteryGift}

	originTagMachine, _ := url.Parse("http://localhost:9001/")
	directorTagMachine := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originTagMachine.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originTagMachine.Host
	}

	proxyMap["beta.mysterygift.org"] =
		&httputil.ReverseProxy{Director: directorTagMachine}

	originTeleSoft, _ := url.Parse("http://localhost:9002/")
	directorTeleSoft := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originTeleSoft.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originTeleSoft.Host
	}

	proxyMap["telesoft.network"] =
		&httputil.ReverseProxy{Director: directorTeleSoft}
	proxyMap["tagmachine.telesoft.network"] =
		&httputil.ReverseProxy{Director: directorTagMachine}

	originToneDef, _ := url.Parse("http://localhost:9003/")
	directorToneDef := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originToneDef.Host)
		req.URL.Scheme = "http"
		req.URL.Host = originToneDef.Host
	}

	proxyMap["tonedef.telesoft.network"] =
		&httputil.ReverseProxy{Director: directorToneDef}

}

func main() {
	http.HandleFunc("/", entryPoint)

	go http.ListenAndServeTLS(":8443",
		"/home/john/mgTLSFiles/fullchain.pem",
		"/home/john/mgTLSFiles/privkey.pem", nil)

	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(upgradeToTLS)))
}

func entryPoint(w http.ResponseWriter, r *http.Request) {
	if host, ok := proxyMap[r.Host]; ok {
		host.ServeHTTP(w, r)
		return
	}
	notFound(w, r)
}

func upgradeToTLS(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)
	if r.Host == "mysterygift.org" {
		target := "https://" + r.Host + r.URL.Path
		if len(r.URL.RawQuery) > 0 {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	} else {
		fmt.Println(r.Host)
		entryPoint(w, r)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("coming soon"))
}

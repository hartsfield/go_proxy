package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http/httputil"
	"os"
)

// config is the configuration file for bolt-proxy
type config struct {
	ProxyDir     string                  `json:"proxy_dir"`
	HttpPort     string                  `json:"http_port"`
	TLSPort      string                  `json:"tls_port"`
	AdminUser    string                  `json:"admin_user"`
	LiveDir      string                  `json:"live_dir"`
	StageDir     string                  `json:"stage_dir"`
	CertDir      string                  `json:"cert_dir"`
	TlsCerts     *tlsCerts               `json:"tls_certs"`
	ServiceRepos []string                `json:"service_repos"`
	Services     map[string]*serviceConf `json:"services"`
}

// tlsCerts are used for the tls server
type tlsCerts struct {
	Privkey   string `json:"privkey"`
	Fullchain string `json:"fullchain"`
}

type env map[string]string

// serviceConf is a type of application running on a port
type serviceConf struct {
	App          app    `json:"app"`
	GCloud       gcloud `json:"gcloud"`
	ReverseProxy *httputil.ReverseProxy
}

type app struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Env        env    `json:"env"`
	Port       string `json:"port"`
	AlertsOn   bool   `json:"alertsOn"`
	DomainName string `json:"domain_name"`
	TLSEnabled bool   `json:"tls_enabled"`
}

type gcloud struct {
	Command   string `json:"command"`
	Zone      string `json:"zone"`
	Project   string `json:"project"`
	User      string `json:"user"`
	LiveDir   string `json:"livedir"`
	ProxyConf string `json:"proxyConf"`
}

var (
	globalHalt context.CancelFunc
	certs      *tlsCerts = pc.TlsCerts
	httpPort   string    = pc.HttpPort
	tlsPort    string    = pc.TLSPort
	confPath   string    = os.Getenv("proxConfPath")
	pc         config    = config{}
)

// init sets flags that tell log to log the date and line number. Init also
// reads the configuration file
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc.Services = make(map[string]*serviceConf)
	proxyConf()
	scan()
}

func scan() {
	dir, err := os.ReadDir(pc.LiveDir)
	if err != nil {
		log.Println(err)
	}
	for _, d := range dir {
		b, err := os.ReadFile(pc.LiveDir + d.Name() + "/bolt.conf.json")
		if err != nil {
			log.Println(err)
		}
		sc := serviceConf{}
		err = json.Unmarshal(b, &sc)
		if err != nil {
			log.Println(err)
		}

		pc.Services[sc.App.DomainName] = makeProxy(&sc)
		pc.Services["www."+sc.App.DomainName] = pc.Services[sc.App.DomainName]
	}
}

func proxyConf() {
	if len(confPath) < 1 {
		confPath = "/home/john/go_proxy/prox.conf"
	}
	file, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &pc)
	if err != nil {
		log.Println(err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httputil"
	"os"
)

// config is the configuration file for bolt-proxy
type config struct {
	ProxyDir     string `json:"proxy_dir"`
	AdminUser    string `json:"admin_user"`
	LiveDir      string `json:"live_dir"`
	StageDir     string `json:"stage_dir"`
	CertDir      string `json:"cert_dir"`
	ServiceRepos []string
	TlsCerts     tlsCerts
	Services     map[string]serviceConf
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
	AlertsOn   string `json:"alertsOn"`
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
	confPath      string = os.Getenv("proxConf")
	proxyConfPath string = os.Getenv("proxConfPath")
	// proxyMap is a map of host names to services running on the server.
	proxyMap map[string]serviceConf = make(map[string]serviceConf)

	// fMap map[string]*stringFlag = make(map[string]*stringFlag)
	pc *config
)

// type stringFlag struct {
// 	set   bool
// 	value string
// 	do    func()
// }

// func (sf *stringFlag) Set(x string) error {
// 	sf.value = x
// 	sf.set = true
// 	return nil
// }

// func (sf *stringFlag) String() string {
// 	return sf.value
// }

// init sets flags that tell log to log the date and line number. Init also
// reads the configuration file
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// fMap["scan"] = &stringFlag{do: scan}
	// flag.Var(fMap["scan"], "deploy", "Deploys project to server")

	if len(confPath) < 1 {
		confPath = "/home/john/prox.conf"
	}

	proxyConf()
	scan()
}

func scan() {
	dir, err := os.ReadDir(pc.LiveDir)
	if err != nil {
		log.Println(err)
	}
	for _, d := range dir {
		b, err := os.ReadFile(d.Name() + "/bolt.conf.json")
		if err != nil {
			log.Println(err)
		}
		sc := serviceConf{}
		err = json.Unmarshal(b, &sc)
		if err != nil {
			log.Println(err)
		}

		pc.Services[sc.App.DomainName] = makeProxy(sc)
		pc.Services["www."+sc.App.DomainName] = pc.Services[sc.App.DomainName]
	}
}

func proxyConf() {
	if len(proxyConfPath) < 1 {
		proxyConfPath = "/home/john/go_proxy/prox.conf"
	}
	file, err := os.ReadFile(proxyConfPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, pc)
	if err != nil {
		log.Println(err)
	}
}

// func conf() {
// 	file, err := os.Open(confPath)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		sc := strings.Split(scanner.Text(), ":")
// 		s := &serviceConf{
// 			Port:       sc[0],
// 			DomainName: sc[3],
// 		}
// 		if sc[1] == "true" {
// 			s.TLSEnabled = true
// 		}
// 		if sc[2] == "true" {
// 			s.AlertsOn = true
// 		}

// 		proxyMap[s.DomainName] = makeProxy(s)
// 		proxyMap["www."+s.DomainName] = makeProxy(s)
// 	}

// 	if err := scanner.Err(); err != nil {
// 		log.Panicln(err)
// 	}

// 	if certs.Fullchain == "" || certs.Privkey == "" {
// 		certs.Fullchain = "~/tlsCerts/fullchain.pem"
// 		certs.Privkey = "~/tlsCerts/privkey.pem"
// 	}
// }

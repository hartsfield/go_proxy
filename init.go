package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
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
	}

	if err := scanner.Err(); err != nil {
		log.Panicln(err)
	}

	if certs.Fullchain == "" || certs.Privkey == "" {
		certs.Fullchain = "~/tlsCerts/fullchain.pem"
		certs.Privkey = "~/tlsCerts/privkey.pem"
	}
}
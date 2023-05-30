#!/bin/bash
privkey=~/tlsCerts/privkey.pem
fullchain=~/tlsCerts/fullchain.pem
prox80=8080
prox443=8443
proxConf=prox.config
trap -- '' SIGTERM
git pull
go build -o go_proxy
pkill -f go_proxy
nohup ./go_proxy > /dev/null & disown
sleep 2

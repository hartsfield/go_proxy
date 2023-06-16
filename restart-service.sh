#!/bin/bash
export privkey=~/tlsCerts/privkey
export fullchain=~/tlsCerts/fullchain
export prox80=8080
export prox443=8443
export proxConf=~/prox.conf
# trap -- '' SIGTERM
git pull
go build -o go_proxy
pkill -f go_proxy
nohup ./go_proxy > /dev/null & disown
# sleep 2

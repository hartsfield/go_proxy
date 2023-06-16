#!/bin/bash
export privkey=~/tlsCerts/privkey.pem
export fullchain=~/tlsCerts/fullchain.pem
export prox80=8080
export prox443=8443
export proxConf=~/prox.conf
# trap -- '' SIGTERM
git pull
go build -o go_proxy
pkill -f go_proxy
# nohup ./go_proxy > /dev/null & disown
./go_proxy
sleep 2

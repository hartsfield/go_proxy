# go_proxy

This is a small program that acts as a proxy server for my websites. You can 
point multiple hostnames at one IP address, and once configured, this proxy 
will read the host name and direct the client to the appropriate port based on 
that hostname. 

It can easily be modified for your websites. See the comments in `main.go`. 

NOTE:

This program runs on port `8080` for http connections and port `8443` for https.
This is so you won't need to run it with administrator priviliges. To make this
program work properly you should use the following command (for linux) to 
redirect traffic from port `80` and `443` respectively:

    sudo iptables -t nat -A PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443
    sudo iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080

When starting the server, tell goprox where your private key and fullchain 
files are by specifying the `fullchain` and `privkey` environment variables at 
runtime:

    fullchain=/path/to/fullchain.pem privkey=/path/to/privkey.pem ./goprox &

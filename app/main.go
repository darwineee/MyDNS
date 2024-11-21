package main

import (
	"com.sentry.dev/app/server"
	"flag"
)

func main() {
	resolver := flag.String("resolver", "", "Host is used to lookup addresses")
	flag.Parse()
	if *resolver == "" {
		*resolver = "1.1.1.1:53"
	}
	udpServer := server.UDPServer{
		RecursiveHost: resolver,
	}
	udpServer.Start()
}

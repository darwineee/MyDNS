package main

import (
	"bufio"
	"com.sentry.dev/app/server"
	"flag"
	"fmt"
	"os"
	"runtime"
)

func main() {
	resolver := flag.String("resolver", "1.1.1.1:53", "Host is used to lookup addresses")
	numWorkers := flag.Int("workers", runtime.NumCPU()*2, "Number of workers")
	flag.Parse()
	udpServer := server.UDPServer{
		RecursiveHost: resolver,
		Workers:       make(chan struct{}, *numWorkers),
	}

	go udpServer.Start()

	commandChan := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			commandChan <- scanner.Text()
		}
	}()

	for {
		select {
		case cmd := <-commandChan:
			fmt.Println("--------------------------------")
			switch cmd {
			case "stop":
				fmt.Println("Stopping server...")
				udpServer.Stop()
				fmt.Println("Server stopped gracefully")
				return
			default:
				fmt.Println("Unknown command:", cmd)
			}
		}
	}
}

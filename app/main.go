package main

import (
	"bufio"
	"com.sentry.dev/app/config"
	"com.sentry.dev/app/server"
	"com.sentry.dev/app/utils"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	appConfig := config.Load()
	ctx, cancel := context.WithCancel(context.Background())

	udpServer := server.UDPServer{
		Config:     appConfig,
		Context:    ctx,
		CancelFunc: cancel,
	}

	utils.PrintBanner()
	udpServer.Start()
	fmt.Println("Server started successfully!")
	utils.PrintSeparator()

	commandChan := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			commandChan <- scanner.Text()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case cmd := <-commandChan:
			switch cmd {
			case "stop":
				stopServer(&udpServer)
				return
			default:
				fmt.Println("Unknown command ", cmd)
			}
			utils.PrintSeparator()
		case <-ctx.Done():
			stopServer(&udpServer)
			return
		case <-sigChan:
			stopServer(&udpServer)
			return
		}
	}
}

func stopServer(server *server.UDPServer) {
	fmt.Println("Stopping server...")
	server.Stop()
	fmt.Println("Server stopped gracefully")
}

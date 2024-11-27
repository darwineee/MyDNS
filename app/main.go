package main

import (
	"bufio"
	"com.sentry.dev/app/config"
	"com.sentry.dev/app/server"
	"com.sentry.dev/app/utils"
	"fmt"
	"os"
)

func main() {
	appConfig := config.Load()

	udpServer := server.UDPServer{
		Config: appConfig,
	}

	utils.PrintBanner()
	go udpServer.Start()
	fmt.Println("Server started successfully!")
	utils.PrintSeparator()

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
			switch cmd {
			case "stop":
				fmt.Println("Stopping server...")
				udpServer.Stop()
				fmt.Println("Server stopped gracefully")
				return
			default:
				fmt.Println("Unknown command ", cmd)
			}
			utils.PrintSeparator()
		}
	}
}

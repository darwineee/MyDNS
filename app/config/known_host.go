package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetKnownHosts() map[string]string {
	knownHosts := make(map[string]string)
	file, err := os.Open("known_hosts")
	if err != nil {
		fmt.Println("Error opening known_hosts file:", err)
		return knownHosts
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing known_hosts file:", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			fmt.Println("Invalid known_hosts file format")
			continue
		}
		knownHosts[parts[0]] = parts[1]
	}
	return knownHosts
}

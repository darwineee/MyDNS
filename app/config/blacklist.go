package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetBlackList() []string {
	var blackList []string
	file, err := os.Open("blacklist")
	if err != nil {
		fmt.Println("Error opening known_hosts file:", err)
		return blackList
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Println("Error closing known_hosts file:", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		blackList = append(blackList, scanner.Text())
	}
	return blackList
}

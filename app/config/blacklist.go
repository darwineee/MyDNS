package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetBlackList(filePath string) []string {
	var blackList []string
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return blackList
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Println(err)
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

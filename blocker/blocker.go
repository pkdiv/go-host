package blocker

import (
	"bufio"
	"fmt"
	"os"
	"slices"
)

func ExtractDomain(data []byte) string {

	if len(data) < 12 {
		return ""
	}

	pos := 12
	var domain string

	for {

		length := int(data[pos])
		if length == 0 {
			break
		}

		if pos+length > len(data) {
			break
		}

		if domain != "" {
			domain += "."
		}

		domain += string(data[pos+1 : pos+length+1])
		pos += length + 1

	}

	return domain

}

func IsBlocked(domain string) bool {

	blockerList, err := LoadBlockList()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return slices.Contains(blockerList, domain)

}

func LoadBlockList() ([]string, error) {

	file, err := os.Open("blocked_domains")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var blockedList []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		blockedList = append(blockedList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return blockedList, nil

}

func LoadAllowList() ([]string, error) {

	file, err := os.Open("allow_domains")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var allowList []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allowList = append(allowList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return allowList, nil

}

func IsAllowed(domain string) bool {

	allowList, err := LoadAllowList()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return slices.Contains(allowList, domain)

}

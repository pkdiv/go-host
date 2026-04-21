package logs

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	logFile *os.File
	mu      sync.Mutex
)

func InitLogFile() {

	var err error
	logFile, err = os.OpenFile("dns_queries.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func LogQuery(domain string, clientIP string, status string) {

	if logFile == nil {
		InitLogFile()
	}

	mu.Lock()
	defer mu.Unlock()

	fmt.Fprintf(logFile, "Time: %s, Domain: %s, Status: %s, Client IP: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		domain,
		status,
		clientIP)

}

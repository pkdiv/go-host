package main

import (
	"fmt"
	"go-host/blocker"
	"go-host/security"
	"net"
	"sync"
	"time"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func main() {
	if err := StartServer(); err != nil {
		fmt.Println(err)
	}
}

func StartServer() error {

	addr, err := net.ResolveUDPAddr("udp", ":53")
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Println("DNS Server running on port ", addr.Port)

	for {
		buffer := bufPool.Get().([]byte)
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		fmt.Println("Request received from client")

		domain := blocker.ExtractDomain(buffer[:n])
		if blocker.IsBlocked(domain) {
			fmt.Println("Domain blocked: ", domain)
			// TODO: send proper dns response
			conn.WriteToUDP([]byte("Domain blocked"), clientAddr)
			continue
		}

		go HandleRequest(buffer[:n], clientAddr, conn)

	}

}

func UpstreamDNS(data []byte) ([]byte, error) {

	UpstreamDNS := "1.1.1.1:53"

	upstreamConn, err := net.DialTimeout("udp", UpstreamDNS, 5*time.Second)
	if err != nil {
		fmt.Printf("Error connecting to upstream DNS: %s", UpstreamDNS)
		return nil, err
	}
	defer upstreamConn.Close()

	fmt.Printf("Connected to upstream DNS: %s\n", UpstreamDNS)

	upstreamConn.SetReadDeadline(time.Now().Add(5 * time.Second))

	fmt.Printf("Sending request to upstream DNS\n")
	_, err = upstreamConn.Write(data)
	if err != nil {
		return nil, err
	}

	respBuffer := make([]byte, 4096)
	n, err := upstreamConn.Read(respBuffer)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Response received from upstream DNS: %d bytes\n", n)

	return respBuffer[:n], nil
}

func HandleRequest(data []byte, clientAddr *net.UDPAddr, conn *net.UDPConn) {

	limiter := security.NewClientLimiter(1*time.Minute, 10)

	if !limiter.Allow(clientAddr.IP.String()) {
		fmt.Println("Client rate limited")
		// TODO: send proper dns response
		conn.WriteToUDP([]byte("Client rate limited"), clientAddr)
		return
	}

	recivedData := data
	resp, err := UpstreamDNS(recivedData)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.WriteToUDP(resp, clientAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Response sent to client")

}

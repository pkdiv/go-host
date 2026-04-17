package main

import (
	"fmt"
	"net"
	"time"
)

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

	fmt.Println("DNS Server running on :53")

	for {
		buffer := make([]byte, 4096)
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		fmt.Println("Request received from client")

		recivedData := buffer[:n]
		resp, err := UpstreamDNS(recivedData)
		if err != nil {
			fmt.Println(err)
			continue
		}

		_, err = conn.WriteToUDP(resp, addr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Response sent to client")

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

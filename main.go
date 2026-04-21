package main

import (
	"fmt"
	"go-host/blocker"
	"go-host/logs"
	"go-host/security"
	"net"
	"sync"
	"time"
)

type RCODE int

const (
	NOERROR RCODE = iota
	FORMAT_ERROR
	SERVER_FAILURE
	NXDOMAIN
	NOT_IMPLEMENTED
	REFUSED
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
		fmt.Println("Domain: ", domain)
		if !blocker.IsAllowed(domain) {
			if blocker.IsBlocked(domain) {
				logs.LogQuery(domain, clientAddr.IP.String(), "Blocked")
				resp := blockResponse(buffer[:n], NXDOMAIN)
				conn.WriteToUDP(resp, clientAddr)
				continue
			}
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
	domain := blocker.ExtractDomain(data)

	if !limiter.Allow(clientAddr.IP.String()) {
		fmt.Println("Client rate limited")
		resp := blockResponse(data, SERVER_FAILURE)
		logs.LogQuery(domain, clientAddr.IP.String(), "Rate Limited")
		conn.WriteToUDP(resp, clientAddr)
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

	logs.LogQuery(domain, clientAddr.IP.String(), "Success")

	fmt.Println("Response sent to client")

}

func blockResponse(data []byte, rcode RCODE) []byte {

	// DNS Message Structure
	// Header
	// 		ID (2 bytes)
	// 		Flags (2 bytes)
	// 			QR (1 bit, 0-query & 1-response)
	//	 		Opcode (4 bits, 0-standard query) 0 for most cases
	//	 		AA (1 bit, 0-not authoritative & 1-authoritative)
	//	 		TC (1 bit, 0-not truncated & 1-truncated) if 1 client may retry over TCP
	//	 		RD (1 bit, 0-recursion not desired & 1-recursion desired) Set by client, tells the server to recurisvely resolve the query
	//	 		RA (1 bit, 0-recursion not available & 1-recursion available) Set by server, tells the client that the response was from a recursive query
	//	 		Z (1 bit, 0-reserved) Always 0
	//	 		AD (1 bit, 0-not authenticated & 1-authenticated) Set by server, tells the client that the response was authenticated, i.e DNSSEC validation
	//	 		CD (1 bit, 0-not checking & 1-checking) Set by client, tells the server to check the response, i.e to skip DNSSEC validation
	//	 		RCODE (4 bits, 0-no error, 1-format error, 2-server failure, 3-name error (NXDOMAIN), 4-not implemented, 5-refused)
	// 	QDCOUNT (2 bytes)
	// 		Tells the client that there is one question
	// 	ANCOUNT (2 bytes)
	//  	Tells the client that there are no answers
	// 	NSCOUNT (2 bytes)
	// 		Tells the client that there are no authority records ,i.e tells the client who is authoritative for the domain
	// 	ARCOUNT (2 bytes)
	// 		Tells the client that there are no additional records, i.e extra information about the domain to prevent further queries
	// Question Section
	// 		QNAME (variable length)
	// 		QTYPE (2 bytes)
	// 		QCLASS (2 bytes)
	// Answer Section
	// 		NAME (2 bytes)
	// 		TYPE (2 bytes)
	// 		CLASS (2 bytes)
	// 		TTL (4 bytes)
	// 		RDLENGTH (2 bytes)
	// 		RDATA (variable length)
	// Authority Section
	// 		NAME (2 bytes)
	// 		TYPE (2 bytes)
	// 		CLASS (2 bytes)
	// 		TTL (4 bytes)
	// 		RDLENGTH (2 bytes)
	// 		RDATA (variable length)
	// Additional Section
	// 		NAME (2 bytes)
	// 		TYPE (2 bytes)
	// 		CLASS (2 bytes)
	// 		TTL (4 bytes)
	// 		RDLENGTH (2 bytes)
	// 		RDATA (variable length)

	resp := make([]byte, len(data))
	copy(resp, data)

	resp[2] |= 0x80 // QR
	resp[3] &= 0xF0 // Clear the RCODE field

	switch rcode {
	case NOERROR:
		resp[3] |= 0x00
	case FORMAT_ERROR:
		resp[3] |= 0x01
	case SERVER_FAILURE:
		resp[3] |= 0x02
	case NXDOMAIN:
		resp[3] |= 0x03
	case NOT_IMPLEMENTED:
		resp[3] |= 0x04
	case REFUSED:
		resp[3] |= 0x05
	}
	resp[6], resp[7] = 0, 0   // ANCOUNT
	resp[8], resp[9] = 0, 0   // NSCOUNT
	resp[10], resp[11] = 0, 0 // ARCOUNT

	offset := 12

	for {

		if offset >= len(resp) {
			return resp
		}

		length := int(resp[offset])
		offset += 1

		if length == 0 {
			break
		}

		offset += length

	}

	if offset >= len(resp) {
		return resp
	}

	resp = resp[:offset+4]

	return resp
}

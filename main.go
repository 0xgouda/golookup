package main

import (
	"fmt"
	"net"
	"os"
)

// Fixed Root DNS Servers addresses
const (
	A_ROOT_SERVER = "198.41.0.4"
)

func SendQuery(queryPacket []byte, serverAddr string) DNSResponse {
	// For simplicity lets hope it will make it to the server and back :)
	socket, _ := net.Dial("udp", serverAddr + ":53")
	defer socket.Close()

	socket.Write(queryPacket)

	buf := make([]byte, 1024)
	socket.Read(buf)
	resp := ParseDNSResponse(buf)

	if resp.Header.ANcount > 0 {
		return resp
	} 
	
	var nsIp string
	var nsAns []DNSRecord
	if resp.Header.ARcount > 0  {
		nsAns = resp.AdditionalRecords
	} else {
		nsQuery := GenerateDNSQuery(resp.NameServers[0].RData)
		nsAns = SendQuery(nsQuery, A_ROOT_SERVER).Answers
	}
	
	for _, ans := range nsAns {
		if ans.Type_ == A_TYPE {
			nsIp = ans.RData
		}
	}

	return SendQuery(queryPacket, nsIp)
}

func main() {
	queryPakcet := GenerateDNSQuery(os.Args[1])
	resp := SendQuery(queryPakcet, A_ROOT_SERVER)

	for _, ans := range resp.Answers {
		fmt.Println("Domain:", ans.DomainName, "IP:", ans.RData)
	}
}

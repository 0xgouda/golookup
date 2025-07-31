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
	fmt.Println("Querying:", serverAddr)

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

	if serverAddr == A_ROOT_SERVER {
		fmt.Printf("Received TLD Server Address for \"%s\": \"%s\"", resp.NameServers[0].DomainName, resp.NameServers[0].RData)
	} else {
		fmt.Printf("Received Authoritative Server Address for \"%s\": \"%s\"", resp.NameServers[0].DomainName, resp.NameServers[0].RData)
	}
	fmt.Println()
	
	var nsIp string
	var nsAns []DNSRecord
	if resp.Header.ARcount > 0  {
		nsAns = resp.AdditionalRecords
	} else {
		fmt.Println()
		fmt.Println("IP not in packet, starting new query for:", resp.NameServers[0].RData)

		nsQuery := GenerateDNSQuery(resp.NameServers[0].RData)
		nsAns = SendQuery(nsQuery, A_ROOT_SERVER).Answers
	}
	
	for _, ans := range nsAns {
		if ans.Type_ == A_TYPE {
			nsIp = ans.RData
		}
	}

	fmt.Println("Resolved IP:", nsIp)
	if resp.Header.ARcount == 0 {
		fmt.Println("new query done")
	}
	fmt.Println()

	return SendQuery(queryPacket, nsIp)
}

func main() {
	fmt.Println("Fixed Root Server IP:", A_ROOT_SERVER)
	fmt.Println()

	queryPakcet := GenerateDNSQuery(os.Args[1])
	resp := SendQuery(queryPakcet, A_ROOT_SERVER)
	fmt.Println("Received:")

	for _, ans := range resp.Answers {
		fmt.Println("Domain:", ans.DomainName, "IP:", ans.RData)
	}
}

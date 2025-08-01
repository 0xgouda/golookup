package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// Fixed Root DNS Servers addresses
// From: https://www.iana.org/domains/root/servers
const (
	A_ROOT_SERVER = "198.41.0.4"
)

func SendDNSQuery(queryPacket []byte, serverAddr string) (*DNSResponse, error) {
	fmt.Println("Querying:", serverAddr)

	socket, _ := net.Dial("udp", serverAddr + ":53")
	defer socket.Close()

	buf := make([]byte, 1024)
	var err error
	for range 5 {
		socket.SetDeadline(time.Now().Add(5 * time.Second))
		_, err = socket.Write(queryPacket)
		if err == nil {
			_, err = socket.Read(buf)
			if err == nil {
				break
			}
		}
	}

	if errors.Is(err, os.ErrDeadlineExceeded) {
		return nil, fmt.Errorf("error connecting to address %s, UDP packets didn't make it", serverAddr)
	}

	resp := ParseDNSResponse(buf)
	if resp.Header.ANcount > 0 {
		return resp, nil
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
		nsResp, err := SendDNSQuery(nsQuery, A_ROOT_SERVER)
		if err != nil {
			return nil, err
		}
		nsAns = nsResp.Answers
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

	return SendDNSQuery(queryPacket, nsIp)
}

func main() {
	fmt.Println("Fixed Root Server IP:", A_ROOT_SERVER)
	fmt.Println()

	queryPakcet := GenerateDNSQuery(os.Args[1])
	resp, err := SendDNSQuery(queryPakcet, A_ROOT_SERVER)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Received:")

	for _, ans := range resp.Answers {
		fmt.Println("Domain:", ans.DomainName, "IP:", ans.RData)
	}
}

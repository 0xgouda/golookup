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

func Resolve(query DNSQuery) (*DNSResponse, error) {
	serverToQuery := A_ROOT_SERVER
	for {
		fmt.Println("Querying:", serverToQuery)
		buf, err := SendDNSQuery(query, serverToQuery)
		if err != nil {
			return nil, err
		}

		resp := ParseDNSResponse(buf)
		if resp.Header.ANcount > 0 {
			// answer found, return.
			return resp, nil
		}

		ns, nsIp := GetNameServerToQuery(resp)
		if len(resp.NameServers) == 0 || ns == "" {
			return nil, fmt.Errorf("no answer or name servers records found in packet")
		}

		if serverToQuery == A_ROOT_SERVER {
			fmt.Printf("Received TLD Server Address for \"%s\": \"%s\"", query.Questions[0].Qname, ns)
		} else {
			fmt.Printf("Received Authoritative Server Address for \"%s\": \"%s\"", query.Questions[0].Qname, ns)
		}
		fmt.Println()
		
		if nsIp == "" {
			fmt.Println()
			fmt.Println("name server IP not in packet")
			fmt.Println("starting new query for:", ns)

			newQuery := GenerateDNSQuery(ns, A_TYPE)
			nsResp, err := Resolve(newQuery)
			if err != nil {
				return nil, err
			}
			nsIp = nsResp.Answers[0].RData
			fmt.Println("new query done, found name server IP")
			fmt.Println()
		}
		
		serverToQuery = nsIp
		fmt.Println("Resolved IP:", nsIp)
		fmt.Println()
	}
}

func SendDNSQuery(query DNSQuery, serverAddr string) ([]byte, error) {
	socket, _ := net.Dial("udp", serverAddr + ":53")
	defer socket.Close()

	buf := make([]byte, 1024)
	var err error
	for range 5 {
		socket.SetDeadline(time.Now().Add(5 * time.Second))
		_, err = socket.Write(query.ToBytes())
		if err == nil {
			_, err = socket.Read(buf)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return nil, fmt.Errorf("error connecting to address %s, UDP packets didn't make it", serverAddr)
		}
		return nil, err
	} 

	return buf, nil
}

func GetNameServerToQuery(resp *DNSResponse) (string, string) {
	for _, ar := range resp.AdditionalRecords {
		if ar.Type == A_TYPE {
			return ar.DomainName, ar.RData
		}
	}

	for _, nsRecord := range resp.NameServers {
		if nsRecord.RData != "" {
			return nsRecord.RData, ""
		}
	}	

	return "", ""
}
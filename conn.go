package main

import (
	"fmt"
	"net"
	"time"
)

// Fixed Root DNS Servers addresses
// From: https://www.iana.org/domains/root/servers
const (
	A_ROOT_SERVER = "198.41.0.4"
)

// Iterates through the DNS hierarchy to get query answer
func Resolve(query DNSPacket) (DNSPacket, error) {
	serverToQuery := A_ROOT_SERVER
	for {
		buf, err := SendDNSQuery(query, serverToQuery)
		if err != nil {
			return DNSPacket{}, err
		}

		resp := ParseDNSPacket(buf)
		if resp.Header.ANcount > 0 {
			// answer found
			return resp, nil
		}

		if resp.Header.NScount == 0 {
			return DNSPacket{}, fmt.Errorf("no answer or name server records found in packet")
		}
		ns, nsIp := GetNameServer(resp)

		if nsIp == "" {
			fmt.Println("name server IP not found in packet")
			fmt.Println("starting new query for:", ns)
			newQuery := GenerateDNSQuery(ns, A_TYPE)
			nsResp, err := Resolve(newQuery)
			if err != nil {
				return DNSPacket{}, err
			}
			nsIp = nsResp.Answers[0].RData
			fmt.Printf("new query done, found \"%s\" IP: \"%s\"", ns, nsIp)
		}
		serverToQuery = nsIp
	}
}

// Sends the DNS query to server address over UDP with timeout and retrial
func SendDNSQuery(query DNSPacket, serverAddr string) ([]byte, error) {
	fmt.Println("Querying:", serverAddr)
	conn, _ := net.Dial("udp", serverAddr + ":53")
	defer conn.Close()

	buf := make([]byte, 1024)
	var err error
	for range 5 {
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		_, err = conn.Write(query.ToBytes())
		if err == nil {
			_, err = conn.Read(buf)
			if err == nil {
				break
			}
		}
	}
	return buf, err
}

// Gets domain name and ip for a name server record from 
// the given dns response packet.
//
// If any field is not found it returns its zero value.
func GetNameServer(resp DNSPacket) (string, string) {
	for i, nsr := range resp.NameServers {
		for _, ar := range resp.AdditionalRecords {
			if ar.Type == A_TYPE && ar.DomainName == nsr.RData {
				fmt.Printf("Received Name Server for \"%s\": \"%s\"\n", nsr.DomainName, nsr.RData)
				fmt.Printf("Received \"%s\" IP: \"%s\"\n", ar.DomainName, ar.RData)
				return nsr.RData, ar.RData
			}
		}

		if i + 1 == len(resp.NameServers) {
			fmt.Printf("Received Name Server for \"%s\": \"%s\"\n", nsr.DomainName, nsr.RData)
			return nsr.RData, ""
		}
	}	
	return "", ""
}

// Starts a DNS server listening at the given port
func serveDNS(port int) error {
	addr := net.UDPAddr{
        Port: port,
        IP:   net.ParseIP("0.0.0.0"),
    }
    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        return err
    }
	conn.SetReadDeadline(time.Time{}) // No Read Deadline
    defer conn.Close()

    for {
		buffer := make([]byte, 1024)
        _, addr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println(err)
            continue
        }

		query := ParseDNSPacket(buffer)
		newQuery := GenerateDNSQuery(query.Questions[0].Qname, query.Questions[0].Qtype)
		resp, err := Resolve(newQuery)
		if err != nil {
            fmt.Println(err)
			continue
		}
		resp.Header.Id = query.Header.Id
		resp.Header.Flags = 1 << 15

		for range 5 {
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err = conn.WriteToUDP(resp.ToBytes(), addr)
			if err == nil {
				break
			}
		}
		if err != nil {
			fmt.Println(err)
		}
    }
}
package main

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
)

type DNSResponse struct {
	Header 			  DNSHeader
	Questions         []DNSQuestion
	Answers           []DNSRecord
	NameServers       []DNSRecord
	AdditionalRecords []DNSRecord
}

type DNSRecord struct {
	DomainName   string
	Type_        uint16
	Class        uint16
	TTL          uint32
	RDLength     uint16
	RData        string
}

func ParseDNSHeader(buf *bytes.Buffer) DNSHeader {
	header := DNSHeader{}
	binary.Read(buf, binary.BigEndian, &header)
	return header
}

func ParseDomainName(buf *bytes.Buffer) string {
	var labels []string
	var compressionMask uint8 = 0b1100_0000

	for {
		var num uint8
		binary.Read(buf, binary.BigEndian, &num)
		// check if DNS domain name compression is used
		if (num & compressionMask) != 0 {
			// compression is used, revert the read byte
			// and treat the next 2 bytes as 0b11(compression mask)offset
			// and start reading from offset
			buf.UnreadByte()
			var index uint16
			binary.Read(buf, binary.BigEndian, &index)
			index &= 0b00

			buf = bytes.NewBuffer(buf.Bytes()[index:])
			num, _ = buf.ReadByte()
		} 

		if num == 0 {
			break
		}
		labelBytes := buf.Next(int(num))
		labels = append(labels, string(labelBytes))
	}
		
	return strings.Join(labels, ".")
}

func ParseDNSRecord(buf *bytes.Buffer) DNSRecord {
	record := DNSRecord{
		DomainName: ParseDomainName(buf),
	}
	binary.Read(buf, binary.BigEndian, &record.Type_)
	binary.Read(buf, binary.BigEndian, &record.Class)
	binary.Read(buf, binary.BigEndian, &record.TTL)
	binary.Read(buf, binary.BigEndian, &record.RDLength)

	var RData [4]byte
	binary.Read(buf, binary.BigEndian, &RData)

	var octets [4]string
	for i, num := range RData {
		octets[i] = strconv.Itoa(int(num))
	}
	record.RData = strings.Join(octets[:], ".")

	return record
}

func ParseDNSQuestion(buf *bytes.Buffer) DNSQuestion {
	question := DNSQuestion{
		Qname: ParseDomainName(buf),
	}
	binary.Read(buf, binary.BigEndian, &question.Qtype)
	binary.Read(buf, binary.BigEndian, &question.Qclass)

	return question
}

func ParseDNSResponse(buf []byte) DNSResponse {
	bytesBuf := bytes.NewBuffer(buf)

	header := ParseDNSHeader(bytesBuf)
	response := DNSResponse{
		Header: header,
		Questions: make([]DNSQuestion, 0, header.QDcount),
		Answers: make([]DNSRecord, 0, header.ANcount),
		NameServers: make([]DNSRecord, 0, header.NScount),
		AdditionalRecords: make([]DNSRecord, 0, header.ARcount),
	}

	for range header.QDcount {
		response.Questions = append(response.Questions, ParseDNSQuestion(bytesBuf))
	}

	for range header.ANcount {
		response.Answers = append(response.Answers, ParseDNSRecord(bytesBuf))
	}

	for range header.NScount {
		response.NameServers = append(response.AdditionalRecords, ParseDNSRecord(bytesBuf))
	}

	for range header.ARcount {
		response.AdditionalRecords = append(response.AdditionalRecords, ParseDNSRecord(bytesBuf))
	}

	return response
}
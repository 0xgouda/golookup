package main

import (
	"bytes"
	"encoding/binary"
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
	RData        []byte
}

func ParseDNSHeader(buf []byte) DNSHeader {
	iobuf := bytes.NewBuffer(buf)
	header := DNSHeader{}
	binary.Read(iobuf, binary.BigEndian, &header)
	return header
}

func ParseDomainName(buf []byte) (string, uint8) {
	var labels []string
	var index uint8 = 0

	for index < uint8(len(buf)) {
		num := buf[index]
		if num == 0 {
			break
		}
		label := string(buf[index + 1:index + 1 + num])
		labels = append(labels, label)
		index += num + 1
	}
	return strings.Join(labels, "."), index
}

func ParseDNSRecord(buf []byte) (DNSRecord, uint8) {
	domain, newStartIdx := ParseDomainName(buf)
	record := DNSRecord{
		DomainName: domain,
		Type_: binary.BigEndian.Uint16(buf[newStartIdx:]),
		Class: binary.BigEndian.Uint16(buf[newStartIdx + 2:]),
		TTL: binary.BigEndian.Uint32(buf[newStartIdx + 6:]),
		RDLength: binary.BigEndian.Uint16(buf[newStartIdx + 8:]),
		RData: buf[newStartIdx + 10:],
	}
	var nextRecStartIdx uint8 = newStartIdx + 10 + uint8(record.RDLength)
	record.RData = buf[newStartIdx + 10: nextRecStartIdx]
	return record, nextRecStartIdx
}

func ParseDNSQuestion(buf []byte) (DNSQuestion, uint8){
	domain, byteIdxAfterDomain := ParseDomainName(buf)
	recType := RecordType(binary.BigEndian.Uint16(buf[byteIdxAfterDomain:]))
	classType := QueryClass(binary.BigEndian.Uint16(buf[byteIdxAfterDomain + 2:]))
	question := DNSQuestion{
		Qname: domain,
		Qtype: recType,
		Qclass: classType,
	}
	return question, byteIdxAfterDomain + 4
}

func ParseDNSResponse(buf []byte) DNSResponse {
	header := ParseDNSHeader(buf)
	response := DNSResponse{
		Header: header,
		Questions: make([]DNSQuestion, header.QDcount),
		Answers: make([]DNSRecord, header.ANcount),
		NameServers: make([]DNSRecord, header.NScount),
		AdditionalRecords: make([]DNSRecord, header.ARcount),
	}
	buf = buf[DNSHeaderSize:]

	for range header.QDcount {
		question, newStartIdx := ParseDNSQuestion(buf)
		response.Questions = append(response.Questions, question)
		buf = buf[newStartIdx:]
	}

	for range header.ANcount {
		record, newStartIdx := ParseDNSRecord(buf)
		response.Answers = append(response.Answers, record)
		buf = buf[newStartIdx:]
	}

	for range header.NScount {
		record, newStartIdx := ParseDNSRecord(buf)
		response.NameServers = append(response.AdditionalRecords, record)
		buf = buf[newStartIdx:]
	}

	for range header.ARcount {
		record, newStartIdx := ParseDNSRecord(buf)
		response.AdditionalRecords = append(response.AdditionalRecords, record)
		buf = buf[newStartIdx:]
	}

	return response
}
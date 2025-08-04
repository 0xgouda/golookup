package main

import (
	"bytes"
	"encoding/binary"
	"io"
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
	Type_        RecordType
	Class        QueryClass
	TTL          uint32
	RDLength     uint16
	RData        string
}

func ParseDNSHeader(buf *bytes.Reader) DNSHeader {
	header := DNSHeader{}
	binary.Read(buf, binary.BigEndian, &header)
	return header
}

func ParseDomainName(buf *bytes.Reader) string {
	var labels []string
	var compressionMask uint8 = 0b1100_0000
	var currPos int64 = -1

	for {
		var num uint8
		binary.Read(buf, binary.BigEndian, &num)
		if num == 0 {
			break
		}

		// check if DNS domain name compression is used
		if (num & compressionMask) != 0 {
			// compression used, revert the read byte
			// and treat the next 2 bytes as 0b11(compression mask)+offset
			// and start reading from offset
			buf.UnreadByte()
			var offset uint16
			binary.Read(buf, binary.BigEndian, &offset)
			offset &= 0b0011_1111_1111_1111

			if currPos == -1 {
				currPos, _ = buf.Seek(0, io.SeekCurrent)
			}
			buf.Seek(int64(offset), io.SeekStart)
			continue
		} 

		labelBytes := make([]byte, num)
		buf.Read(labelBytes)
		labels = append(labels, string(labelBytes))
	}
		
	if currPos != -1 {
		buf.Seek(currPos, io.SeekStart)
	}
	return strings.Join(labels, ".")
}

func ParseTXTRdata(buf *bytes.Reader, RDLength uint16) string {
	var strs []string
	var readLen uint16

	for readLen < RDLength {
		var num uint8
		binary.Read(buf, binary.BigEndian, &num)
		readLen += uint16(num) + 1
		strBytes := make([]byte, num)
		buf.Read(strBytes)
		strs = append(strs, string(strBytes))
	}
	return strings.Join(strs, "")
}

func ParseDNSRecord(buf *bytes.Reader) DNSRecord {
	record := DNSRecord{
		DomainName: ParseDomainName(buf),
	}
	binary.Read(buf, binary.BigEndian, &record.Type_)
	binary.Read(buf, binary.BigEndian, &record.Class)
	binary.Read(buf, binary.BigEndian, &record.TTL)
	binary.Read(buf, binary.BigEndian, &record.RDLength)

	switch record.Type_ {
	case NS_TYPE, CNAME_TYPE:
		record.RData = ParseDomainName(buf)
	case MX_TYPE:
		var preference uint16
		binary.Read(buf, binary.BigEndian, &preference)
		record.RData = strconv.Itoa(int(preference)) + " "
		record.RData += ParseDomainName(buf)
	case A_TYPE:
		var RData [4]byte
		binary.Read(buf, binary.BigEndian, &RData)

		var octets [4]string
		for i, num := range RData {
			octets[i] = strconv.Itoa(int(num))
		}
		record.RData = strings.Join(octets[:], ".")
	case TXT_TYPE:
		record.RData = ParseTXTRdata(buf, record.RDLength)
	default:
		// move buf cursor and ignore the data
		buf.Read(make([]byte, record.RDLength))
	}

	return record
}

func ParseDNSQuestion(buf *bytes.Reader) DNSQuestion {
	question := DNSQuestion{
		Qname: ParseDomainName(buf),
	}
	binary.Read(buf, binary.BigEndian, &question.Qtype)
	binary.Read(buf, binary.BigEndian, &question.Qclass)

	return question
}

func ParseDNSResponse(buf []byte) *DNSResponse {
	bytesBuf := bytes.NewReader(buf)

	header := ParseDNSHeader(bytesBuf)
	response := &DNSResponse{
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
		response.NameServers = append(response.NameServers, ParseDNSRecord(bytesBuf))
	}

	for range header.ARcount {
		response.AdditionalRecords = append(response.AdditionalRecords, ParseDNSRecord(bytesBuf))
	}

	return response
}
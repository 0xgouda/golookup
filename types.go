package main

import (
	"bytes"
	"encoding/binary"
	"strings"
)

type RecordType uint16

const (
	A_TYPE     RecordType = 1
	NS_TYPE    RecordType = 2
	CNAME_TYPE RecordType = 5
	MX_TYPE    RecordType = 15
	TXT_TYPE   RecordType = 16
)

type QueryClass uint16

const (
	Internet QueryClass = 1
)

type DNSHeader struct {
	Id      uint16
	Flags   uint16
	QDcount uint16
	ANcount uint16
	NScount uint16
	ARcount uint16
}

func (h *DNSHeader) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, h)
	return buf.Bytes()
}

type DNSQuestion struct {
	Qname  string
	Qtype  RecordType
	Qclass QueryClass
}

func (q *DNSQuestion) ToBytes() []byte {
	buf := new(bytes.Buffer)

	var label string
	domain := q.Qname
	for domain != "" {
		label, domain, _ = strings.Cut(domain, ".")

		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	buf.WriteByte(0)

	binary.Write(buf, binary.BigEndian, q.Qtype)
	binary.Write(buf, binary.BigEndian, q.Qclass)

	return buf.Bytes()
}

type DNSQuery struct {
	Header   DNSHeader
	Questions []DNSQuestion
}

func (query DNSQuery) ToBytes() []byte {
	buf := query.Header.ToBytes()
	for _, question := range query.Questions {
		buf = append(buf, question.ToBytes()...)
	}
	return buf
}

type DNSResponse struct {
	Header 			  DNSHeader
	Questions         []DNSQuestion
	Answers           []DNSRecord
	NameServers       []DNSRecord
	AdditionalRecords []DNSRecord
}

type DNSRecord struct {
	DomainName   string
	Type         RecordType
	Class        QueryClass
	TTL          uint32
	RDLength     uint16
	RData        string
}

var incrementalId uint16 = 0
func GenerateDNSQuery(domain string, qtype RecordType) DNSQuery {
	header := DNSHeader{
		Id: incrementalId,
		QDcount: 1,
	}
	incrementalId++

	question := DNSQuestion{
		Qname: domain,
		Qtype: qtype,
		Qclass: Internet,
	}

	return DNSQuery{
		Header: header,
		Questions: []DNSQuestion{question},
	}
}
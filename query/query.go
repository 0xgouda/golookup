package query

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

func (h *DNSHeader) to_bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, h)
	return buf.Bytes()
}

type DNSQuestion struct {
	Qname  string
	Qtype  RecordType
	Qclass QueryClass
}

func (q *DNSQuestion) to_bytes() []byte {
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

var incrementalId uint16 = 0
func GenerateDNSQuery(domain string, qtype RecordType) []byte {
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

	buf := header.to_bytes()
	return append(buf, question.to_bytes()...)
}
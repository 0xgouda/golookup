package dns

import (
	"bytes"
	"encoding/binary"
	"strconv"
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
	writeDomainToBuffer(q.Qname, buf)
	binary.Write(buf, binary.BigEndian, q.Qtype)
	binary.Write(buf, binary.BigEndian, q.Qclass)
	return buf.Bytes()
}

type DNSRecord struct {
	DomainName   string
	Type         RecordType
	Class        QueryClass
	TTL          uint32
	RDLength     uint16
	RData        string
}

func (r DNSRecord) ToBytes() []byte {
	buf := new(bytes.Buffer)
	writeDomainToBuffer(r.DomainName, buf)
	binary.Write(buf, binary.BigEndian, r.Type)
	binary.Write(buf, binary.BigEndian, r.Class)
	binary.Write(buf, binary.BigEndian, r.TTL)

	switch r.Type {
	case A_TYPE:
		binary.Write(buf, binary.BigEndian, r.RDLength)
		octets := strings.Split(r.RData, ".")
		for _, o := range octets {
			num, _ := strconv.Atoi(o)
			binary.Write(buf, binary.BigEndian, byte(num))
		}
	case NS_TYPE, CNAME_TYPE:
		// len(r.RData) + 2 instead of r.RDLength because it may 
		// have been calculated after domain name compression, 
		// which we don't support for now.
		binary.Write(buf, binary.BigEndian, uint16(len(r.RData) + 2))
		writeDomainToBuffer(r.RData, buf)
	case MX_TYPE:
		preference, smtpDomain, _ := strings.Cut(r.RData, " ")
		preferenceInt, _ := strconv.Atoi(preference)
		// len(r.RData) + 4 instead of r.RDLength because it may 
		// have been calculated after domain name compression, 
		// which we don't support for now.
		binary.Write(buf, binary.BigEndian, uint16(len(smtpDomain) + 4))
		binary.Write(buf, binary.BigEndian, uint16(preferenceInt))
		writeDomainToBuffer(smtpDomain, buf)
	case TXT_TYPE:
		// TODO
	}

	return buf.Bytes()
}

type DNSPacket struct {
	Header            DNSHeader
	Questions         []DNSQuestion
	Answers           []DNSRecord
	NameServers       []DNSRecord
	AdditionalRecords []DNSRecord
}

func (p DNSPacket) ToBytes() []byte {
	// Temp Workaround: 
	// don't put AdditionalRecords in packet 
	// and set ARcount (by value) to 0 to avoid 
	// breaking client parsers as we currently 
	// don't support AAAA records that may be there
	p.Header.ARcount = 0
	//for _, ar := range p.AdditionalRecords {
		//buf = append(buf, ar.ToBytes()...)
	//}

	var buf []byte
	buf = append(buf, p.Header.ToBytes()...)

	for _, q := range p.Questions {
		buf = append(buf, q.ToBytes()...)
	}

	for _, ans := range p.Answers {
		buf = append(buf, ans.ToBytes()...)
	}

	for _, n := range p.NameServers {
		buf = append(buf, n.ToBytes()...)
	}

	return buf
}

var incrementalId uint16 = 0
// Creates a DNS Query with domain and qtype fields in the question. 
func GenerateDNSQuery(domain string, qtype RecordType) DNSPacket {
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

	return DNSPacket{
		Header: header,
		Questions: []DNSQuestion{question},
	}
}

func GetQtype(type_ string) RecordType {
	var qtype RecordType
	switch strings.ToUpper(type_) {
	case "NS":
		qtype = NS_TYPE
	case "CNAME":
		qtype = CNAME_TYPE
	case "MX":
		qtype = MX_TYPE
	case "TXT":
		qtype = TXT_TYPE
	default:
		qtype = A_TYPE
	}
	return qtype
}

// writes domain name to buffer as a sequence of <character-string>s. 
func writeDomainToBuffer(domainName string, buf *bytes.Buffer) {
	labels := strings.Split(domainName, ".")
	for _, label := range labels { 
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	buf.WriteByte(0)
}
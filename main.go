package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("usage: golookup [domain] [A|NS|CNAME|MX|TXT]")
		return 
	}

	var qtype RecordType
	var resultName string
	switch strings.ToUpper(args[1]) {
	case "A":
		qtype = A_TYPE
		resultName = "IP"
	case "NS":
		qtype = NS_TYPE
		resultName = "Name Server"
	case "CNAME":
		qtype = CNAME_TYPE
		resultName = "CNAME"
	case "MX":
		qtype = MX_TYPE
		resultName = "Mail Server"
	case "TXT":
		qtype = TXT_TYPE
		resultName = "Text"
	default:
		fmt.Println("unsupported record type", os.Args[2])
		return
	}

	fmt.Println("Fixed Root Server IP:", A_ROOT_SERVER)
	fmt.Println()

	query := GenerateDNSQuery(args[0], qtype)
	resp, err := Resolve(query)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Answers:")
	for _, ans := range resp.Answers {
		fmt.Println(resultName + ":", ans.RData)
	}
}
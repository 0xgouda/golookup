package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/0xgouda/golookup/conn"
	"github.com/0xgouda/golookup/query"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("usage: golookup [domain] [A|NS|CNAME|MX|TXT]")
		return 
	}

	var qtype query.RecordType
	var resultName string
	switch strings.ToUpper(args[1]) {
	case "A":
		qtype = query.A_TYPE
		resultName = "IP"
	case "NS":
		qtype = query.NS_TYPE
		resultName = "Name Server"
	case "CNAME":
		qtype = query.CNAME_TYPE
		resultName = "CNAME"
	case "MX":
		qtype = query.MX_TYPE
		resultName = "Mail Server"
	case "TXT":
		qtype = query.TXT_TYPE
		resultName = "Text"
	default:
		fmt.Println("unsupported record type", os.Args[2])
		return
	}

	fmt.Println("Fixed Root Server IP:", conn.A_ROOT_SERVER)
	fmt.Println()

	resp, err := conn.Resolve(args[0], qtype)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Answers:")
	for _, ans := range resp.Answers {
		fmt.Println(resultName + ":", ans.RData)
	}
}
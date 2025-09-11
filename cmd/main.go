package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/0xgouda/golookup/dns"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 || (args[0] != "serve" && len(args) < 2) {
		fmt.Println("usage:")
		fmt.Println("\tgolookup [domain] [A|NS|MX|CNAME|TXT]")
		fmt.Println("\tgolookup serve --port=<port-number>")
		return 
	}

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	servePort := serveCmd.Int("port", 8080, "server Port number.")
	if args[0] == "serve" {
		serveCmd.Parse(args[1:])
		err := dns.ServeDNS(*servePort)
		if err != nil {
			log.Fatal(err)
		}
	}


	fmt.Println("Fixed Root Server IP:", dns.A_ROOT_SERVER)
	qtype := dns.GetQtype(args[1])
	query := dns.GenerateDNSQuery(args[0], qtype)
	resp, err := dns.Resolve(query)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Answers:")
	for _, ans := range resp.Answers {
		fmt.Println(ans.RData)
	}
}
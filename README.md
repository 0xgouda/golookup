# GoLookup

**golookup** is a simple, dig-like DNS client written in Go.  

It was developed to get hands-on with networking and deepen my understanding of DNS resolution.  


The tool supports `A`, `MX`, `CNAME`, `TXT`, and `NS` queries, and logs each step of the recursive resolution process.

### Usage Examples

```
go build .
./golookup 
usage: golookup [domain] [A|NS|CNAME|MX|TXT]
```

```
./golookup google.com a   
Fixed Root Server IP: 198.41.0.4

Querying: 198.41.0.4
Received TLD Server Address for "google.com": "l.gtld-servers.net"
Resolved IP: 192.41.162.30

Querying: 192.41.162.30
Received Authoritative Server Address for "google.com": "ns2.google.com"
Resolved IP: 216.239.34.10

Querying: 216.239.34.10
Answers:
IP: 142.251.37.206
```

```
./golookup facebook.com mx
Fixed Root Server IP: 198.41.0.4

Querying: 198.41.0.4
Received TLD Server Address for "facebook.com": "l.gtld-servers.net"
Resolved IP: 192.41.162.30

Querying: 192.41.162.30
Received Authoritative Server Address for "facebook.com": "a.ns.facebook.com"
Resolved IP: 129.134.30.12

Querying: 129.134.30.12
Answers:
Mail Server: 10 smtpin.vvv.facebook.com
```

```
./golookup example.com ns
Fixed Root Server IP: 198.41.0.4

Querying: 198.41.0.4
Received TLD Server Address for "example.com": "l.gtld-servers.net"
Resolved IP: 192.41.162.30

Querying: 192.41.162.30
Received Authoritative Server Address for "example.com": "a.iana-servers.net"

name server IP not in packet
starting new query for: a.iana-servers.net
Querying: 198.41.0.4
Received TLD Server Address for "a.iana-servers.net": "m.gtld-servers.net"
Resolved IP: 192.55.83.30

Querying: 192.55.83.30
Received Authoritative Server Address for "a.iana-servers.net": "a.iana-servers.net"
Resolved IP: 199.43.135.53

Querying: 199.43.135.53
new query done, found name server IP

Resolved IP: 199.43.135.53

Querying: 199.43.135.53
Answers:
Name Server: a.iana-servers.net
Name Server: b.iana-servers.net
```
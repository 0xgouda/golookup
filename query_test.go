package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNSQuerier(t *testing.T) {
	a := assert.New(t)

	query := GenerateDNSQuery("google.com")

	socket, err := net.Dial("udp", "8.8.8.8:53")
	a.NoError(err)
	defer socket.Close()

	_, err = socket.Write(query)
	a.NoError(err)

	buf := make([]byte, 1024)
	num, err := socket.Read(buf)
	a.NoError(err)
	a.Greater(num, 1)
}
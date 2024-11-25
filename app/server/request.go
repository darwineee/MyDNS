package server

import (
	"com.sentry.dev/app/dns"
	"net"
)

type Request struct {
	ClientAddr *net.UDPAddr
	Header     *dns.Header
	Questions  []*dns.Question
}

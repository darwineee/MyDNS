package server

import (
	"com.sentry.dev/app/config"
	"com.sentry.dev/app/dns"
	"errors"
	"log"
	"net"
	"time"
)

const (
	DefaultPort     = 2053
	DefaultProtocol = "udp"
)

type UDPServer struct {
	conn          *net.UDPConn
	Running       bool
	RecursiveHost *string
}

// Start the UDP DNS server
func (server *UDPServer) Start() {
	addr := &net.UDPAddr{
		Port: DefaultPort,
		IP:   net.IPv4zero,
	}
	conn, err := net.ListenUDP(DefaultProtocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	server.conn = conn
	server.run()
}

func (server *UDPServer) run() {
	if server.Running {
		return
	}
	server.Running = true
	for {
		if !server.Running {
			break
		}
		//handle request
		clientAddr, header, questions, err := server.HandleRequest()
		if err != nil {
			continue
		}
		// handle response
		err = server.HandleResponse(clientAddr, header, questions)
		if err != nil {
			continue
		}
	}
	if err := server.conn.Close(); err != nil {
		log.Println("close udp server:", err)
	}
}

func (server *UDPServer) lookUp(questions []*dns.Question) (answer []*dns.Answer, err error) {
	return nil, errors.New("feature not implemented yet")
}

func (server *UDPServer) forward(query []byte) (response []byte, err error) {
	var conn net.Conn
	if conn, err = net.Dial(DefaultProtocol, *server.RecursiveHost); err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		if err = conn.Close(); err != nil {
			log.Println("close connection to forwarding server:", err)
		}
	}(conn)
	response = make([]byte, config.PkgLimitRFC1035)
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if _, err = conn.Write(query); err != nil {
		return nil, err
	}
	var bytesRead int
	if bytesRead, err = conn.Read(response); err != nil {
		return nil, err
	}
	return response[:bytesRead], nil
}

package server

import (
	"com.sentry.dev/app/config"
	"com.sentry.dev/app/dns"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

const (
	DefaultPort     = 2053
	DefaultProtocol = "udp"
)

type UDPServer struct {
	RecursiveHost *string
	Workers       chan struct{}

	conn       *net.UDPConn
	eventQueue chan *Request
	shutdown   chan struct{}
	wg         sync.WaitGroup
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
	server.eventQueue = make(chan *Request, 1000)
	server.shutdown = make(chan struct{})

	server.wg.Add(2)
	go server.dispatchRequests()
	go server.processRequests()
}

// Stop the UDP DNS server gracefully
func (server *UDPServer) Stop() {
	close(server.shutdown)
	if err := server.conn.Close(); err != nil {
		log.Println("Error closing UDP server:", err)
	}
	close(server.eventQueue)
	server.wg.Wait()
}

func (server *UDPServer) dispatchRequests() {
	defer server.wg.Done()
	for {
		select {
		case <-server.shutdown:
			return
		default:
			req, err := server.HandleRequest()
			if err != nil {
				continue
			}
			select {
			case server.eventQueue <- req:
			case <-time.After(500 * time.Millisecond):
				log.Println("event queue is full")
			}
		}
	}
}

func (server *UDPServer) processRequests() {
	defer server.wg.Done()
	for {
		select {
		case <-server.shutdown:
			return
		case req, ok := <-server.eventQueue:
			if !ok {
				return
			}
			select {
			case <-server.shutdown:
				return
			case server.Workers <- struct{}{}:
				go func(req *Request) {
					defer func() { <-server.Workers }()
					if err := server.HandleResponse(req); err != nil {
						log.Println("handle response:", err)
					}
				}(req)
			}
		}
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

package server

import (
	"com.sentry.dev/app/config"
	"com.sentry.dev/app/utils"
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"sync"
	"time"
)

type UDPServer struct {
	Config     *config.Config
	Context    context.Context
	CancelFunc context.CancelFunc

	conn        *net.UDPConn
	cache       *redis.Client
	eventQueue  chan *Request
	workers     chan struct{}
	eventLoopGr sync.WaitGroup
}

// Start the UDP DNS server
func (server *UDPServer) Start() {
	server.configConnection()
	server.configRedis()

	server.eventQueue = make(chan *Request, 1000)
	server.workers = make(chan struct{}, server.Config.Server.Workers)

	server.eventLoopGr.Add(2)
	go server.dispatchRequests()
	go server.processRequests()
}

func (server *UDPServer) configConnection() {
	addr := &net.UDPAddr{
		Port: server.Config.Server.Port,
		IP:   net.IPv4zero,
	}
	conn, err := net.ListenUDP(server.Config.Server.Protocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	server.conn = conn
}

func (server *UDPServer) configRedis() {
	rHost := server.Config.Redis.Host
	rPass := server.Config.Redis.Password
	rDB := server.Config.Redis.DB
	server.cache = redis.NewClient(&redis.Options{
		Addr:     rHost,
		Password: rPass,
		DB:       rDB,
	})

	blackList := config.GetBlackList()
	server.cache.SAdd(server.Context, utils.BlackList, blackList)

	knownHosts := config.GetKnownHosts()
	server.cache.HSet(server.Context, utils.KnownHost, knownHosts)
}

// Stop the UDP DNS server gracefully
func (server *UDPServer) Stop() {
	server.CancelFunc()
	if err := server.conn.Close(); err != nil {
		log.Println("Error closing UDP server:", err)
	}
	close(server.eventQueue)
	if err := server.cache.Close(); err != nil {
		log.Println("Error closing memcached client:", err)
	}
	server.eventLoopGr.Wait()
}

func (server *UDPServer) dispatchRequests() {
	defer server.eventLoopGr.Done()
	for {
		select {
		case <-server.Context.Done():
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
	defer server.eventLoopGr.Done()
	for {
		select {
		case <-server.Context.Done():
			return
		case req, ok := <-server.eventQueue:
			if !ok {
				return
			}
			select {
			case server.workers <- struct{}{}:
				go func(req *Request) {
					defer func() { <-server.workers }()
					if err := server.HandleResponse(req); err != nil {
						log.Println("handle response:", err)
					}
				}(req)
			}
		}
	}
}

func (server *UDPServer) lookUp(hostName string) (ip string, err error) {
	ip, err = server.cache.HGet(server.Context, utils.KnownHost, hostName).Result()
	return
}

func (server *UDPServer) isBlackListed(hostName string) (yes bool, err error) {
	yes, err = server.cache.SIsMember(server.Context, utils.BlackList, hostName).Result()
	return
}

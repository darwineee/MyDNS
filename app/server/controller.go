package server

import (
	"com.sentry.dev/app/dns"
	_type "com.sentry.dev/app/dns/type"
	"com.sentry.dev/app/utils"
	"fmt"
	"net"
	"sync"
)

// HandleRequest handle incoming request from client
func (server *UDPServer) HandleRequest() (
	request *Request,
	err error,
) {
	buf := make([]byte, server.Config.UDP.PkgLimitRFC1035)
	_, clientAddr, err := server.conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	header, questions, err := dns.ParseMessage(buf, server.Config.UDP.PkgLimitRFC1035)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &Request{
		ClientAddr: clientAddr,
		Header:     header,
		Questions:  questions,
	}, nil
}

// HandleResponse prepare response and send it to client
func (server *UDPServer) HandleResponse(
	req *Request,
) error {
	result := make([]byte, server.Config.UDP.PkgLimitRFC1035)

	pointerMap, remaining, err := server.writeQuestions(result, req.Questions)
	if err != nil {
		return err
	}

	answers := server.processQuestions(req.Questions)

	ansCount, remaining, err := server.writeAnswers(remaining, req.Questions, answers, pointerMap)
	if err != nil {
		return err
	}

	req.Header.RecursionAvailable = true
	req.Header.QueryResponse = true
	if req.Header.OperationCode != 0 {
		req.Header.ResponseCode = 4
	}
	req.Header.AnswerCount = ansCount

	if ansCount == 0 && server.handleNoAnswer(req.ClientAddr, req.Header) != nil {
		return err
	}

	if _, err = req.Header.WriteTo(result); err != nil {
		return err
	}

	size := len(result) - len(remaining)
	if _, err = server.conn.WriteToUDP(result[:size], req.ClientAddr); err != nil {
		return err
	}

	return nil
}

func (server *UDPServer) writeQuestions(buf []byte, questions []*dns.Question) (
	pointerMap map[int]uint16,
	remaining []byte,
	err error,
) {
	pointerMap = make(map[int]uint16)
	var pos uint16 = 12
	remaining = buf[pos:]
	for i, question := range questions {
		remaining, err = question.WriteTo(remaining)
		if err != nil {
			return nil, buf, err
		}
		pointerMap[i] = pos
		pos += uint16(question.Size())
	}
	return pointerMap, remaining, nil
}

func (server *UDPServer) processQuestions(questions []*dns.Question) []net.IP {
	var lookUpWg sync.WaitGroup
	answers := make([]net.IP, len(questions))
	for i, question := range questions {
		lookUpWg.Add(1)
		go func(question *dns.Question, order int) {
			defer lookUpWg.Done()
			yes, _ := server.isBlackListed(question.Name.String)
			if yes {
				return
			}
			var addr []byte
			if ip, err := server.lookUp(question.Name.String); err == nil {
				addr = net.ParseIP(ip).To4()
				if addr != nil {
					answers[order] = addr
					return
				}
			}
			if ips, err := net.LookupIP(question.Name.String); err == nil {
				for _, ip := range ips {
					addr = ip.To4()
					if addr != nil {
						answers[order] = addr
						ipStr := ip.String()
						fmt.Println("Cached:", question.Name.String, "=>", ipStr)
						server.cache.HSet(server.Context, utils.KnownHost, question.Name.String, ipStr)
						server.cache.HExpire(
							server.Context,
							utils.KnownHost,
							server.Config.Server.CacheTTLDuration(),
							question.Name.String,
						)
						return
					}
				}
			}
		}(question, i)
	}
	lookUpWg.Wait()
	return answers
}

func (server *UDPServer) writeAnswers(
	buf []byte,
	questions []*dns.Question,
	answers []net.IP,
	pointerMap map[int]uint16,
) (ansCount uint16, remaining []byte, err error) {
	remaining = buf
	for i, answer := range answers {
		if answer != nil {
			pointer := _type.ToPointer(pointerMap[i])
			ansStruct := dns.Answer{
				Pointer:  pointer,
				Type:     questions[i].Type,
				Class:    questions[i].Class,
				TTL:      server.Config.Server.CacheTTLSec,
				RDLength: uint16(len(answer)),
				RData:    answer,
			}
			if remaining, err = ansStruct.WriteTo(remaining); err != nil {
				return 0, nil, err
			}
			ansCount++
		}
	}
	return ansCount, remaining, nil
}

func (server *UDPServer) handleNoAnswer(
	clientAddr *net.UDPAddr,
	respHeader *dns.Header,
) (err error) {
	buf := make([]byte, respHeader.Size())
	if _, err = respHeader.WriteTo(buf); err != nil {
		return err
	}
	if _, err = server.conn.WriteToUDP(buf, clientAddr); err != nil {
		return err
	}
	return nil
}

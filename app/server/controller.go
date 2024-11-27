package server

import (
	"com.sentry.dev/app/dns"
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
) (err error) {
	var answers []*dns.Answer
	if req.Header.RecursionDesired {
		req.Header.RecursionAvailable = true
		answers, err = server.lookUp(req.Questions)
		if err != nil {
			err = server.handleRecursiveResponse(req.ClientAddr, req.Header, *req.Header, req.Questions)
		} else {
			err = server.handleNormalResponse(req.ClientAddr, req.Header, req.Questions, answers)
		}
	} else {
		err = server.handleNormalResponse(req.ClientAddr, req.Header, req.Questions, answers)
	}

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (server *UDPServer) handleRecursiveResponse(
	clientAddr *net.UDPAddr,
	reqHeader *dns.Header,
	respHeader dns.Header,
	questions []*dns.Question,
) (err error) {
	result := make([]byte, server.Config.UDP.PkgLimitRFC1035)
	remaining := result[12:]
	reqHeader.QuestionCount = 1
	var ansCount uint16 = 0

	for _, question := range questions {
		if remaining, err = question.WriteTo(remaining); err != nil {
			return err
		}
	}
	var wg sync.WaitGroup
	answers := make([][]byte, len(questions))
	for i, question := range questions {
		wg.Add(1)
		go func(question *dns.Question, order int) {
			defer wg.Done()
			var query []byte
			if query, err = buildQuery(reqHeader, question); err != nil {
				return
			}
			var response []byte
			if response, err = server.forward(query); err == nil && len(response) > len(query) {
				answer := response[len(query):]
				answers[order] = answer
			}
		}(question, i)
	}
	wg.Wait()
	for _, answer := range answers {
		if answer != nil {
			copy(remaining, answer)
			remaining = remaining[len(answer):]
			ansCount++
		}
	}

	respHeader.QueryResponse = true
	if respHeader.OperationCode != 0 {
		respHeader.ResponseCode = 4
	}
	respHeader.AnswerCount = ansCount

	if ansCount == 0 && server.handleNoAnswer(clientAddr, &respHeader) != nil {
		return err
	}

	if _, err = respHeader.WriteTo(result); err != nil {
		return err
	}

	size := len(result) - len(remaining)
	if _, err = server.conn.WriteToUDP(result[:size], clientAddr); err != nil {
		return err
	}

	return nil
}

func buildQuery(
	reqHeader *dns.Header,
	question *dns.Question,
) ([]byte, error) {
	size := reqHeader.Size() + question.Size()
	res := make([]byte, size)

	remaining, err := reqHeader.WriteTo(res)
	if err != nil {
		return nil, err
	}

	_, err = question.WriteTo(remaining)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (server *UDPServer) handleNormalResponse(
	clientAddr *net.UDPAddr,
	respHeader *dns.Header,
	questions []*dns.Question,
	answers []*dns.Answer,
) (err error) {
	ansCount := uint16(len(answers))

	respHeader.QueryResponse = true
	if respHeader.OperationCode != 0 {
		respHeader.ResponseCode = 4
	}
	respHeader.AnswerCount = ansCount

	if ansCount == 0 && server.handleNoAnswer(clientAddr, respHeader) != nil {
		return err
	}

	message := &dns.Message{
		Header:    respHeader,
		Questions: questions,
		Answers:   answers,
	}
	buf := make([]byte, message.Size())
	if _, err = message.WriteTo(buf); err != nil {
		return err
	}
	if _, err = server.conn.WriteToUDP(buf, clientAddr); err != nil {
		return err
	}
	return nil
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

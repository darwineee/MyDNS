package server

import (
	"com.sentry.dev/app/config"
	"com.sentry.dev/app/dns"
	"fmt"
	"net"
)

// HandleRequest handle incoming request from client
func (server *UDPServer) HandleRequest() (
	clientAddr *net.UDPAddr,
	header *dns.Header,
	questions []*dns.Question,
	err error,
) {
	buf := make([]byte, config.PkgLimitRFC1035)
	_, clientAddr, err = server.conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	header, questions, err = dns.ParseMessage(buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// HandleResponse prepare response and send it to client
func (server *UDPServer) HandleResponse(
	clientAddr *net.UDPAddr,
	header *dns.Header,
	questions []*dns.Question,
) (err error) {
	var answers []*dns.Answer
	if header.RecursionDesired {
		header.RecursionAvailable = true
		answers, err = server.lookUp(questions)
		if err != nil {
			err = server.handleRecursiveResponse(clientAddr, header, *header, questions)
		} else {
			err = server.handleNormalResponse(clientAddr, header, questions, answers)
		}
	} else {
		err = server.handleNormalResponse(clientAddr, header, questions, answers)
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
	result := make([]byte, config.PkgLimitRFC1035)
	remaining := result[12:]
	reqHeader.QuestionCount = 1
	var ansCount uint16 = 0

	for _, question := range questions {
		if remaining, err = question.WriteTo(remaining); err != nil {
			return err
		}
	}

	for _, question := range questions {
		var query []byte
		if query, err = buildQuery(reqHeader, question); err != nil {
			return err
		}
		var response []byte
		if response, err = server.forward(query); err != nil {
			return err
		}
		//response contains answer
		if len(response) > len(query) {
			answer := response[len(query):]
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

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

var MaxBuffer = 1024

type Handler struct {
	Server *Peer
}

func (handler *Handler) HandleRequest(conn net.Conn) {
	buffer := make([]byte, MaxBuffer)

	_, err := conn.Read(buffer)

	request := string(buffer)

	if err != nil {
		log.Printf("Error reading message from connection [%s]: %s\n", conn.RemoteAddr().String(), err.Error())
	} else {
		log.Printf("Request from %s: %s\n", conn.RemoteAddr().String(), request)
	}

	response := []byte(`request type not recognized`)
	switch request {
	case "PING":
		response = handler.HealthCheck()
	case "REGISTER":
		response = handler.RegisterHandler(conn)
	case "ECHO":
		response = handler.EchoHandler(request)
	default:
		// do nothing
	}

	conn.Write(response)

	conn.Close()
}

// Ping will "ping" a peer and determine if the peer is alive
func (handler *Handler) Ping(peer string) bool {
	response := false

	conn, err := net.Dial("tcp", peer)
	if err != nil {
		log.Printf("Error dialing %s: %s", peer, err.Error())
	} else {
		conn.Write([]byte(`PING`))
		status, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("Error recieving PING from %s: %s\n", peer, err.Error())
			handler.Server.RemovePeer(peer)
		} else {
			log.Printf("Successful response from %s: %s\n", peer, status)
			response = true
		}
	}

	return response
}

func (handler *Handler) HealthCheck() []byte {
	return []byte(`PONG`)
}

func (handler *Handler) RegisterHandler(conn net.Conn) []byte {
	addr := conn.RemoteAddr()
	io.WriteString(handler.Server.Hash, fmt.Sprintf("%s:%d", addr.String(), 8888))
	peerID := fmt.Sprintf("%x", handler.Server.Hash.Sum(nil))

	err := handler.Server.AddPeer(peerID, addr)
	if err != nil {
		return []byte(fmt.Sprintf(`[ERROR] Unable to add to this peer: %s`, err.Error()))
	} else {
		return []byte(`Successfully added Peer!`)
	}
}

func (handler *Handler) EchoHandler(message string) []byte {
	return []byte(message)
}

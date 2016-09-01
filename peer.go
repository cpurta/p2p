package main

import (
	"log"
	"math"
	"net"
	"net/http"
	"sync"
)

var (
	const BufferLength = 2048
)

type Peer struct {
	Debug    bool
	MaxPeers int
	Port     int
	Host     string
	PeerID   string
	PeerLock *sync.Mutex
	Peers    map[string]int
	Shutdown bool
	Handlers map[string]http.HandleFunc
}

func NewPeer(maxPeers int) *Peer {
	p := &Peer{
		Debug:    false,
		MaxPeers: maxPeers,
		Port:     80,
		Peers:    make(map[string]int),
		Handlers: make(map[string]http.HandleFunc),
	}

	if maxPeers == 0 {
		p.MaxPeers = math.MaxInt64
	}

	p.init()

	return p
}

func (p *Peer) init() {
	// we need to get the ip address of our machine
	ifaces, err := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			p.Host = ip.String()
			p.PeerID = p.Host + ":" + p.Port
		}
	}
}

func (p *Peer) Start() {
	l, err := net.Listen("tcp", "127.0.0.1:"+p.Port)
	if err != nil {
		log.Fatalln("Error creating listener on 127.0.0.1:"+p.Port+": ", err.Error())
	}

	defer l.Close()

	log.Printf("Peer [%s] listening on %s:%d", p.PeerID, p.Host, p.Port)

	for !p.Shutdown {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting request from %s: %s", conn.RemoteAddr().String(), err.Error())
		}

		go handleRequest(conn)
	}

	log.Printf("Peer [%s] is shutting down.\n", p.PeerID)
}

func (p *Peer) SetDebug(debug bool) {
	p.Debug = debug
}

func (p *Peer) Shutdown() {
	p.Shutdown = true
}

func handleRequest(conn net.Conn) {
	buffer := make([]byte, BufferLength)

	reqLen, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading message from connection [%s]: %s\n", conn.RemoteAddr().String(), err.Error())
	}

	conn.Write([]byte("Message Recieved"))

	conn.Close()
}

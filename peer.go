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

	go p.listenForShutdown()

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

func (p *Peer) AddPeer(peerID string) {
	p.PeerLock.Lock()
	defer p.PeerLock.Unlock()
	if _, ok := p.Peers[peerID]; !ok {
		p.Peers[peerID] = 1
	} else {
		log.Printf("Peer [%s] already in list of know peers.\n", peerID)
	}
}

func (p *Peer) RemovePeer(peerID string) error {
	p.PeerLock.Lock()
	defer p.PeerLock.Unlock()
	if _, ok := p.Peers[peerID]; ok {
		delete(p.Peers[peerID])
	} else {
		log.Printf("Unable to remove peer [%s] from list of know peers", peerID)
	}
}

func (p *Peer) CheckLivePeers() {
	toDelete := make([]string, 0)

	for host, port := range p.Peers {
		peerID := fmt.Sprintf("%s:%d", host, port)

		conn, err := net.Dial("tcp", host + ":" + string(port))
		if err != nil {
			log.Printf("Error dialing %s: %s", peerID, err.Error())
		} else {
			conn.Write(`PING`)
			status, err := bufio.NewReader(conn).ReadString("\n")
			if err != nil {
				log.Printf("Error recieving PING from %s\n", peerID, err.Error())
				p.RemovePeer(peerID)
			} else {
				log.Printf("Successful response from %s\n", peerID)
			}
		}
	}
}

func handleRequest(conn net.Conn) {
	buffer := make([]byte, BufferLength)

	reqLen, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading message from connection [%s]: %s\n", conn.RemoteAddr().String(), err.Error())
	}

	conn.Write([]byte("Message Recieved\n"))

	conn.Close()
}

func (p *Peer) listenForShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, unix.SIGTERM)
	go func() {
		msg := "Starting shutdown."
		for s := range sig {
			log.Printf("Received signal: %v. %s\n", s, msg)

			msg = "Shutdown in progress."
			p.Shutdown()
		}
	}()
}

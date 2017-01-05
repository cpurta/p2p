package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type Peer struct {
	Hash     hash.Hash
	MaxPeers int
	Port     string
	Host     string
	PeerID   string
	PeerLock *sync.Mutex
	Peers    map[string]net.Addr
	shutdown bool
	Handler  *Handler
}

func NewPeer(maxPeers int) *Peer {
	p := &Peer{
		Hash:     sha1.New(),
		MaxPeers: peers,
		Port:     "8888",
		Peers:    make(map[string]net.Addr),
		shutdown: false,
	}

	if maxPeers == 0 {
		p.MaxPeers = math.MaxInt64
	}

	p.Handler = &Handler{Server: p}

	p.init()

	return p
}

func (p *Peer) init() {
	// we need to get the ip address of our machine
	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
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
			io.WriteString(p.Hash, fmt.Sprintf("%s:%s", p.Host, p.Port))
			p.PeerID = fmt.Sprintf("%x", p.Hash.Sum(nil))
		}
	}
}

func (p *Peer) Start() {
	l, err := net.Listen("tcp", "127.0.0.1:"+p.Port)
	if err != nil {
		log.Fatalln("Error creating listener on 127.0.0.1:"+p.Port+": ", err.Error())
	}

	defer l.Close()

	log.Printf("Peer [%s] listening on 127.0.0.1:8888 and has a registered IP of %s:%s", p.PeerID, p.Host, p.Port)

	go p.listenForShutdown()

	go p.CheckLivePeers()

	for !p.shutdown {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting request from %s: %s", conn.RemoteAddr().String(), err.Error())
		}

		go p.Handler.HandleRequest(conn)
	}

	log.Printf("Peer [%s] is shutting down.\n", p.PeerID)
}

func (p *Peer) Shutdown() {
	p.shutdown = true
}

func (p *Peer) AddPeer(peerID string, addr net.Addr) error {
	p.PeerLock.Lock()
	defer p.PeerLock.Unlock()
	if len(p.Peers) == p.MaxPeers {
		return errors.New("Max peer limit has been reached for peer: " + p.PeerID)
	}

	if _, ok := p.Peers[peerID]; !ok {
		p.Peers[peerID] = addr
		log.Println("Added", peerID, "to list of known peers")
		return nil
	}

	log.Printf("Peer [%s] already in list of know peers.\n", peerID)
	return nil
}

func (p *Peer) RemovePeer(peerID string) error {
	p.PeerLock.Lock()
	defer p.PeerLock.Unlock()
	if _, ok := p.Peers[peerID]; ok {
		delete(p.Peers, peerID)
		log.Println("Successfully removed peer:", peerID)
		return nil
	}
	log.Printf("Unable to remove peer [%s] from list of know peers", peerID)

	return fmt.Errorf("Peer [%s] not found in known peers", peerID)
}

func (p *Peer) CheckLivePeers() {
	ticker := time.NewTicker(time.Minute * 5)

	defer ticker.Stop()
	for range ticker.C {
		if p.shutdown {
			return
		}
		log.Println("Checking if known peers are alive...")

		for _, addr := range p.Peers {
			peerID := fmt.Sprintf("%s:%s", addr.String(), p.Port)

			if !p.Handler.Ping(peerID) {
				log.Println("Unable to ping:", peerID)
				log.Println("Removing peer:", peerID)
				p.RemovePeer(peerID)
			}
		}
	}
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
			return
		}
	}()
}

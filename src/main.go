package main

import "flag"

var (
	peers int
)

func main() {
	initFlags()
	flag.Parse()

	peer := NewPeer(peers)

	peer.Start()
}

func initFlags() {
	flag.IntVar(&peers, "peers", 1000, "The number of peers that will be allowed to connect to the running server, setting peers to 0 will allow an unlimited amount of peers to connect to the server.")
}

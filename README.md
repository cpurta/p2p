# p2p

A simple peer to peer network that communicates over tcp. This is currently being
expanded upon to demonstrate various distributed data structures and mostly a pet
projects with peer-to-peer techniques.


## Getting started

Currently there are no dependencies that this project is reliant on, so you can build
the p2p binary directly as such using `wgo` or `go`:

```
$ git clone https://github.com/cpurta/p2p.git
$ cd p2p/src
$ wgo build -o p2p
$ mv p2p ../bin
```

## Running a single Peer node

To run a node for this peer2peer network just run the binary:

```
$ ./bin/p2p
```

You can test that the node is running and accepting requests by sending a TCP request
to the server running locally:

```
$ echo -n "PING" | nc 127.0.0.1 8888
PONG
```

## LICENSE

MIT

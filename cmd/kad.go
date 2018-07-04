package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	dht "gx/ipfs/QmNg6M98bwS97SL9ArvrRxKujFps3eV6XvmKgduiYga8Bn/go-libp2p-kad-dht"
	net "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	multiaddr "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	libp2p "gx/ipfs/QmZ86eLPtXkQ1Dfa992Q8NpXArUoWWh3y728JDcWvzRrvC/go-libp2p"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	host "gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

func dieIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func addrForPort(p string) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", p))
}

func streamHandlerFor(name string, kad *dht.IpfsDHT) func(s net.Stream) {
	fn := func(s net.Stream) {
		conn := s.Conn()

		log.Printf("Opened new stream %s: %v", name, s.Protocol())
		log.Printf("  Local Addr:  %s", conn.LocalMultiaddr().String())
		log.Printf("  Remote Addr: %s", conn.RemoteMultiaddr().String())
		log.Printf("  Remote Peer: %s", conn.RemotePeer().Pretty())
	}

	return fn
}

var protocols = [4]string{"/multistream/1.0.0", "/ipfs/id/1.0.0", "/ipfs/kad/1.0.0", "/ipfs/dht"}

func generateHost(ctx context.Context, port int64) (host.Host, *dht.IpfsDHT) {
	randBytes := rand.New(rand.NewSource(port))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randBytes)
	dieIfError(err)

	hostAddr, err := addrForPort(fmt.Sprintf("%d", port))
	dieIfError(err)

	opts := []libp2p.Option{
		libp2p.ListenAddrs(hostAddr),
		libp2p.Identity(prvKey),
	}

	host, err := libp2p.New(ctx, opts...)
	dieIfError(err)

	kadDHT, err := dht.New(ctx, host)
	dieIfError(err)

	for _, proto := range protocols {
		host.SetStreamHandler(protocol.ID(proto), streamHandlerFor(proto, kadDHT))
	}

	fmt.Println("Generated host: ", host.ID().Pretty())

	return host, kadDHT
}

func addPeers(h host.Host, peerStr string) {
	if len(peerStr) == 0 {
		return
	}

	portStrs := strings.Split(peerStr, ",")
	for i := 0; i < len(portStrs); i++ {
		addr, err := addrForPort(portStrs[i])
		dieIfError(err)
		pid := "QmcxsSTeHBEfaWBb2QKe5UZWK8ezWJkxJfmcb5rQV374M6" //peer.ID(fmt.Sprintf("QmcxsSTeHBEfaWBb2QKe5UZWK8ezWJkxJfmcb5rQV374M6", portStrs[i]))
		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			fmt.Printf("Decode pid %v\n", err)
		}

		h.Peerstore().AddAddr(peerid, addr, 24*time.Hour)
		_, err = h.NewStream(context.Background(), peerid, "/multistream/1.0.0", "/ipfs/id/1.0.0", "/ipfs/kad/1.0.0", "/ipfs/dht")
		fmt.Printf("Error on new stream: %v\n", err)
	}
}

func main() {
	fmt.Println("Kademlia DHT test")

	port := flag.Int64("port", 0, "Port to listen on")
	peers := flag.String("peers", "", "Initial peers")
	flag.Parse()

	ctx := context.Background()
	srvHost, kad := generateHost(ctx, *port)
	_ = kad

	addPeers(srvHost, *peers)

	fmt.Printf("Listening on %v\n", srvHost.Addrs())
	fmt.Printf("Protocols supported: %v\n", srvHost.Mux().Protocols())

	<-make(chan struct{})

	// srcHost := generateHost(ctx, 3001)
	// fmt.Println(srcHost.ID().Pretty())

	// // dataStore := memstore.NewIntMemstore()

	// kadDHT, _ := dht.New(ctx, srcHost)

	// peers, _ := kadDHT.GetClosestPeers(ctx, "foo")

	// fmt.Println("Close peers ", peers)
}

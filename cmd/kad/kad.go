package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	utils "github.com/govi218/kademlia/utils"

	logging "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("main")

func addrForPort(p string) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", p))
}

func generateHost(ctx context.Context, port int64) (host.Host, *dht.IpfsDHT) {
	prvKey := utils.GeneratePrivateKey(port)

	hostAddr, err := addrForPort(fmt.Sprintf("%d", port))
	if err != nil {
		log.Fatal(err)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(hostAddr),
		libp2p.Identity(prvKey),
	}

	host, err := libp2p.New(ctx, opts...)
	if err != nil {
		log.Fatal(err)
	}

	kadDHT, err := dht.New(ctx, host, dhtopts.Validator(utils.NullValidator{}))
	if err != nil {
		log.Fatal(err)
	}

	hostID := host.ID()
	log.Infof("Host MultiAddress: %s/ipfs/%s (%s)", host.Addrs()[0].String(), hostID.Pretty(), hostID.String())

	return host, kadDHT
}

func addPeers(ctx context.Context, h host.Host, kad *dht.IpfsDHT, peersArg string) {
	if len(peersArg) == 0 {
		return
	}

	peerStrs := strings.Split(peersArg, ",")
	for i := 0; i < len(peerStrs); i++ {
		peerID, peerAddr := utils.MakePeer(peerStrs[i])

		h.Peerstore().AddAddr(peerID, peerAddr, peerstore.PermanentAddrTTL)
		kad.Update(ctx, peerID)
	}
}

func main() {
	logwriter.Configure(logwriter.Output(os.Stdout))
	log.Info("Kademlia DHT test")

	port := flag.Int64("port", 3001, "Port to listen on")
	peers := flag.String("peers", "", "Initial peers")
	flag.Parse()

	ctx := context.Background()
	srvHost, kad := generateHost(ctx, *port)
	addPeers(ctx, srvHost, kad, *peers)

	log.Infof("Listening on %v (Protocols: %v)", srvHost.Addrs(), srvHost.Mux().Protocols())
	<-make(chan struct{})
}

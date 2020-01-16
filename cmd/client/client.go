package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/vincentchu/kademlia/utils"

	logging "github.com/ipfs/go-log"
	host "github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"

	libp2p "github.com/libp2p/go-libp2p"
)

var log = logging.Logger("client")

func makeHost(ctx context.Context) host.Host {
	prvKey := utils.GeneratePrivateKey(999)

	h, err := libp2p.New(ctx, libp2p.Identity(prvKey))
	if err != nil {
		log.Fatalf("Err on creating host: %v", err)
	}

	return h
}

func parseCmd(tokens []string) (string, string, string) {
	switch len(tokens) {
	case 2:
		return tokens[0], tokens[1], ""
	case 3:
		return tokens[0], tokens[1], tokens[2]
	default:
		log.Fatalf("Improper command format: %v", tokens)
		return "", "", ""
	}
}

func main() {
	logwriter.Configure(logwriter.Output(os.Stdout), logwriter.LevelInfo)
	dest := flag.String("dest", "", "Destination to connect to")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := makeHost(ctx)
	destID, destAddr := utils.MakePeer(*dest)

	h.Peerstore().AddAddr(destID, destAddr, 24*time.Hour)
	kad, err := dht.New(ctx, h, dhtopts.Client(true), dhtopts.Validator(utils.NullValidator{}))
	if err != nil {
		log.Fatalf("Error creating DHT: %v", err)
	}
	kad.Update(ctx, destID)

	cmd, key, val := parseCmd(flag.Args())
	switch cmd {
	case "put":
		log.Infof("PUT %s => %s", key, val)
		err = kad.PutValue(ctx, key, []byte(val))
		if err != nil {
			log.Fatalf("Error on PUT: %v", err)
		}

	case "get":
		log.Infof("GET %s", key)
		fetchedBytes, err := kad.GetValue(ctx, key, dht.Quorum(1))
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		log.Infof("RESULT: %s", string(fetchedBytes))

	default:
		log.Fatalf("Command %s unrecognized", cmd)
	}
}

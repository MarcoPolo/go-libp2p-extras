package libp2pextras_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	libp2pextras "github.com/marcopolo/go-libp2p-extras"
)

func TestLogsMissingDeadline(t *testing.T) {
	h, err := libp2p.New()
	if err != nil {
		t.Fatal(err)
	}
	h2, err := libp2p.New()
	if err != nil {
		t.Fatal(err)
	}

	// Wrap the host with a logger that will log when deadlines are not set.
	logger := log.New(os.Stderr, "libp2p-extras: ", log.LstdFlags)
	called := make(chan struct{}, 1)
	h = libp2pextras.WrapWithLogMissingSetDeadlines(h, func(s string) {
		select {
		case called <- struct{}{}:
		default:
		}
		logger.Println(s)
	})

	h.Connect(context.Background(), peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	})

	h.SetStreamHandler("/ipfs/ping/1.0.0", func(s network.Stream) {
		defer s.Close()
		b := make([]byte, 32)
		s.Read(b)
		s.Write(b)
	})

	<-ping.Ping(context.Background(), h2, h.ID())

	<-called
}

package libp2pextras

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type wrappedHost struct {
	host.Host
	logger func(string)
}

var _ host.Host = (*wrappedHost)(nil)

func WrapWithLogMissingSetDeadlines(host host.Host, logger func(string)) *wrappedHost {
	return &wrappedHost{host, logger}
}

type wrappedStream struct {
	network.Stream
	proto                  protocol.ID
	logger                 func(string)
	setDeadlineCalled      atomic.Bool
	setWriteDeadlineCalled atomic.Bool
	setReadDeadlineCalled  atomic.Bool
	closed                 atomic.Bool
}

func (s *wrappedStream) deadlineProperlyCalled() bool {
	return s.setDeadlineCalled.Load() || (s.setWriteDeadlineCalled.Load() && s.setReadDeadlineCalled.Load())
}

func (s *wrappedStream) startTracking() func() {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				if !s.deadlineProperlyCalled() && !s.closed.Load() {
					s.logger(fmt.Sprintf("stream for %s probably never set a deadline!", s.proto))
				}
				return
			case <-time.After(5 * time.Second):
				if !s.deadlineProperlyCalled() && !s.closed.Load() {
					s.logger(fmt.Sprintf("stream for %s probably never set a deadline!", s.proto))
				}
			}
		}
	}()

	return func() {
		close(done)
	}
}

func (s *wrappedStream) SetDeadline(t time.Time) error {
	s.setDeadlineCalled.Store(true)
	return s.Stream.SetDeadline(t)
}

func (s *wrappedStream) SetReadDeadline(t time.Time) error {
	s.setReadDeadlineCalled.Store(true)
	return s.Stream.SetReadDeadline(t)
}

func (s *wrappedStream) SetWriteDeadline(t time.Time) error {
	s.setWriteDeadlineCalled.Store(true)
	return s.Stream.SetWriteDeadline(t)
}

func (s *wrappedStream) Write(b []byte) (int, error) {
	if !(s.setDeadlineCalled.Load() || s.setWriteDeadlineCalled.Load()) {
		s.logger(fmt.Sprintf("stream for %s probably never set a write deadline!", s.proto))
	}
	return s.Stream.Write(b)
}

func (s *wrappedStream) Read(b []byte) (int, error) {
	if !(s.setDeadlineCalled.Load() || s.setReadDeadlineCalled.Load()) {
		s.logger(fmt.Sprintf("stream for %s probably never set a read deadline!", s.proto))
	}
	return s.Stream.Read(b)
}

func (s *wrappedStream) Close() error {
	s.closed.Store(true)
	return s.Stream.Close()
}
func (s *wrappedStream) Reset() error {
	s.closed.Store(true)
	return s.Stream.Reset()
}

func (s *wrappedStream) CloseRead() error {
	s.closed.Store(true)
	return s.Stream.CloseRead()
}

func (s *wrappedStream) CloseWrite() error {
	s.closed.Store(true)
	return s.Stream.CloseWrite()
}

func (h *wrappedHost) SetStreamHandler(protocol protocol.ID, handler network.StreamHandler) {
	h.Host.SetStreamHandler(protocol, func(s network.Stream) {
		w := &wrappedStream{
			Stream: s,
			logger: h.logger,
			proto:  protocol,
		}
		cancel := w.startTracking()
		defer cancel()
		handler(w)
	})
}

func (h *wrappedHost) SetStreamHandlerMatch(protocol protocol.ID, matcher func(protocol.ID) bool, handler network.StreamHandler) {
	h.Host.SetStreamHandlerMatch(protocol, matcher, func(s network.Stream) {
		w := &wrappedStream{
			Stream: s,
			logger: h.logger,
			proto:  protocol,
		}
		cancel := w.startTracking()
		defer cancel()
		handler(w)
	})
}

func (h *wrappedHost) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	s, err := h.Host.NewStream(ctx, p, pids...)
	if err != nil {
		return nil, err
	}
	w := &wrappedStream{
		Stream: s,
		logger: h.logger,
		proto:  pids[0],
	}
	cancel := w.startTracking()
	defer cancel()
	return w, nil
}

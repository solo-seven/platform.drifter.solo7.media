package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/solo-seven/platform.drifter.solo7.media/generated/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GameStateServer struct {
	proto.UnimplementedMUDServiceServer
	grpcServer *grpc.Server
	listener   net.Listener

	mu      sync.Mutex
	running bool
	wg      sync.WaitGroup
}

func NewGameStateServer() *GameStateServer {
	s := grpc.NewServer()
	server := &GameStateServer{
		grpcServer: s,
	}
	proto.RegisterMUDServiceServer(s, server)
	reflection.Register(s)
	return server
}

func (g *GameStateServer) Start(addr string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.running {
		return fmt.Errorf("server already running")
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	g.listener = lis
	g.running = true
	go func() {
		err = g.grpcServer.Serve(g.listener)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (g *GameStateServer) Stop() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.running {
		return fmt.Errorf("server not running")
	}
	g.grpcServer.Stop()
	g.running = false
	return nil
}

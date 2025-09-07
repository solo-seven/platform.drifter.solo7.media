package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/solo-seven/platform.drifter.solo7.media/generated/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
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

func (g *GameStateServer) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (g *GameStateServer) ServerList(ctx context.Context, in *proto.ServerListRequest) (*proto.ServerListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ServerList not implemented")
}

func (g *GameStateServer) AvatarConnect(ctx context.Context, in *proto.AvatarConnectRequest) (*proto.AvatarConnectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AvatarConnect not implemented")
}

func (g *GameStateServer) AvatarCommand(ctx context.Context, in *proto.AvatarCommandRequest) (*proto.AvatarCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AvatarCommand not implemented")
}

func (g *GameStateServer) AvatarMessages(ctx context.Context, in *proto.AvatarMessageRequest) (*proto.AvatarMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AvatarMessages not implemented")
}

func (g *GameStateServer) AdminCommand(ctx context.Context, in *proto.AdminCommandRequest) (*proto.AdminCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminCommand not implemented")
}

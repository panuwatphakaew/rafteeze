package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	pb "github.com/panuwatphakaew/rafteeze/proto"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Transport struct {
	pb.UnimplementedRaftTransportServer
	nodeID      uint64
	listenAddr  string
	peers       map[uint64]string
	clients     map[uint64]pb.RaftTransportClient
	clientConns map[uint64]*grpc.ClientConn
	messageRecv chan raftpb.Message
	server      *grpc.Server
	mu          sync.RWMutex
}

func NewTransport(nodeID uint64, listenAddr string, peers map[uint64]string) (*Transport, error) {
	t := &Transport{
		nodeID:      nodeID,
		listenAddr:  listenAddr,
		peers:       peers,
		clients:     make(map[uint64]pb.RaftTransportClient),
		clientConns: make(map[uint64]*grpc.ClientConn),
		messageRecv: make(chan raftpb.Message, 100),
	}

	t.server = grpc.NewServer()
	pb.RegisterRaftTransportServer(t.server, t)

	return t, nil
}

func (t *Transport) Send(msg raftpb.Message) error {
	client, err := t.getOrCreateClient(msg.To)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.SendMessage(ctx, &msg)
	return err
}

func (t *Transport) getOrCreateClient(peerID uint64) (pb.RaftTransportClient, error) {
	t.mu.RLock()
	client, ok := t.clients[peerID]
	t.mu.RUnlock()

	if ok {
		return client, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if client, ok := t.clients[peerID]; ok {
		return client, nil
	}

	peerAddr, ok := t.peers[peerID]
	if !ok {
		return nil, fmt.Errorf("unknown peer: %d", peerID)
	}

	conn, err := grpc.NewClient(
		peerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %d: %w", peerID, err)
	}

	client = pb.NewRaftTransportClient(conn)
	t.clients[peerID] = client
	t.clientConns[peerID] = conn

	return client, nil
}

func (t *Transport) SendMessage(ctx context.Context, msg *raftpb.Message) (*pb.SendMessageResponse, error) {
	select {
	case t.messageRecv <- *msg:
		return &pb.SendMessageResponse{Success: true}, nil
	default:
		slog.Warn("message receive buffer full, dropping message")
		return &pb.SendMessageResponse{
			Success: false,
			Error:   "receive buffer full",
		}, status.Error(codes.ResourceExhausted, "receive buffer full")
	}
}

func (t *Transport) ReceiveMessage() <-chan raftpb.Message {
	return t.messageRecv
}

func (t *Transport) Start() error {
	lis, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", t.listenAddr, err)
	}

	slog.Info("gRPC transport server starting", slog.String("address", t.listenAddr))

	if err := t.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (t *Transport) Stop() error {
	slog.Info("stopping gRPC transport server")

	t.server.GracefulStop()

	t.mu.Lock()
	defer t.mu.Unlock()

	for peerID, conn := range t.clientConns {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close connection to peer",
				slog.Uint64("peer", peerID),
				slog.Any("error", err))
		}
	}

	return nil
}

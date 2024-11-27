package server

import (
	"context"
	"log"
	"net"
	pb "papers/pkg/grpc/pb/balanceService"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Config struct {
	Port string `yaml:"port"`
}

type server struct {
	pb.UnimplementedBalanceManagementServer
	db IDBManager
}

type Service struct {
	Server   *grpc.Server
	Listener *net.Listener
	cfg      *Config
	db       IDBManager
}

type IDBManager interface {
	GetUserBalance(userID uuid.UUID) (float32, error)
	AddBalance(userID uuid.UUID, cash float32) (float32, error)
	TakeFromBalance(userID uuid.UUID, cash float32) (float32, error)
}

// Создание сервера сервиса аутентификации
func New(cfg *Config, db IDBManager) (*Service, error) {
	log.Println(cfg.Port)
	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	pb.RegisterBalanceManagementServer(s, &server{db: db})
	log.Printf("Auth server listening at %v\n", lis.Addr())
	return &Service{Server: s, Listener: &lis, cfg: cfg, db: db}, nil
}

func (s *server) GetBalance(_ context.Context, in *pb.User) (*pb.Balance, error) {
	balance, err := s.db.GetUserBalance(uuid.UUID(in.Id))
	if err != nil {
		return nil, err
	}
	return &pb.Balance{Cash: balance}, nil
}

func (s *server) AddBalance(_ context.Context, in *pb.Money) (*pb.Balance, error) {
	balance, err := s.db.AddBalance(uuid.UUID(in.Id), in.Cash)
	if err != nil {
		return nil, err
	}
	return &pb.Balance{Cash: balance}, nil
}

func (s *server) TakeFromBalance(_ context.Context, in *pb.Money) (*pb.Balance, error) {
	balance, err := s.db.TakeFromBalance(uuid.UUID(in.Id), in.Cash)
	if err != nil {
		return nil, err
	}
	return &pb.Balance{Cash: balance}, nil
}

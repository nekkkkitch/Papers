package server

import (
	"context"
	"encoding/json"
	"log"
	"net"
	pb "papers/pkg/grpc/pb/papersService"
	"papers/pkg/models"

	"google.golang.org/grpc"
)

type Config struct {
	Port string `yaml:"port"`
}

type server struct {
	pb.UnimplementedPapersManagementServer
	rds IRDS
	db  IDB
}

type Service struct {
	PpsServer *grpc.Server
	Listener  *net.Listener
	cfg       *Config
	rds       IRDS
	db        IDB
}

type IRDS interface {
	GetAvailablePapers() ([]models.Paper, error)
}

type IDB interface {
}

func New(cfg *Config, db IDB, redis IRDS) (*Service, error) {
	log.Println(cfg.Port)
	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	pb.RegisterPapersManagementServer(s, &server{rds: redis, db: db})
	log.Printf("Auth server listening at %v\n", lis.Addr())
	return &Service{PpsServer: s, Listener: &lis, cfg: cfg, db: db, rds: redis}, nil
}

func (s *server) GetAvailablePapers(_ context.Context, _ *pb.Request) (*pb.AvailablePapers, error) {
	papers, err := s.rds.GetAvailablePapers()
	if err != nil {
		log.Println("Failed to get papers:", err)
		return nil, err
	}
	jsoned, err := json.Marshal(papers)
	if err != nil {
		log.Println("Failed to marshap papers:", err)
	}
	return &pb.AvailablePapers{Papers: jsoned}, nil
}

func (s *server) BuyPaper(_ context.Context, paper *pb.Paper) (*pb.Status, error) {
	return nil, nil
}

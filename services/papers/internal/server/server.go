package server

import (
	"context"
	"encoding/json"
	"log"
	"net"
	problems "papers/internal/pkg/customErrors"
	"papers/internal/pkg/models"
	pb "papers/internal/pkg/papersService"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Config struct {
	Port string `yaml:"port"`
}

type server struct {
	pb.UnimplementedPapersManagementServer
	redis IRDS
	db    IDB
}

type Service struct {
	PpsServer *grpc.Server
	Listener  *net.Listener
	cfg       *Config
	redis     IRDS
	db        IDB
}

type IRDS interface {
	GetAvailablePapers() ([]models.Paper, error)
	GetPaperPrice(name string) (float32, error)
	GetUserPapers(userId uuid.UUID) ([]models.Paper, error)
	UpdateUserPapers(userId uuid.UUID, papers []models.Paper) error
}

type IDB interface {
	GetUserPapers(userId uuid.UUID) ([]models.Paper, error)
	GetUserBalance(userId uuid.UUID) (float32, error)
	ChangeBalance(userId uuid.UUID, change float32) error
	ChangePaperAmount(userId uuid.UUID, name string, amount int) error
	GetPaperAmount(userId uuid.UUID, name string) (int, error)
	AddPaper(userId uuid.UUID, name string, amount int) error
}

func New(cfg *Config, db IDB, redis IRDS) (*Service, error) {
	log.Println(cfg.Port)
	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	pb.RegisterPapersManagementServer(s, &server{redis: redis, db: db})
	log.Printf("Auth server listening at %v\n", lis.Addr())
	return &Service{PpsServer: s, Listener: &lis, cfg: cfg, db: db, redis: redis}, nil
}

func (s *server) GetAvailablePapers(_ context.Context, _ *pb.Request) (*pb.AvailablePapers, error) {
	papers, err := s.redis.GetAvailablePapers()
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

func (s *server) BuyPaper(_ context.Context, paperreq *pb.PaperRequest) (*pb.Status, error) {
	userId, err := uuid.FromBytes(paperreq.UserId)
	if err != nil {
		log.Println("Cant get uuid from given bytes:", err)
		return nil, err
	}
	userBalance, err := s.db.GetUserBalance(userId)
	if err != nil {
		log.Println("Cant get user balance:", err)
		return nil, err
	}
	paperPrice, err := s.redis.GetPaperPrice(paperreq.PaperName)
	if err != nil {
		log.Println("Cant get requested paper price:", err)
		return nil, err
	}
	if paperPrice == 0 {
		log.Printf("Paper %v does not exists\n", paperreq.PaperName)
		return &pb.Status{Response: problems.NoPaper}, nil
	}
	cost := paperPrice * float32(paperreq.PaperAmount)
	if cost > userBalance {
		return &pb.Status{Response: problems.LowBalance}, nil
	}
	err = s.db.ChangeBalance(userId, -cost)
	if err != nil {
		log.Println("Cant change user balance:", err)
		return nil, err
	}
	am, err := s.db.GetPaperAmount(userId, paperreq.PaperName)
	if err != nil {
		log.Println("Cant get user amount of papers:", err)
		s.db.ChangeBalance(userId, cost)
		return nil, err
	}
	if am == -1 {
		err = s.db.AddPaper(userId, paperreq.PaperName, int(paperreq.PaperAmount))
		if err != nil {
			log.Println("Cant add user papers:", err)
			s.db.ChangeBalance(userId, cost)
			return nil, err
		}
	} else {
		err = s.db.ChangePaperAmount(userId, paperreq.PaperName, int(paperreq.PaperAmount))
	}
	if err != nil {
		log.Println("Cant change user amount of papers:", err)
		s.db.ChangeBalance(userId, cost)
		return nil, err
	}
	papers, err := s.db.GetUserPapers(userId)
	if err != nil {
		log.Println("Cant get user papers:", err)
		return nil, err
	}
	err = s.redis.UpdateUserPapers(userId, papers)
	if err != nil {
		log.Println("Cant update user papers:", err)
		return nil, err
	}
	return nil, nil
}

func (s *server) SellPaper(_ context.Context, paperreq *pb.PaperRequest) (*pb.Status, error) {
	userId, err := uuid.FromBytes(paperreq.UserId)
	if err != nil {
		log.Println("Cant get uuid from given bytes:", err)
		return nil, err
	}
	availableAmount, err := s.db.GetPaperAmount(userId, paperreq.PaperName)
	if err != nil {
		log.Println("Cant get amount of papers:", err)
		return nil, err
	}
	if paperreq.PaperAmount > int32(availableAmount) {
		return &pb.Status{Response: problems.LowPaper}, nil
	}
	paperPrice, err := s.redis.GetPaperPrice(paperreq.PaperName)
	if err != nil {
		log.Println("Cant get paper price:", err)
		return nil, err
	}
	if paperPrice == 0 {
		return &pb.Status{Response: problems.NoPaper}, nil
	}
	err = s.db.ChangePaperAmount(userId, paperreq.PaperName, int(-paperreq.PaperAmount))
	if err != nil {
		log.Println("Cant change user amount of paper:", err)
		return nil, err
	}
	err = s.db.ChangeBalance(userId, float32(paperreq.PaperAmount)*paperPrice)
	if err != nil {
		log.Println("Cant change user balance:", err)
		s.db.ChangePaperAmount(userId, paperreq.PaperName, int(paperreq.PaperAmount))
		return nil, err
	}
	papers, err := s.db.GetUserPapers(userId)
	if err != nil {
		log.Println("Cant get user papers:", err)
		return nil, err
	}
	err = s.redis.UpdateUserPapers(userId, papers)
	if err != nil {
		log.Println("Cant update user papers:", err)
		return nil, err
	}
	return nil, nil
}

func (s *server) GetUserPapers(_ context.Context, user *pb.User) (*pb.AvailablePapers, error) {
	userId, err := uuid.FromBytes(user.Id)
	if err != nil {
		log.Println("Cant get uuid from given bytes:", err)
		return nil, err
	}
	papers, err := s.redis.GetUserPapers(userId)
	if err != nil {
		log.Println("Cant get user papers:", err)
		return nil, err
	}
	for i := range papers {
		papers[i].Price, err = s.redis.GetPaperPrice(papers[i].Name)
		if err != nil {
			log.Println("Cant get paper price:", err)
			return nil, err
		}
	}
	jsonedPapers, err := json.Marshal(papers)
	if err != nil {
		log.Println("Cant marshal user papers:", err)
		return nil, err
	}
	return &pb.AvailablePapers{Papers: jsonedPapers}, nil
}

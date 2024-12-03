package service

import (
	"log"
	"math/rand/v2"
	"papers/pkg/models"
	"time"
)

type Server struct {
	Papers []models.Paper
	db     IDB
	rds    IRDS
}

type IDB interface {
	GetListOfPapers() ([]models.Paper, error)
	PutValue(name string, value float32) error
}

type IRDS interface {
	UpdateStock(name string, value float32) error
}

var BasePapers = []models.Paper{{Name: "BasePaper", Price: 1}}

func New(db IDB, rds IRDS) (*Server, error) {
	server := &Server{db: db, rds: rds}
	var err error
	server.Papers, err = server.db.GetListOfPapers()
	if err != nil {
		log.Println("Cant get papers:", err)
		return nil, err
	}
	if len(server.Papers) == 0 {
		copy(server.Papers, BasePapers)
	}
	for _, value := range server.Papers {
		server.rds.UpdateStock(value.Name, value.Price)
	}
	return server, nil
}

func (s *Server) StartFun() {
	for {
		time.Sleep(5 * time.Second)
		choice := rand.IntN(len(s.Papers)) // the choice is yooooooours
		s.Papers[choice].Price *= (0.95 + float32(rand.IntN(10)+1)/100)
		log.Printf("Trying to update %v with value %v\n", s.Papers[choice].Name, s.Papers[choice].Price)
		err := s.db.PutValue(s.Papers[choice].Name, s.Papers[choice].Price)
		if err != nil {
			log.Printf("Failed to update %v in db with value %v: %v\n", s.Papers[choice].Name, s.Papers[choice].Price, err)
		}
		err = s.rds.UpdateStock(s.Papers[choice].Name, s.Papers[choice].Price)
		if err != nil {
			log.Printf("Failed to update %v in redis with value %v: %v\n", s.Papers[choice].Name, s.Papers[choice].Price, err)
		}
	}
}

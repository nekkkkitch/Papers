package service

import (
	"context"
	"flag"
	"log"

	papersService "gateway/internal/pkg/grpc/pb/papersService"
	"gateway/internal/pkg/models"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Host string `yaml:"host" env-prefix:"AUTHHOST"`
	Port string `yaml:"port" env-prefix:"AUTHPORT"`
}

type Client struct {
	client papersService.PapersManagementClient
	conn   *grpc.ClientConn
}

// Создание клиента для authService
func New(cfg *Config) (*Client, error) {
	flag.Parse()
	conn, err := grpc.NewClient(cfg.Host+cfg.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := papersService.NewPapersManagementClient(conn)
	log.Println("Connecting to papers on " + cfg.Host + cfg.Port)
	return &Client{client: c, conn: conn}, nil
}

func (c *Client) GetAvailablePapers() ([]byte, error) {
	resp, err := c.client.GetAvailablePapers(context.Background(), &papersService.Request{})
	if err != nil {
		log.Println("Failed to get available papers:", err)
		return nil, err
	}
	return resp.Papers, nil
}

func (c *Client) BuyPaper(userID *uuid.UUID, paper models.Paper) (string, error) {
	marshaledId, err := userID.MarshalBinary()
	if err != nil {
		log.Println("Failed to marshal user uuid")
		return "", err
	}
	resp, err := c.client.BuyPaper(context.Background(), &papersService.PaperRequest{UserId: marshaledId, PaperName: paper.Name, PaperAmount: paper.Amount})
	if err != nil {
		log.Println("Failed to buy paper:", err)
		return "", err
	}
	return resp.Response, nil
}

func (c *Client) SellPaper(userID *uuid.UUID, paper models.Paper) (string, error) {
	marshaledId, err := userID.MarshalBinary()
	if err != nil {
		log.Println("Failed to marshal user uuid")
		return "", err
	}
	resp, err := c.client.SellPaper(context.Background(), &papersService.PaperRequest{UserId: marshaledId, PaperName: paper.Name, PaperAmount: paper.Amount})
	if err != nil {
		log.Println("Failed to sell paper:", err)
		return "", err
	}
	return resp.Response, nil
}

func (c *Client) GetUserPapers(userID *uuid.UUID) ([]byte, error) {
	marshaledId, err := userID.MarshalBinary()
	if err != nil {
		log.Println("Failed to marshal user uuid")
		return nil, err
	}
	resp, err := c.client.GetUserPapers(context.Background(), &papersService.User{Id: marshaledId})
	if err != nil {
		log.Println("Failed to get available papers:", err)
		return nil, err
	}
	return resp.Papers, nil
}

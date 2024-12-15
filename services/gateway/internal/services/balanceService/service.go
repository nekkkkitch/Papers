package service

import (
	"context"
	"flag"
	"log"

	balanceService "papers/pkg/grpc/pb/balanceService"
	"papers/pkg/models"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Host string `yaml:"host" env-prefix:"AUTHHOST"`
	Port string `yaml:"port" env-prefix:"AUTHPORT"`
}

type Client struct {
	client balanceService.BalanceManagementClient
	conn   *grpc.ClientConn
}

func New(cfg *Config) (*Client, error) {
	flag.Parse()
	conn, err := grpc.NewClient(cfg.Host+cfg.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := balanceService.NewBalanceManagementClient(conn)
	log.Println("Connecting to balance service on " + cfg.Host + cfg.Port)
	return &Client{client: c, conn: conn}, nil
}

func (c *Client) GetBalance(userID *uuid.UUID) (float32, error) {
	marshaledId, err := userID.MarshalBinary()
	if err != nil {
		log.Println("Failed to marshal user uuid")
		return -1, err
	}
	balance, err := c.client.GetBalance(context.Background(), &balanceService.User{Id: marshaledId})
	if err != nil {
		log.Println(err)
		return -1, err
	}
	return balance.Cash, nil
}

func (c *Client) AddBalance(req *models.Money) (string, error) {
	marshaledId, err := req.ID.MarshalBinary()
	if err != nil {
		log.Println("Failed to marshal user uuid")
		return "", err
	}
	resp, err := c.client.AddBalance(context.Background(), &balanceService.Money{Id: marshaledId, Cash: req.Cash})
	if err != nil {
		log.Println(err)
		return "", err
	}
	return resp.Response, nil
}

func (c *Client) TakeBalance(req *models.Money) (string, error) {
	marshaledId, err := req.ID.MarshalBinary()
	if err != nil {
		log.Println("Failed to marshal user uuid")
		return "", err
	}
	resp, err := c.client.TakeBalance(context.Background(), &balanceService.Money{Id: marshaledId, Cash: req.Cash})
	if err != nil {
		log.Println(err)
		return "", err
	}
	return resp.Response, nil
}

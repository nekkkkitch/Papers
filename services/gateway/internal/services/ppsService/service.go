package service

import (
	"context"
	"flag"
	"log"

	papersService "papers/pkg/grpc/pb/papersService"

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
	log.Println("Connecting to aus on " + cfg.Host + cfg.Port)
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

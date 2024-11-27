package service

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"flag"
	"log"

	authService "papers/pkg/grpc/pb/authService"
	"papers/pkg/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Host string `yaml:"auth_host" env-prefix:"AUTHHOST"`
	Port string `yaml:"auth_port" env-prefix:"AUTHPORT"`
}

type Client struct {
	client authService.AuthentificationClient
	conn   *grpc.ClientConn
}

// Создание клиента для authService
func New(cfg *Config) (*Client, error) {
	flag.Parse()
	conn, err := grpc.NewClient(cfg.Host+cfg.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := authService.NewAuthentificationClient(conn)
	log.Println("Connecting to aus on " + cfg.Host + cfg.Port)
	return &Client{client: c, conn: conn}, nil
}

// Вызов функции регистрации
func (c *Client) Register(user models.User) (*models.AuthData, error) {
	authDataGed, err := c.client.Register(context.Background(), &authService.User{Login: user.Login, Password: user.Password})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &models.AuthData{AccessToken: authDataGed.AccessToken, RefreshToken: authDataGed.RefreshToken}, nil
}

// Вызов функции аутентификации пользователя
func (c *Client) Login(user models.User) (*models.AuthData, error) {
	authDataGed, err := c.client.Login(context.Background(), &authService.User{Login: user.Login, Password: user.Password})
	if err != nil {
		return nil, err
	}
	return &models.AuthData{AccessToken: authDataGed.AccessToken, RefreshToken: authDataGed.RefreshToken}, nil
}

// Вызов функции апдейта токенов
func (c *Client) UpdateTokens(tokens models.AuthData) (*models.AuthData, error) {
	authDataGed, err := c.client.UpdateTokens(context.Background(), &authService.AuthData{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken})
	if err != nil {
		return nil, err
	}
	return &models.AuthData{AccessToken: authDataGed.AccessToken, RefreshToken: authDataGed.RefreshToken}, nil
}

// Получени приватного ключа для jwt(см jwt.NewWithKey())
func (c *Client) GetPrivateKey() (*rsa.PrivateKey, error) {
	data, err := c.client.GetPrivateKey(context.Background(), &authService.KeyRequest{})
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKCS1PrivateKey(data.Key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

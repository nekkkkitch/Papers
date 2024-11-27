package authserver

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"log"
	"net"

	"papers/pkg/crypt"
	pb "papers/pkg/grpc/pb/authService"
	"papers/pkg/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	Port string `yaml:"auth_port" env-prefix:"AUTHPORT"`
}

type IJWTManager interface {
	GetPrivateKey() *rsa.PrivateKey
	CreateTokens(user_id uuid.UUID) (string, string, error)
	GetIDFromToken(token string) (*uuid.UUID, error)
}

type IDBManager interface {
	CheckSameLogin(login string) (bool, error)
	AddUser(user models.User) (*uuid.UUID, error)
	GetUserByID(id uuid.UUID) (models.User, error)
	GetUserByLogin(login string) (models.User, error)
	InsertRefreshToken(token string, id uuid.UUID) error
	GetRefreshToken(id uuid.UUID) (string, error)
}

type server struct {
	pb.UnimplementedAuthentificationServer
	jwt IJWTManager
	db  IDBManager
}

type Service struct {
	AuthServer *grpc.Server
	Listener   *net.Listener
	cfg        *Config
	jwt        IJWTManager
	db         IDBManager
}

// Создание сервера сервиса аутентификации
func New(cfg *Config, jwt IJWTManager, db IDBManager) (*Service, error) {
	log.Println(cfg.Port)
	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	pb.RegisterAuthentificationServer(s, &server{jwt: jwt, db: db})
	log.Printf("Auth server listening at %v\n", lis.Addr())
	return &Service{AuthServer: s, Listener: &lis, cfg: cfg, jwt: jwt, db: db}, nil
}

// Регистрация пользователя(в т.ч. валидация ника и пароля, проверка на наличие пользователя в БД, добавление его в БД и возврат токенов)
func (s *server) Register(_ context.Context, in *pb.User) (*pb.AuthData, error) {
	log.Println("User to register: " + in.Login)
	if in.Login == "" || in.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request missing login or password")
	}
	if same, err := s.db.CheckSameLogin(in.Login); err != nil || same {
		if same {
			return nil, status.Errorf(codes.AlreadyExists, "login occupied")
		}
		log.Println("Something went wrong when checked for the same login: " + err.Error())
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	cryptedPassword, err := crypt.CryptPassword(in.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	id, err := s.db.AddUser(models.User{Login: in.Login, Password: string(cryptedPassword)})
	if err != nil {
		log.Println("Something went wrong when added user: " + err.Error())
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	access, refresh, err := s.jwt.CreateTokens(*id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	err = s.db.InsertRefreshToken(refresh, *id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.AuthData{AccessToken: access, RefreshToken: refresh}, nil
}

// Проверка данных аутентификации и создание и возвращение токенов в случае успеха
func (s *server) Login(_ context.Context, in *pb.User) (*pb.AuthData, error) {
	log.Println("User to login: " + in.Login)
	if in.Login == "" || in.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request missing login or password")
	}
	if same, err := s.db.CheckSameLogin(in.Login); err != nil || !same {
		if !same {
			return nil, status.Errorf(codes.InvalidArgument, "login does not exist")
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	user, err := s.db.GetUserByLogin(in.Login)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		if err.Error() == bcrypt.ErrMismatchedHashAndPassword.Error() {
			return nil, status.Errorf(codes.InvalidArgument, "wrong password")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	access, refresh, err := s.jwt.CreateTokens(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	err = s.db.InsertRefreshToken(refresh, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.AuthData{AccessToken: access, RefreshToken: refresh}, nil
}

// Создание новых токенов(нужно для обновления access токена при его истекшем сроке годности)
func (s *server) UpdateTokens(_ context.Context, in *pb.AuthData) (*pb.AuthData, error) {
	user_id, err := s.jwt.GetIDFromToken(in.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	realRefreshToken, err := s.db.GetRefreshToken(*user_id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if realRefreshToken != in.RefreshToken {
		return nil, status.Errorf(codes.InvalidArgument, "Refresh token was changed, consider relogin")
	}
	access, refresh, err := s.jwt.CreateTokens(*user_id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	err = s.db.InsertRefreshToken(access, *user_id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.AuthData{AccessToken: access, RefreshToken: refresh}, nil
}

// Получение приватного ключа(см jwt.NewWithKey())
func (s *server) GetPrivateKey(_ context.Context, in *pb.KeyRequest) (*pb.PrivateKey, error) {
	return &pb.PrivateKey{Key: x509.MarshalPKCS1PrivateKey(s.jwt.GetPrivateKey())}, nil
}

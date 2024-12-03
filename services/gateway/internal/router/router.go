package router

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"papers/pkg/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/google/uuid"
)

type Router struct {
	App    *fiber.App
	Config *Config
	pps    IPapersService
	asvc   IAuthService
	jwt    IJWTManager
}

type Config struct {
	Host string `yaml:"router_host" env-prefix:"ROUTERHOST"`
	Port string `yaml:"router_port" env-prefix:"ROUTERPORT"`
}

type IAuthService interface {
	Register(user models.User) (*models.AuthData, error)
	Login(user models.User) (*models.AuthData, error)
	UpdateTokens(tokens models.AuthData) (*models.AuthData, error)
}

type IPapersService interface {
	GetAvailablePapers() ([]byte, error)
}

type IJWTManager interface {
	GetPublicKey() *rsa.PublicKey
	GetIDFromToken(token string) (*uuid.UUID, error)
	ValidateToken(c *fiber.Ctx, key string) (bool, error)
	AuthFilter(c *fiber.Ctx) bool
	RefreshFilter(c *fiber.Ctx) bool
}

// Создание рутов для запросов с применением middleware для проверки валидности токенов и началом получения сообщений из брокера
func New(cfg *Config, auservice IAuthService, pps IPapersService, jwt IJWTManager) (*Router, error) {
	app := fiber.New()
	router := Router{App: app, Config: cfg, jwt: jwt, asvc: auservice, pps: pps}
	router.App.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			err := c.Next()
			if err != nil {
				log.Println(err.Error())
				return err
			}
		}
		return nil
	})
	router.App.Use(cors.New(cors.Config{
		AllowHeaders: "X-Access-Token, X-Refresh-Token",
	}))
	router.App.Use(keyauth.New(keyauth.Config{
		Next:         router.jwt.AuthFilter,
		KeyLookup:    "header:X-Access-Token",
		Validator:    router.jwt.ValidateToken,
		ErrorHandler: router.ErrorHandler(),
	}))
	router.App.Use(keyauth.New(keyauth.Config{
		Next:         router.jwt.RefreshFilter,
		KeyLookup:    "header:X-Refresh-Token",
		Validator:    router.jwt.ValidateToken,
		ErrorHandler: router.ErrorHandler(),
	}))
	router.App.Post("/login", router.Login())
	router.App.Post("/register", router.Register())
	router.App.Get("/refresh", router.UpdateTokens())
	router.App.Get("/ping", Ping)
	router.App.Get("/papers", router.GetPapers())
	return &router, nil
}

func (r *Router) Listen() {
	r.App.Listen(r.Config.Host + r.Config.Port)
}

func (r *Router) GetPapers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bod, err := r.pps.GetAvailablePapers()
		if err != nil {
			log.Println("Failed to get available papers:", err)
			c.Status(500)
			return err
		}
		c.Write(bod)
		return nil
	}
}

// Логиним пользователя
func (r *Router) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user models.User
		err := json.Unmarshal(c.Body(), &user)
		if err != nil {
			return err
		}
		log.Printf("User to login: %s\n", user.Login)
		authData, err := r.asvc.Login(user)
		if err != nil {
			switch err.Error() {
			case "rpc error: code = AlreadyExists desc = login occupied":
				c.Status(fiber.StatusBadRequest)
				return nil
			}
			return err
		}
		log.Printf("Tokens to return:\nAccess token: %s\nRefresh token: %s", authData.AccessToken[:20], authData.RefreshToken[:20])
		c.Context().Response.Header.Set("X-Access-Token", authData.AccessToken)
		c.Context().Response.Header.Set("X-Refresh-Token", authData.RefreshToken)
		return nil
	}
}

// Регистрация пользователя
func (r *Router) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user models.User
		err := json.Unmarshal(c.Body(), &user)
		if err != nil {
			return err
		}
		log.Printf("User to register: %s\n", user.Login)
		authData, err := r.asvc.Register(user)
		if err != nil {
			return err
		}
		log.Printf("Tokens to return:\nAccess token: %s\nRefresh token: %s", authData.AccessToken[:20], authData.RefreshToken[:20])
		c.Context().Response.Header.Set("X-Access-Token", authData.AccessToken)
		c.Context().Response.Header.Set("X-Refresh-Token", authData.RefreshToken)
		return nil
	}
}

// Создание новой пары токенов
func (r *Router) UpdateTokens() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authData := models.AuthData{AccessToken: c.GetReqHeaders()["X-Access-Token"][0], RefreshToken: c.GetReqHeaders()["X-Refresh-Token"][0]}
		log.Printf("Got tokens:\nAccess token: %s\nRefresh token: %s", authData.AccessToken[:20], authData.RefreshToken[:20])
		authDataResp, err := r.asvc.UpdateTokens(authData)
		if err != nil {
			return err
		}

		log.Printf("Tokens to return:\nAccess token: %s\nRefresh token: %s", authDataResp.AccessToken[:20], authDataResp.RefreshToken[:20])
		c.Context().Response.Header.Set("X-Access-Token", authDataResp.AccessToken)
		c.Context().Response.Header.Set("X-Refresh-Token", authDataResp.RefreshToken)
		return nil
	}
}

// Хэндл ошибок для мидлвейра
func (r *Router) ErrorHandler() func(c *fiber.Ctx, err error) error {
	return func(c *fiber.Ctx, err error) error {
		log.Println("Bad access token: ", c.GetReqHeaders()["X-Access-Token"])
		log.Println("Bad refresh token: ", c.GetReqHeaders()["X-Refresh-Token"])
		log.Println("Wrong jwts: " + err.Error())
		return err
	}
}

// Пингуем сервер
func Ping(c *fiber.Ctx) error {
	log.Println("Ping")
	return c.JSON("Ping")
}

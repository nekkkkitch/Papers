package router

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"gateway/internal/pkg/models"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	App     *fiber.App
	Config  *Config
	balance IBalanceService
	pps     IPapersService
	asvc    IAuthService
	jwt     IJWTManager
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

type IBalanceService interface {
	GetBalance(*uuid.UUID) (float32, error)
	AddBalance(*models.Money) (string, error)
	TakeBalance(*models.Money) (string, error)
}

type IPapersService interface {
	GetAvailablePapers() ([]byte, error)
	GetUserPapers(userID *uuid.UUID) ([]byte, error)
	SellPaper(userID *uuid.UUID, paper models.Paper) (string, error)
	BuyPaper(userID *uuid.UUID, paper models.Paper) (string, error)
}

type IJWTManager interface {
	GetPublicKey() *rsa.PublicKey
	GetIDFromToken(token string) (*uuid.UUID, error)
	ValidateToken(c *fiber.Ctx, key string) (bool, error)
	AuthFilter(c *fiber.Ctx) bool
	RefreshFilter(c *fiber.Ctx) bool
}

var (
	AllPapersRequest = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "AllPapers",
		Help: "Number of times users requiered to see all papers",
	})

	GeneralAmountOfRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "General",
		Help: "Number of times users requiered to see all papers",
	})
)

// Создание рутов для запросов с применением middleware для проверки валидности токенов и началом получения сообщений из брокера
func New(cfg *Config, auservice IAuthService, pps IPapersService, balance IBalanceService, jwt IJWTManager) (*Router, error) {
	app := fiber.New()
	router := Router{App: app, Config: cfg, jwt: jwt, asvc: auservice, pps: pps, balance: balance}
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
	registerMetrics()

	router.App.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	router.App.Get("/ping", Ping)

	router.App.Post("/login", router.Login())
	router.App.Post("/register", router.Register())
	router.App.Get("/refresh", router.UpdateTokens())

	router.App.Get("/papers", router.GetPapers())
	router.App.Post("/buypaper", router.BuyPaper())
	router.App.Post("/sellpaper", router.SellPaper())
	router.App.Get("/mypapers", router.GetUserPapers())

	router.App.Get("/balance", router.GetBalance())
	router.App.Post("/addbalance", router.AddBalance())
	router.App.Post("/takebalance", router.TakeBalance())
	return &router, nil
}

func (r *Router) Listen() error {
	err := r.App.Listen(r.Config.Host + r.Config.Port)
	if err != nil {
		return err
	}
	return nil
}

func (r *Router) GetPapers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer AllPapersRequest.Inc()
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

func (r *Router) SellPaper() fiber.Handler {
	return func(c *fiber.Ctx) error {
		access := c.GetReqHeaders()["X-Access-Token"][0]
		var paper models.Paper
		err := json.Unmarshal(c.Body(), &paper)
		if err != nil {
			log.Println("unmarshal error:", err)
			c.Status(500)
			return nil
		}
		userID, err := r.jwt.GetIDFromToken(access)
		if err != nil {
			log.Println("getting id from token error:", err)
			c.Status(500)
			return nil
		}
		bod, err := r.pps.SellPaper(userID, paper)
		if err != nil {
			log.Println("selling paper error:", err)
			c.Status(500)
			return nil
		}
		c.Status(200)
		c.WriteString(bod)
		return nil
	}
}

func (r *Router) BuyPaper() fiber.Handler {
	return func(c *fiber.Ctx) error {
		access := c.GetReqHeaders()["X-Access-Token"][0]
		var paper models.Paper
		err := json.Unmarshal(c.Body(), &paper)
		if err != nil {
			log.Println("unmarshal error:", err)
			c.Status(500)
			return nil
		}
		userID, err := r.jwt.GetIDFromToken(access)
		if err != nil {
			log.Println("getting id from token error:", err)
			c.Status(500)
			return nil
		}
		bod, err := r.pps.BuyPaper(userID, paper)
		if err != nil {
			log.Println("buying paper error:", err)
			c.Status(500)
			return nil
		}
		c.Status(200)
		c.WriteString(bod)
		return nil
	}
}

func (r *Router) GetUserPapers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		access := c.GetReqHeaders()["X-Access-Token"][0]
		userID, err := r.jwt.GetIDFromToken(access)
		if err != nil {
			log.Println("getting id from token error:", err)
			c.Status(500)
			return nil
		}
		bod, err := r.pps.GetUserPapers(userID)
		if err != nil {
			log.Println("getting user papers error:", err)
			c.Status(500)
			return nil
		}
		c.Status(200)
		c.Write(bod)
		return nil
	}
}

func (r *Router) GetBalance() fiber.Handler {
	return func(c *fiber.Ctx) error {
		access := c.GetReqHeaders()["X-Access-Token"][0]
		userID, err := r.jwt.GetIDFromToken(access)
		if err != nil {
			log.Println("getting id from token error:", err)
			c.Status(500)
			return nil
		}
		balance, err := r.balance.GetBalance(userID)
		if err != nil {
			log.Println("getting user balance error:", err)
			c.Status(500)
			return nil
		}
		c.Status(200)
		c.WriteString(fmt.Sprintf("Your balance is %.2f", balance))
		return nil
	}
}

func (r *Router) AddBalance() fiber.Handler {
	return func(c *fiber.Ctx) error {
		access := c.GetReqHeaders()["X-Access-Token"][0]
		money := models.Money{}
		err := json.Unmarshal(c.Body(), &money)
		if err != nil {
			log.Println("unmarshal error:", err)
			c.Status(500)
			return nil
		}
		userID, err := r.jwt.GetIDFromToken(access)
		if err != nil {
			log.Println("getting id from token error:", err)
			c.Status(500)
			return nil
		}
		money.ID = *userID
		bod, err := r.balance.AddBalance(&money)
		if err != nil {
			log.Println("adding balance error:", err)
			c.Status(500)
			return nil
		}
		c.Status(200)
		c.WriteString(bod)
		return nil
	}
}

func (r *Router) TakeBalance() fiber.Handler {
	return func(c *fiber.Ctx) error {
		access := c.GetReqHeaders()["X-Access-Token"][0]
		money := models.Money{}
		err := json.Unmarshal(c.Body(), &money)
		if err != nil {
			log.Println("unmarshal error:", err)
			c.Status(500)
			return nil
		}
		userID, err := r.jwt.GetIDFromToken(access)
		if err != nil {
			log.Println("getting id from token error:", err)
			c.Status(500)
			return nil
		}
		money.ID = *userID
		bod, err := r.balance.TakeBalance(&money)
		if err != nil {
			log.Println("taking from balance error:", err)
			c.Status(500)
			return nil
		}
		c.Status(200)
		c.WriteString(bod)
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

func registerMetrics() {
	prometheus.MustRegister(AllPapersRequest)
	prometheus.MustRegister(GeneralAmountOfRequests)
}

package main

import (
	"log"
	"papers/pkg/jwt"
	rtr "papers/services/gateway/internal/router"
	aus "papers/services/gateway/internal/services/authService"
	balance "papers/services/gateway/internal/services/balanceService"
	pps "papers/services/gateway/internal/services/ppsService"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	JWTConfig     *jwt.Config     `yaml:"jwt" env-prefix:"JWT_"`
	AUSConfig     *aus.Config     `yaml:"aus" env-prefix:"AUS_"`
	PPSConfig     *pps.Config     `yaml:"pps"`
	RTRConfig     *rtr.Config     `yaml:"rtr" env-prefix:"RTR_"`
	BalanceConfig *balance.Config `yaml:"balance" env-prefix:"BALANCE_"`
}

func readConfig(filename string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(filename, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := readConfig("./cfg.yml")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Config file read successfully")
	jwt, err := jwt.New(cfg.JWTConfig)
	if err != nil {
		log.Fatalln("Failed to create jwt: " + err.Error())
	}
	authService, err := aus.New(cfg.AUSConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Auth service connected successfully")
	ppsService, err := pps.New(cfg.PPSConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Auth service connected successfully")
	balanceService, err := balance.New(cfg.BalanceConfig)
	router, err := rtr.New(cfg.RTRConfig, authService, ppsService, balanceService, &jwt)
	if err != nil {
		log.Fatalln("Failed to host router:", err.Error())
	}
	log.Printf("Router is listening on %v:%v\n", router.Config.Host, router.Config.Port)
	router.Listen()
}

package main

import (
	pg "balance/internal/db"
	server "balance/internal/server"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBConfig            *pg.Config     `yaml:"db" env-prefix:"DB_"`
	BalanceServerConfig *server.Config `yaml:"balance"`
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
	log.Println(cfg.DBConfig)
	db, err := pg.New(cfg.DBConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("DB connected successfully")
	server, err := server.New(cfg.BalanceServerConfig, db)
	if err != nil {
		log.Fatalln(err)
	}
	if err := server.Server.Serve(*server.Listener); err != nil {
		log.Fatalln(err)
	}
	log.Println("Service connected successfully")
}

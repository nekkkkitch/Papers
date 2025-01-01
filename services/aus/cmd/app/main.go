package main

import (
	pg "aus/internal/db"
	"aus/internal/pkg/jwt"
	server "aus/internal/server"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	JWTConfig        *jwt.Config    `yaml:"jwt" env-prefix:"JWT_"`
	DBConfig         *pg.Config     `yaml:"db" env-prefix:"DB_"`
	AuthServerConfig *server.Config `yaml:"aus" env-prefix:"AS_"`
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
	jwt, err := jwt.New(cfg.JWTConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("JWT created successfully")
	service, err := server.New(cfg.AuthServerConfig, &jwt, db)
	if err != nil {
		log.Fatalln(err)
	}
	if err := service.AuthServer.Serve(*service.Listener); err != nil {
		log.Fatalln(err)
	}
	log.Println("Service connected successfully")
}

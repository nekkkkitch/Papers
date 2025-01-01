package main

import (
	"log"
	"papers/internal/db"
	"papers/internal/redis"
	"papers/internal/server"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBConfig  *db.Config     `yaml:"db"`
	RDSConfig *redis.Config  `yaml:"rds"`
	PPSConfig *server.Config `yaml:"pps"`
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
	db, err := db.New(cfg.DBConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("DB connected successfully")
	rds, err := redis.New(cfg.RDSConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Redis client created successfully")
	server, err := server.New(cfg.PPSConfig, db, rds)
	if err != nil {
		log.Fatalln(err)
	}
	err = server.PpsServer.Serve(*server.Listener)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Server started successfully")
}

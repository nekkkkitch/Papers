package main

import (
	"log"
	pg "market/internal/db"
	"market/internal/redis"
	"market/internal/service"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBConfig  *pg.Config    `yaml:"db" env-prefix:"DB_"`
	RDSConfig *redis.Config `yaml:"rds"`
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
	rds, err := redis.New(cfg.RDSConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Redis connected successfully")
	svc, err := service.New(db, rds)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Market started successfully")
	svc.StartFun()
}

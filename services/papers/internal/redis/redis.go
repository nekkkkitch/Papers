package redis

import (
	"context"
	"log"
	"papers/pkg/models"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"dbnum"`
}

type Redis struct {
	client *redis.Client
}

func New(cfg *Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ret := Redis{rdb}
	return &ret, nil
}

func (r *Redis) GetAvailablePapers() ([]models.Paper, error) {
	var keys []string
	var cursor uint64
	keys, _, err := r.client.Scan(context.Background(), cursor, "stock:*", 0).Result()
	if err != nil {
		return nil, err
	}
	log.Println("Got papers:", keys)
	papers := make([]models.Paper, 0, len(keys))
	for _, key := range keys {
		price, err := r.client.Get(context.Background(), key).Result()
		if err != nil {
			log.Println("Cant read stock:", err)
			return nil, err
		}
		priceFloat, err := strconv.ParseFloat(price, 32)
		if err != nil {
			log.Println("Cant int price:", err)
			return nil, err
		}
		papers = append(papers, models.Paper{Name: key, Price: float32(priceFloat)})
	}
	return papers, nil
}

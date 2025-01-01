package redis

import (
	"context"
	"encoding/json"
	"log"
	"papers/internal/pkg/models"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
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
	log.Printf("Attempt to get available papers\n")
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
			log.Println("Cant parse price:", err)
			return nil, err
		}
		papers = append(papers, models.Paper{Name: key, Price: float32(priceFloat)})
	}
	return papers, nil
}

func (r *Redis) GetPaperPrice(name string) (float32, error) {
	log.Printf("Attempt to get paper %v price\n", name)
	stringPrice, err := r.client.Get(context.Background(), "stock:"+name).Result()
	if err != nil {
		log.Println("Cant get paper price:", err)
		return 0, err
	}
	price, err := strconv.ParseFloat(stringPrice, 32)
	if err != nil {
		log.Println("Cant parse price:", err)
		return 0, err
	}
	return float32(price), nil
}
func (r *Redis) GetUserPapers(userId uuid.UUID) ([]models.Paper, error) {
	log.Printf("Attempt to get user with uuid %v papers from redis\n", userId)
	stringJson, err := r.client.Get(context.Background(), "user_papers:"+userId.String()).Result()
	if err != nil {
		log.Println("Cant get user papers:", err)
		return nil, err
	}
	var papers []models.Paper
	err = json.Unmarshal([]byte(stringJson), &papers)
	if err != nil {
		log.Println("Cant unmarshal user papers:", err)
		return nil, err
	}
	return papers, nil
}
func (r *Redis) UpdateUserPapers(userId uuid.UUID, papers []models.Paper) error {
	log.Printf("Attempt to update user with uuid %v papers\n", userId)
	bytes, err := json.Marshal(papers)
	if err != nil {
		log.Println("Cant marhsal papers:", err)
		return err
	}
	err = r.client.Set(context.Background(), "user_papers:"+userId.String(), string(bytes), 0).Err()
	if err != nil {
		log.Println("Cant update user papers:", err)
		return err
	}
	return nil
}

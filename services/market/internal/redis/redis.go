package redis

import (
	"context"
	"log"

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

func (r *Redis) UpdateStock(name string, value float32) error {
	name = "stock:" + name
	value = float32(int(value*100)) / 100
	err := r.client.Set(context.Background(), name, value, 0)
	if err.Err() != nil {
		log.Printf("Failed to update stock named %v with value %v: %v\n", name, value, err.Err())
		return err.Err()
	}
	return nil
}

// пока не нужно ну да пускай будет(удалить при Finale)
/*
func (r *Redis) GetStock(name string) (float32, error) {
	valueString, err := r.client.Get(context.Background(), "stock:"+name).Result()
	if err != nil {
		log.Printf("Failed to get stock named %v: %v\n", name, err)
		return -1, err
	}
	value, err := strconv.ParseFloat(valueString, 32)
	if err != nil {
		log.Printf("Failed to parse stock value=%v: %v\n", valueString, err)
		return -1, err
	}
	return float32(value), nil
}
*/

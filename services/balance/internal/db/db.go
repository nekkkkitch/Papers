package db

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Config struct {
	Host     string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"PORT" env-default:"5432"`
	User     string `yaml:"user" env:"USER" env-default:"postgres"`
	Password string `yaml:"password" env:"password" env-default:"postgres"`
	DBName   string `yaml:"dbname" env:"DBNAME" env-default:"chat"`
}

type DB struct {
	config *Config
	db     *pgx.Conn
}

// Создает соединение с существующей БД
func New(cfg *Config) (*DB, error) {
	d := &DB{config: cfg}
	connection := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := pgx.Connect(context.Background(), connection)
	log.Println("Connecting to: " + connection)
	if err != nil {
		return nil, err
	}
	d.db = db
	return d, nil
}

// Закрывает соединение с БД
func (d *DB) Close() error {
	return d.db.Close(context.Background())
}

func (d *DB) GetUserBalance(userID uuid.UUID) (float32, error) {
	var balance pgtype.Float4
	err := d.db.QueryRow(context.Background(), `select balance from public.users where id=$1`, userID).Scan(&balance)
	if err != nil {
		return -1, err
	}
	return balance.Float32, nil
}

func (d *DB) ChangeBalance(userID uuid.UUID, cash float32) error {
	_, err := d.db.Exec(context.Background(), `update public.users set balance = balance + $1 where id = $2`, cash, userID)
	if err != nil {
		log.Println("Cant change user balance")
		return err
	}
	return nil
}

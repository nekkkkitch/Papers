package db

import (
	"context"
	"fmt"
	"log"
	"papers/pkg/models"

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

func (d *DB) GetListOfPapers() ([]models.Paper, error) {
	res, err := d.db.Query(context.Background(), `select name, price from public.papers`)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []models.Paper{}, nil
		}
		log.Println("Failed to get papers from db:", err)
		return nil, err
	}
	papers := []models.Paper{}
	for res.Next() {
		var name pgtype.Text
		var value pgtype.Float4
		err := res.Scan(&name, &value)
		if err != nil {
			log.Println("Failed to scan result:", err)
			return nil, err
		}
		papers = append(papers, models.Paper{Name: name.String, Price: value.Float32})
	}
	return papers, nil
}

func (d *DB) PutValue(name string, value float32) error {
	_, err := d.db.Exec(context.Background(), `update public.papers set price = $1 where name = $2`, value, name)
	if err != nil {
		log.Println("Failed to put new value:", err)
		return err
	}
	return nil
}

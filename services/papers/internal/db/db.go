package db

import (
	"context"
	"fmt"
	"log"
	"papers/pkg/models"

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

func (d *DB) GetUserPapers(userId uuid.UUID) ([]models.Paper, error) {
	log.Printf("Attempt to get user with uuid %v papers", userId)
	res, err := d.db.Query(context.Background(), `select paper_name, amount from public.storage where id = $1`, userId)
	if err != nil {
		log.Println("Cant get user papers:", err)
		return nil, err
	}
	papers := make([]models.Paper, 0, 3)
	for res.Next() {
		var paper models.Paper
		err := res.Scan(&paper.Name, &paper.Amount)
		if err != nil {
			log.Println("Cant scan paper:", err)
			return nil, err
		}
		papers = append(papers, paper)
	}
	return papers, nil
}

func (d *DB) GetUserBalance(userId uuid.UUID) (float32, error) {
	log.Printf("Attempt to get user with uuid %v balance\n", userId)
	var balance pgtype.Float4
	err := d.db.QueryRow(context.Background(), `select balance from public.users where id = $1`, userId).Scan(&balance)
	if err != nil {
		log.Println("Cant get user balance:", err)
		return 0, err
	}
	return balance.Float32, nil
}

func (d *DB) ChangeBalance(userId uuid.UUID, change float32) error {
	log.Printf("Attempt to change user with uuid %v balance, adding %v\n", userId, change)
	_, err := d.db.Exec(context.Background(), `update public.users set balance = balance + $1 where id = $2`, change, userId)
	if err != nil {
		log.Println("Cant change user balance:", err)
		return err
	}
	return nil
}

func (d *DB) AddPaper(userId uuid.UUID, name string, amount int) error {
	log.Printf("Attempt to add user with uuid %v papers %v, adding %v\n", userId, name, amount)
	_, err := d.db.Exec(context.Background(), `insert into public.storage(id, paper_name, amount) values($1, $2, $3)`, userId, name, amount)
	if err != nil {
		log.Println("Cant change user paper amount:", err)
		return err
	}
	return nil
}

func (d *DB) ChangePaperAmount(userId uuid.UUID, name string, amount int) error {
	log.Printf("Attempt to change user with uuid %v amount of papers %v, adding %v\n", userId, name, amount)
	_, err := d.db.Exec(context.Background(), `update public.storage set amount = amount + $1 where id = $2 and paper_name = $3`, amount, userId, name)
	if err != nil {
		log.Println("Cant change user paper amount:", err)
		return err
	}
	return nil
}

func (d *DB) GetPaperAmount(userId uuid.UUID, name string) (int, error) {
	log.Printf("Attempt to get user with uuid %v amount of paper %v\n", userId, name)
	var amount pgtype.Int4
	err := d.db.QueryRow(context.Background(), `select amount from public.storage where id = $1 and paper_name = $2`, userId, name).Scan(&amount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, nil
		}
		log.Println("Cant get user paper amount:", err)
		return 0, err
	}
	return int(amount.Int32), nil
}

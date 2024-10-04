package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"time"
)

type ExchangeRateInfo struct {
	USDBRL CurrencyDetails `json:"USDBRL"`
}

type CurrencyDetails struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varbid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

const ApiUrl = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

func init() {
	err := createDB("cotacao.db")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
}
func RetrieveDolarRate(ctx context.Context, urlPath string) (string, error) {
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result ExchangeRateInfo

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.USDBRL.Bid, nil
}

func StoreInDatabase() error {
	ctxAPI, cancelAPI := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelAPI()

	d, err := RetrieveDolarRate(ctxAPI, ApiUrl)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			_ = fmt.Errorf("A operação execedeu o tempo limite 2 %s", err, http.StatusGatewayTimeout)
		}
		return err
	}

	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDB()

	_, err = DBHandle.ExecContext(
		ctxDB,
		`INSERT INTO cotacao (valor, moeda) VALUES (?,?);`, d, "USD",
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Errorf("A operação execedeu o tempo limite 1 %s", err, http.StatusGatewayTimeout)
		}
		return err
	}
	return nil
}

var DBHandle *sql.DB

func createDB(dbPath string) error {
	var err error
	DBHandle, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	_, err = DBHandle.ExecContext(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS cotacao (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		valor FLOAT NOT NULL,
    		moeda TEXT NOT NULL
)`,
	)
	if err != nil {
		return err
	}
	fmt.Println("Database and table created successfully")
	return nil
}

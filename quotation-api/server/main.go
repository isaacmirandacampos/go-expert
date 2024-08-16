package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type MoneyConversion struct {
	USDBRL struct {
		Code   string `json:"code"`
		Codein string `json:"codein"`
		Name   string `json:"name"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Bid    string `json:"bid"`
		Ask    string `json:"ask"`
	} `json:"USDBRL"`
}

type MoneyConversionResponse struct {
	Bid string `json:"price"`
}

func main() {
	err := migration()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/usd-to-brl", handler)
	http.ListenAndServe(":8081", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var moneyConversion MoneyConversion
	err := getMoneyConversion(&moneyConversion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = saveMoneyConversion(&moneyConversion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	moneyConversionResponse := MoneyConversionResponse{
		Bid: moneyConversion.USDBRL.Bid,
	}
	json.NewEncoder(w).Encode(moneyConversionResponse)
}

func getMoneyConversion(c *MoneyConversion) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return errors.New("error creating request")
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return errors.New("error executing request")
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("error reading response")
	}

	err = json.Unmarshal(res, c)
	if err != nil {
		return errors.New("error unmarshalling response")
	}
	select {
	case <-ctx.Done():
		log.Println("Request")
		return errors.New("request timeout")
	case <-time.After(200 * time.Millisecond):
		log.Println("Request between acceptable time")
	}
	return nil
}

func saveMoneyConversion(c *MoneyConversion) error {
	db, err := db_connection()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	defer db.Close()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("error beginning transaction")
	}
	stmt, err := db.PrepareContext(ctx, `INSERT INTO money_conversions (
		code,
		codein,
		name,
		high,
		low,
		bid,
		ask) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return errors.New("error preparing statement")
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		c.USDBRL.Code,
		c.USDBRL.Codein,
		c.USDBRL.Name,
		c.USDBRL.High,
		c.USDBRL.Low,
		c.USDBRL.Bid,
		c.USDBRL.Ask,
	)
	if err != nil {
		return errors.New("error inserting into database")
	}
	if err = tx.Commit(); err != nil {
		return errors.New("error committing transaction")
	}
	select {
	case <-ctx.Done():
		log.Println("Insert Timeout")
		return errors.New("timeout")
	case <-time.After(200 * time.Millisecond):
		log.Println("Insert between acceptable time")
	}
	return nil
}

func db_connection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./money_conversion.db")
	if err != nil {
		return nil, errors.New("error opening database")
	}
	return db, nil
}

func migration() error {
	db, err := db_connection()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS money_conversions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code VARCHAR(10),
		codein VARCHAR(10),
		name VARCHAR(255),
		high REAL,
		low REAL,
		bid REAL,
		ask REAL
	)`)
	if err != nil {
		return errors.New("error creating table")
	}
	return nil
}

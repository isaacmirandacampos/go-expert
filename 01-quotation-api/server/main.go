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

type DollarConversion struct {
	USDBRL struct {
		Price string `json:"bid"`
	} `json:"USDBRL"`
}

type DollarConversionResponse struct {
	Price string `json:"price"`
}

func main() {
	log.Default().Println("Starting...")
	err := migration()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/usd-to-brl", handler)
	log.Default().Println("Running on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var dollarConversion DollarConversion
	err := getDollarConversion(&dollarConversion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = saveDollarConversion(&dollarConversion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dollarConversionResponse := DollarConversionResponse{
		Price: dollarConversion.USDBRL.Price,
	}
	json.NewEncoder(w).Encode(dollarConversionResponse)
}
func getDollarConversion(c *DollarConversion) error {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return errors.New("error creating request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Verifica se o erro foi causado pelo timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Request Timeout economia.awesomeapi.com.br")
			return errors.New("Request Timeout economia.awesomeapi.com.br")
		}
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

	return nil
}

func saveDollarConversion(c *DollarConversion) error {
	db, err := db_connection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Usando contexto com timeout de 10ms para a transação
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Inicia a transação
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		// Verifica se o erro foi causado pelo timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Transaction Timeout")
			return errors.New("Transaction Timeout")
		}
		return errors.New("error beginning transaction")
	}

	// Prepara a consulta de inserção
	stmt, err := db.PrepareContext(ctx, `INSERT INTO usd_to_brl_conversions (price) VALUES (?)`)
	if err != nil {
		// Verifica se o erro foi causado pelo timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Prepare Timeout")
			return errors.New("Prepare Timeout")
		}
		return errors.New("error preparing statement")
	}
	defer stmt.Close()

	// Executa a inserção
	_, err = stmt.Exec(c.USDBRL.Price)
	if err != nil {
		// Verifica se o erro foi causado pelo timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Execute Timeout")
			return errors.New("Execute Timeout")
		}
		return errors.New("error inserting into database")
	}

	// Commit da transação
	if err = tx.Commit(); err != nil {
		// Verifica se o erro foi causado pelo timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Commit Timeout")
			return errors.New("Commit Timeout")
		}
		return errors.New("error committing transaction")
	}

	// Nenhum erro ocorrido, a transação foi bem-sucedida
	return nil
}

func db_connection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./conversions.db")
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
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS usd_to_brl_conversions (
		price REAL
	)`)
	if err != nil {
		return errors.New("error creating table")
	}
	return nil
}

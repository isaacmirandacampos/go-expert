package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Quotation struct {
	Dollar string `json:"price"`
}

func main() {
	log.Default().Println("Starting...")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8080/usd-to-brl", nil)
	if err != nil {
		log.Fatal("error creating request")
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("error executing request")
		return
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("error reading response")
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("StatusCode: %v \n", resp.StatusCode)
		log.Println("Response received: " + string(res))
		return
	}
	var q Quotation
	log.Println("Response received:", string(res))
	err = json.Unmarshal(res, &q)
	if err != nil {
		log.Fatal("error unmarshalling response")
		return
	}
	err = StoreInTxt(&q)
	if err != nil {
		log.Fatal("error storing in txt")
		return
	}
}

func StoreInTxt(q *Quotation) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return errors.New("error creating file")
	}
	defer file.Close()
	_, err = file.WriteString("Dólar: " + q.Dollar + "\n")
	if err != nil {
		return errors.New("error writing to file: " + err.Error())
	}
	return nil
}

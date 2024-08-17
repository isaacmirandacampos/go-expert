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
	Dolar string `json:"price"`
}

func main() {
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8081/usd-to-brl", nil)
	if err != nil {
		http.Error(w, "error creating request", http.StatusInternalServerError)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "error executing request", http.StatusInternalServerError)
		return
	}
	select {
	case <-ctx.Done():
		log.Println("Timeout of request")
		http.Error(w, "timeout", http.StatusRequestTimeout)
		return
	case <-time.After(300 * time.Millisecond):
		log.Println("Response between acceptable time")
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "error reading response", http.StatusInternalServerError)
		return
	}
	var q Quotation
	err = json.Unmarshal(res, &q)
	if err != nil {
		http.Error(w, "error unmarshalling response", http.StatusInternalServerError)
		return
	}
	err = StoreInTxt(&q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(q)

}

func StoreInTxt(q *Quotation) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return errors.New("error creating file")
	}
	defer file.Close()
	_, err = file.WriteString("Dólar: " + q.Dolar)
	if err != nil {
		return errors.New("error writing file")
	}
	return nil
}

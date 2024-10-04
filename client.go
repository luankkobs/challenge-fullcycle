package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancelFunc()

		dolarRate, err := RetrieveDolarRate(ctx, ApiUrl)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "The operation exceeded the time limit 3", http.StatusGatewayTimeout)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = WriteToFile(dolarRate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = StoreInDatabase()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dolarRate)
	})
	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func WriteToFile(dolarRate string) error {
	rateText := []byte(fmt.Sprintf("Dólar: %s \n", dolarRate))
	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(rateText); err != nil {
		return err
	}
	return nil
}

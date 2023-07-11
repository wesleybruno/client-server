package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	initializeSqLIte()
	router := mux.NewRouter()
	router.HandleFunc("/cotacao", buscarCotacao).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

type DolarResponse struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type ApiResponse struct {
	Bid string `json:"bid"`
}

func buscarCotacao(w http.ResponseWriter, r *http.Request) {

	cotacao, err := makeRequest(r)
	if err != nil {
		fmt.Println("Erro ao fazer request:", err)
		sendError(w, err)
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	err = salvarBancoLocal(cotacao, ctx)
	if err != nil {
		fmt.Println("Erro ao fazer parse do result:", err)
		sendError(w, err)
		return
	}

	err = sendResponse(w, cotacao)
	if err != nil {
		fmt.Println("Erro ao fazer parse do result:", err)
		sendError(w, err)
		return
	}
}

func makeRequest(r *http.Request) (*DolarResponse, error) {
	ctxRequest, cancelRequest := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancelRequest()
	req, err := http.NewRequestWithContext(ctxRequest, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return nil, err
	}

	var cotacao DolarResponse
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		fmt.Println("Erro ao fazer parse do result:", err)
		return nil, err
	}
	return &cotacao, nil
}

func openDb() (*sql.DB, error) {

	dbPath := "./db/main.db"
	//Verify if DB exists

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println("Erro ao criar db:", err)
		return nil, err
	}

	return db, nil
}

func salvarBancoLocal(data *DolarResponse, ctx context.Context) error {
	db, err := openDb()
	if err != nil {
		fmt.Println("Erro ao carregar DB:", err)
		return err
	}
	defer db.Close()

	err = createTable(db, ctx)
	if err != nil {
		fmt.Println("Erro ao criar table:", err)
		return err
	}

	// Inserindo os dados na tabela
	_, err = db.ExecContext(ctx, "INSERT INTO cotacao (value, date) VALUES (?, ?)",
		data.Usdbrl.Bid, time.Now().Format(time.RFC3339))
	if err != nil {
		fmt.Println("Erro ao realizar insert:", err)
		return err
	}

	return nil
}

func createTable(db *sql.DB, ctx context.Context) error {
	// Criando a tabela se ela não existir
	_, err := db.ExecContext(ctx,
		`
	CREATE TABLE IF NOT EXISTS cotacao (
		value TEXT,
		date DATE
	)
`)
	if err != nil {

		return err
	}
	return nil
}

func initializeSqLIte() error {

	dbPath := "./db/main.db"
	//Verify if DB exists

	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		fmt.Println("Database not found creating...")
		//Create the databa file and directory
		err = os.Mkdir("./db", os.ModePerm)
		if err != nil {
			return err
		}

		fmt.Println("Database file not found creating...")
		file, err := os.Create(dbPath)
		if err != nil {
			return err
		}
		file.Close()
	}

	return nil
}

func sendResponse(w http.ResponseWriter, cotacao *DolarResponse) error {
	var apiResponse ApiResponse
	apiResponse.Bid = cotacao.Usdbrl.Bid
	value, err := json.Marshal(apiResponse)
	if err != nil {
		return err
	}
	w.Write([]byte(value))
	return nil
}

func sendError(w http.ResponseWriter, error error) {
	w.Write([]byte(error.Error()))
}

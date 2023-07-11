package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {

	cotacao, err := makeRequest()
	if err != nil {
		fmt.Println("Erro ao criar arquivo de cotação:", err)
		return
	}

	err = createCotacaoFile(cotacao)
	if err != nil {
		fmt.Println("Erro ao criar arquivo de cotação:", err)
		return
	}

	fmt.Println("Cotação gravada com sucesso")
}

func makeRequest() (*Cotacao, error) {

	resp, err := http.Get("http://localhost:8080/cotacao")
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		fmt.Println("Erro ao fazer parse do result:", err)
		return nil, err
	}

	return &cotacao, nil
}

func createCotacaoFile(cotacao *Cotacao) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Println("Erro ao criar arquivo de cotação:", err)
		return err
	}

	_, err = file.WriteString(fmt.Sprintf("Cotação atual: %v ", cotacao.Bid))
	if err != nil {
		fmt.Println("Erro ao gravar cotação resposta:", err)
		return err
	}
	return nil
}

package main

import (
	"fmt"
	"log"
	"time"

	conn "modulo/db/conn"
	model "modulo/db/model"

	_ "github.com/lib/pq"

	arquivo "modulo/src"
)

func main() {
	// Inicia o cronometro
	var inicioTempo time.Time = time.Now()

	// Chama a conexao com o banco de dados
	db, err := conn.IniciarConexao()
	if err != nil {
		log.Fatalf("Falha ao conectar ao Postgre: %v", err)
	}
	defer db.Close()

	// Cria a tabela no banco de dados
	err = model.CriarTabela(db)
	if err != nil {
		log.Fatalf("Erro ao criar a tabela: %v", err)
	}

	arquivo.ArquivoLido()

	var tempoFinal time.Time = time.Now()

	var tempoTotal time.Duration = tempoFinal.Sub(inicioTempo)

	fmt.Println("Tempo total da insercao: ", tempoTotal)
}

package main

import (
	"fmt"
	"log"
	"time"

	conn "modulo/db/conn"
	model "modulo/db/model"
	arquivo "modulo/src/arquivo"
)

func main() {
	// Inicia o cronômetro
	var inicioTempo time.Time = time.Now()

	// Conectar ao banco de dados PostgreSQL
	db, err := conn.IniciarConexao()
	if err != nil {
		log.Fatalf("Falha ao conectar ao PostgreSQL: %v", err)
	}
	defer db.Close()

	// Criar a tabela 'pessoas' se não existir
	err = model.CriarTabela(db)
	if err != nil {
		log.Fatalf("Erro ao criar a tabela pessoas: %v", err)
	}

	// Ler dados do arquivo e inserir no banco de dados
	err = arquivo.LeituraDados(db)
	if err != nil {
		log.Fatalf("Erro ao ler dados do arquivo e inserir no banco de dados: %v", err)
	}

	// Medir tempo total de execução
	var tempoFinal time.Time = time.Now()
	var tempoTotal time.Duration = tempoFinal.Sub(inicioTempo)
	fmt.Println("Tempo total da inserção:", tempoTotal)
}

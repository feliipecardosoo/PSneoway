package main

import (
	"fmt"
	"log"
	"time"

	conn "modulo/db/conn"

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

	arquivo.ArquivoLido()

	var tempoFinal time.Time = time.Now()

	var tempoTotal time.Duration = tempoFinal.Sub(inicioTempo)

	fmt.Println("Tempo total da insercao: ", tempoTotal)
}

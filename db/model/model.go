package model

import (
	"database/sql"
	"log"
)

// CriarTabela cria a tabela pessoas no banco de dados se ela não existir
func CriarTabela(db *sql.DB) error {
	// Verifica se a tabela pessoas existe
	if tabelaExiste(db, "pessoas") {
		// Se existir, remove a tabela
		if err := dropTabela(db, "pessoas"); err != nil {
			return err
		}
	}

	// Agora cria a tabela pessoas
	criarTabelaQuery := `
	CREATE TABLE pessoas (
		cpf VARCHAR(20) PRIMARY KEY,
		private BOOLEAN,
		incompleto BOOLEAN,
		data_ultima_compra DATE,
		ticket_medio DECIMAL(10,2),
		ticket_ultima_compra DECIMAL(10,2),
		loja_mais_frequente VARCHAR(255),
		loja_ultima_compra VARCHAR(255)
	);`

	_, err := db.Exec(criarTabelaQuery)
	if err != nil {
		log.Fatalf("Erro ao criar a tabela pessoas: %v", err)
		return err
	}
	log.Println("Tabela pessoas criada ou já existe.")
	return nil
}

// Função para verificar se a tabela existe
func tabelaExiste(db *sql.DB, tableName string) bool {
	query := "SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1;"
	var result string
	err := db.QueryRow(query, tableName).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Fatalf("Erro ao verificar a existência da tabela %s: %v", tableName, err)
		return false
	}
	return true
}

// Função para dropar a tabela
func dropTabela(db *sql.DB, tableName string) error {
	query := "DROP TABLE " + tableName + ";"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Erro ao dropar a tabela %s: %v", tableName, err)
		return err
	}
	log.Printf("Tabela %s dropada com sucesso.", tableName)
	return nil
}

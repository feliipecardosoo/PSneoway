package model

import (
	"database/sql"
	"log"
)

// CriarTabela cria a tabela pessoas no banco de dados se ela não existir
func CriarTabela(db *sql.DB) error {
	criarTabelaQuery := `
	CREATE TABLE IF NOT EXISTS pessoas (
		cpf VARCHAR(14) PRIMARY KEY,
		cpf_valid BOOLEAN NOT NULL,
		private BOOLEAN NOT NULL,
		incompleto BOOLEAN NOT NULL,
		data_ultima_compra DATE,
		ticket_medio DECIMAL(10,2),
		ticket_ultima_compra DECIMAL(10,2),
		loja_mais_frequente VARCHAR(255),
		loja_mais_frequente_valid BOOLEAN NOT NULL,
		loja_ultima_compra VARCHAR(255),
		loja_ultima_compra_valid BOOLEAN NOT NULL
	);`

	_, err := db.Exec(criarTabelaQuery)
	if err != nil {
		log.Fatalf("Erro ao criar a tabela pessoas: %v", err)
		return err
	}
	log.Println("Tabela pessoas criada ou já existe.")
	return nil
}

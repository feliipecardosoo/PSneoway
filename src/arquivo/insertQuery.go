package arquivo

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// InsertBloco insere um bloco de dados na tabela 'pessoas'
func InsertBloco(tx *sql.Tx, valueStrings []string, valueArgs []interface{}) error {
	// Montando a query SQL de inserção
	stmt := fmt.Sprintf("INSERT INTO pessoas (cpf, private, incompleto, data_ultima_compra, ticket_medio, ticket_ultima_compra, loja_mais_frequente, loja_ultima_compra) VALUES %s",
		strings.Join(valueStrings, ","))

	// Executando a query
	_, err := tx.Exec(stmt, valueArgs...)
	if err != nil {
		log.Printf("Erro ao inserir bloco: %v\n", err)
		return err
	}
	return nil
}

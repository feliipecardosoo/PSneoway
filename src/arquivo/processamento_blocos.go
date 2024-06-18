package arquivo

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
)

const maxLinhasPorBloco = 1000

// ProcessarBlocos executa o processamento de blocos de linhas e os insere no banco de dados
func ProcessarBlocos(db *sql.DB, blocoCh <-chan []string) {
	var wg sync.WaitGroup

	for bloco := range blocoCh {
		wg.Add(1)
		go func(bloco []string) {
			defer wg.Done()
			executarProcessamentoDeBloco(db, bloco)
		}(bloco)
	}

	wg.Wait()
}

func executarProcessamentoDeBloco(db *sql.DB, bloco []string) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Erro ao iniciar transação para bloco: %v\n", err)
		return
	}
	defer func() {
		if err != nil {
			log.Printf("Rollback da transação para bloco: %v\n", err)
			tx.Rollback()
			return
		}
		err = tx.Commit()
		if err != nil {
			log.Printf("Erro ao commitar transação para bloco: %v\n", err)
			tx.Rollback()
		}
	}()

	valueStrings := make([]string, 0, len(bloco))
	valueArgs := make([]interface{}, 0, len(bloco)*8)

	for _, linha := range bloco {
		linha = strings.ReplaceAll(linha, ",", ".")

		campos := strings.FieldsFunc(linha, func(r rune) bool {
			return r == ',' || r == ' '
		})

		if len(campos) < 8 {
			log.Printf("Linha inválida: %s\n", linha)
			continue
		}

		cpfSemPontuacao := removerPontuacao(campos[0])
		if !validarCPF(cpfSemPontuacao) {
			log.Printf("CPF inválido: %s\n", campos[0])
			continue
		}

		cnpjMaisFrequenteSemPontuacao := removerPontuacao(campos[6])
		cnpjUltimaCompraSemPontuacao := removerPontuacao(campos[7])
		if !validarCNPJ(cnpjMaisFrequenteSemPontuacao) || !validarCNPJ(cnpjUltimaCompraSemPontuacao) {
			log.Printf("CNPJ inválido: %s ou %s\n", campos[6], campos[7])
			continue
		}

		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			cpfSemPontuacao,
			campos[1] == "1",
			campos[2] == "1",
			converterStringNula(campos[3]),
			converterFloatNulo(campos[4]),
			converterFloatNulo(campos[5]),
			cnpjMaisFrequenteSemPontuacao,
			cnpjUltimaCompraSemPontuacao,
		)
	}

	stmt := fmt.Sprintf("INSERT INTO pessoas (cpf, private, incompleto, data_ultima_compra, ticket_medio, ticket_ultima_compra, loja_mais_frequente, loja_ultima_compra) VALUES %s",
		strings.Join(valueStrings, ","))

	_, err = tx.Exec(stmt, valueArgs...)
	if err != nil {
		log.Printf("Erro ao inserir bloco: %v\n", err)
		return
	}
}

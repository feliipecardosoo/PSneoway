package arquivo

import (
	"database/sql"
	"log"
	"strings"
	"sync"
)

const maxLinhasPorBloco = 1500

// ProcessarBlocos executa o processamento de blocos de linhas e os insere no banco de dados
func ProcessarBlocos(db *sql.DB, blocoCh <-chan []string) {
	// Cria um WaitGroup para aguardar a conclusão de todas as goroutines
	var wg sync.WaitGroup

	// Iteracao sobre cada bloco de linhas recebido
	for bloco := range blocoCh {
		// Incrementa o contador do WaitGroup para indicar que uma nova goroutine está sendo iniciada
		wg.Add(1)

		// Inicia uma nova goroutine para processar o bloco atual
		go func(bloco []string) {
			// Decrementa o contador do WaitGroup quando a goroutine terminar, mesmo se ocorrer um erro
			defer wg.Done()

			// Chama a função executarProcessamentoDeBloco para processar e inserir o bloco no banco de dados
			executarProcessamentoDeBloco(db, bloco)
		}(bloco)
	}

	// Bloqueia a execução do programa até que todas as goroutines sejam concluídas
	wg.Wait()
}

// executarProcessamentoDeBloco é responsavel pelo processamento de um bloco lido
func executarProcessamentoDeBloco(db *sql.DB, bloco []string) {
	// Inicio da transacao
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

	// Valores de cada linha
	valueStrings := make([]string, 0, len(bloco))
	// Valores de cada coluna
	valueArgs := make([]interface{}, 0, len(bloco)*8)

	for _, linha := range bloco {
		// Estou trocando todas as , por .
		// Motivo: Coluna Ticket Médio e Ticket da ultima compra
		linha = strings.ReplaceAll(linha, ",", ".")

		// Divide a linha em campos, separa por virgula ou espaço
		campos := strings.FieldsFunc(linha, func(r rune) bool {
			return r == ',' || r == ' '
		})

		// Se a linha for menos que 8 campos é invalida
		if len(campos) < 8 {
			log.Printf("Linha inválida: %s\n", linha)
			continue
		}

		// Faz a remocao da pontuacao do CPF + Validacao
		cpfSemPontuacao := removerPontuacao(campos[0])
		if !validarCPF(cpfSemPontuacao) {
			log.Printf("CPF inválido: %s\n", campos[0])
			continue
		}

		// Faz a remocao da pontuacao do CNPJ + Validacao
		cnpjMaisFrequenteSemPontuacao := removerPontuacao(campos[6])
		cnpjUltimaCompraSemPontuacao := removerPontuacao(campos[7])
		if !validarCNPJ(cnpjMaisFrequenteSemPontuacao) || !validarCNPJ(cnpjUltimaCompraSemPontuacao) {
			log.Printf("CNPJ inválido: %s ou %s\n", campos[6], campos[7])
			continue
		}

		// Se passar por todas as validacoes os dados sao adicoinados a minha string Value Strings
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

	// Chama a função InsertBloco para inserir os dados no banco
	err = InsertBloco(tx, valueStrings, valueArgs)
	if err != nil {
		return
	}
}

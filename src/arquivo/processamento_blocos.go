package arquivo

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

const maxLinhasPorBloco = 1500

// ProcessarBlocos executa o processamento de blocos de linhas e os insere no banco de dados
func ProcessarBlocos(db *sql.DB, blocoCh <-chan []string) {
	// Cria um WaitGroup para aguardar a conclusão de todas as goroutines 34
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

		// Divide a linha em campos, separa por vírgula ou espaço
		campos := strings.FieldsFunc(linha, func(r rune) bool {
			return r == ',' || r == ' '
		})

		// Se a linha tiver menos que 8 campos é inválida
		if len(campos) < 8 {
			log.Printf("Linha inválida: %s\n", linha)
			continue
		}

		// Faz a remoção da pontuação do CPF + Validação
		cpfSemPontuacao := removerPontuacao(campos[0])
		if !validarCPF(cpfSemPontuacao) {
			log.Printf("CPF inválido: %s\n", campos[0])
			continue
		}

		// Faz a remoção da pontuação do CNPJ + Validação
		cnpjMaisFrequenteSemPontuacao := removerPontuacao(campos[6])
		cnpjUltimaCompraSemPontuacao := removerPontuacao(campos[7])
		if !validarCNPJ(cnpjMaisFrequenteSemPontuacao) || !validarCNPJ(cnpjUltimaCompraSemPontuacao) {
			log.Printf("CNPJ inválido: %s ou %s\n", campos[6], campos[7])
			continue
		}

		// Conversão e validação de campos float, substituindo por NULL se inválidos
		var camposFloat4, camposFloat5 string
		if campos[4] == "" {
			camposFloat4 = "NULL"
		} else {
			val, err := strconv.ParseFloat(campos[4], 8)
			if err != nil {
				camposFloat4 = "NULL"
			} else {
				camposFloat4 = fmt.Sprintf("%f", val)
			}
		}

		if campos[5] == "" {
			camposFloat5 = "NULL"
		} else {
			val, err := strconv.ParseFloat(campos[5], 8)
			if err != nil {
				camposFloat5 = "NULL"
			} else {
				camposFloat5 = fmt.Sprintf("%f", val)
			}
		}

		// Substituição de campos vazios por NULL
		nome := campos[3]
		if nome == "" {
			nome = "NULL"
		} else {
			nome = fmt.Sprintf("'%s'", nome)
		}

		// Tratamento do campo de data, substituindo por NULL se inválido ou vazio
		dataUltimaCompra := campos[3]
		if dataUltimaCompra == "" || dataUltimaCompra == "NULL" {
			dataUltimaCompra = "NULL"
		} else {
			dataUltimaCompra = fmt.Sprintf("'%s'", dataUltimaCompra)
		}

		// Se passar por todas as validações, os dados são adicionados à string valueStrings
		valueStrings = append(valueStrings, fmt.Sprintf("('%s', %t, %t, %s, %s, %s, '%s', '%s')",
			cpfSemPontuacao,
			campos[1] == "1",
			campos[2] == "1",
			dataUltimaCompra,
			camposFloat4,
			camposFloat5,
			cnpjMaisFrequenteSemPontuacao,
			cnpjUltimaCompraSemPontuacao))
	}

	// Chama a função InsertBloco para inserir os dados no banco
	err = InsertBloco(tx, valueStrings, valueArgs)
	if err != nil {
		return
	}
}

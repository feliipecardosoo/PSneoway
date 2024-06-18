package arquivo

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

// LeituraDados realiza a leitura do arquivo base_teste.txt e insere os dados no banco de dados
func LeituraDados(db *sql.DB) error {
	var caminhoArquivo string = "base_teste.txt"

	file, err := os.Open(caminhoArquivo)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo: %v", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	if scanner.Scan() {
		_ = scanner.Text()
	}

	const maxLinhasPorBloco = 1000
	blocoCh := make(chan []string, 10)

	removerPontuacao := func(s string) string {
		return strings.Map(func(r rune) rune {
			if r == '.' || r == '-' || r == '/' {
				return -1
			}
			return r
		}, s)
	}

	validarCPF := func(cpf string) bool {
		if cpf == "NULL" {
			return true
		}
		cpf = removerPontuacao(cpf)
		if len(cpf) != 11 {
			return false
		}

		if cpf == "00000000000" || cpf == "11111111111" || cpf == "22222222222" ||
			cpf == "33333333333" || cpf == "44444444444" || cpf == "55555555555" ||
			cpf == "66666666666" || cpf == "77777777777" || cpf == "88888888888" ||
			cpf == "99999999999" {
			return false
		}

		var soma int
		for i := 0; i < 9; i++ {
			soma += int(cpf[i]-'0') * (10 - i)
		}
		resto := soma % 11
		digitoVerificador1 := 11 - resto
		if digitoVerificador1 >= 10 {
			digitoVerificador1 = 0
		}

		soma = 0
		for i := 0; i < 10; i++ {
			soma += int(cpf[i]-'0') * (11 - i)
		}
		resto = soma % 11
		digitoVerificador2 := 11 - resto
		if digitoVerificador2 >= 10 {
			digitoVerificador2 = 0
		}

		return int(cpf[9]-'0') == digitoVerificador1 && int(cpf[10]-'0') == digitoVerificador2
	}

	validarCNPJ := func(cnpj string) bool {
		if cnpj == "NULL" {
			return true
		}
		cnpj = removerPontuacao(cnpj)
		if len(cnpj) != 14 {
			return false
		}
		re := regexp.MustCompile(`\d{14}`)
		return re.MatchString(cnpj)
	}

	processarBloco := func(bloco []string) {
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
				parseNullString(campos[3]),
				parseNullFloat(campos[4]),
				parseNullFloat(campos[5]),
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

	var wg sync.WaitGroup

	go func() {
		defer close(blocoCh)

		var bloco []string
		for scanner.Scan() {
			linha := scanner.Text()
			bloco = append(bloco, linha)

			if len(bloco) >= maxLinhasPorBloco {
				blocoCh <- bloco
				bloco = nil
			}
		}

		if len(bloco) > 0 {
			blocoCh <- bloco
		}
	}()

	for bloco := range blocoCh {
		wg.Add(1)
		go func(bloco []string) {
			defer wg.Done()
			processarBloco(bloco)
		}(bloco)
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro ao ler o arquivo: %v\n", err)
		return err
	}

	log.Println("Inserção de dados concluída.")
	return nil
}

func parseNullFloat(s string) sql.NullFloat64 {
	var nf sql.NullFloat64
	if s != "NULL" {
		var f float64
		_, err := fmt.Sscanf(s, "%f", &f)
		if err != nil {
			log.Printf("Erro ao converter para NullFloat64: %v\n", err)
			return nf
		}
		nf.Float64 = f
		nf.Valid = true
	}
	return nf
}

func parseNullString(s string) sql.NullString {
	if s == "NULL" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

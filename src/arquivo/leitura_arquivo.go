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
	// Caminho do arquivo a ser lido
	var caminhoArquivo string = "base_teste.txt"

	// Abrir o arquivo para leitura
	file, err := os.Open(caminhoArquivo)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo: %v", err)
		return err
	}
	defer file.Close()

	// Criar um scanner com buffering para leitura eficiente
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024) // Buffer de 64 KB
	scanner.Buffer(buf, 1024*1024)  // Tamanho máximo do buffer de leitura (1 MB)

	// Ignorar a primeira linha se necessário
	if scanner.Scan() {
		_ = scanner.Text() // Ignorar a primeira linha
	}

	// Número máximo de linhas por bloco
	const maxLinhasPorBloco = 1000

	// Canal para enviar blocos de linhas processadas
	blocoCh := make(chan []string, 10) // Buffer de canal maior

	// Função para remover pontuação de CPF ou CNPJ
	removerPontuacao := func(s string) string {
		return strings.Map(func(r rune) rune {
			if r == '.' || r == '-' || r == '/' {
				return -1
			}
			return r
		}, s)
	}

	// Função para validar CPF, incluindo valores NULL
	validarCPF := func(cpf string) bool {
		if cpf == "NULL" {
			return true
		}
		cpf = removerPontuacao(cpf)
		if len(cpf) != 11 {
			return false
		}

		// Verificar se todos os dígitos são iguais
		if cpf == "00000000000" || cpf == "11111111111" || cpf == "22222222222" ||
			cpf == "33333333333" || cpf == "44444444444" || cpf == "55555555555" ||
			cpf == "66666666666" || cpf == "77777777777" || cpf == "88888888888" ||
			cpf == "99999999999" {
			return false
		}

		// Primeiro dígito verificador
		var soma int
		for i := 0; i < 9; i++ {
			soma += int(cpf[i]-'0') * (10 - i)
		}
		resto := soma % 11
		digitoVerificador1 := 11 - resto
		if digitoVerificador1 >= 10 {
			digitoVerificador1 = 0
		}

		// Segundo dígito verificador
		soma = 0
		for i := 0; i < 10; i++ {
			soma += int(cpf[i]-'0') * (11 - i)
		}
		resto = soma % 11
		digitoVerificador2 := 11 - resto
		if digitoVerificador2 >= 10 {
			digitoVerificador2 = 0
		}

		// Comparar com os dígitos verificadores informados
		return int(cpf[9]-'0') == digitoVerificador1 && int(cpf[10]-'0') == digitoVerificador2
	}

	// Função para validar CNPJ, incluindo valores NULL
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

	// Função para processar um bloco de linhas
	processarBloco := func(bloco []string) {
		// Iniciar transação para o bloco de linhas
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

		// Preparar statement de inserção
		stmt, err := tx.Prepare(`
		INSERT INTO pessoas (cpf, private, incompleto, data_ultima_compra, ticket_medio, ticket_ultima_compra, loja_mais_frequente, loja_ultima_compra)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`)
		if err != nil {
			log.Printf("Erro ao preparar statement: %v\n", err)
			return
		}
		defer stmt.Close()

		// Processar e inserir cada linha no batch
		for _, linha := range bloco {
			// Substituir vírgulas por pontos
			linha = strings.ReplaceAll(linha, ",", ".")

			// Dividir a linha por vírgulas, removendo espaços extras
			campos := strings.FieldsFunc(linha, func(r rune) bool {
				return r == ',' || r == ' '
			})

			// Verificar se há pelo menos 8 campos (conforme o exemplo fornecido)
			if len(campos) < 8 {
				log.Printf("Linha inválida: %s\n", linha)
				continue
			}

			// Remover pontuação e validar CPF
			cpfSemPontuacao := removerPontuacao(campos[0])
			if !validarCPF(cpfSemPontuacao) {
				log.Printf("CPF inválido: %s\n", campos[0])
				continue
			}

			// Remover pontuação e validar CNPJ (exemplo para campos 6 e 7)
			cnpjMaisFrequenteSemPontuacao := removerPontuacao(campos[6])
			cnpjUltimaCompraSemPontuacao := removerPontuacao(campos[7])
			if !validarCNPJ(cnpjMaisFrequenteSemPontuacao) || !validarCNPJ(cnpjUltimaCompraSemPontuacao) {
				log.Printf("CNPJ inválido: %s ou %s\n", campos[6], campos[7])
				continue
			}

			// Criar um novo comprador com os dados da linha atual
			comprador := Comprador{
				CPF:                cpfSemPontuacao,
				Private:            campos[1] == "1",
				Incompleto:         campos[2] == "1",
				DataUltimaCompra:   sql.NullString{String: campos[3], Valid: campos[3] != "NULL"},
				TicketMedio:        parseNullFloat(campos[4]),
				TicketUltimaCompra: parseNullFloat(campos[5]),
				LojaMaisFrequente:  cnpjMaisFrequenteSemPontuacao,
				LojaUltimaCompra:   cnpjUltimaCompraSemPontuacao,
			}

			// Executar a query preparada com os valores do comprador
			_, err := stmt.Exec(
				comprador.CPF,
				comprador.Private,
				comprador.Incompleto,
				comprador.DataUltimaCompra,
				comprador.TicketMedio,
				comprador.TicketUltimaCompra,
				comprador.LojaMaisFrequente,
				comprador.LojaUltimaCompra,
			)
			if err != nil {
				log.Printf("Erro ao inserir comprador: %v\n", err)
				continue
			}
		}
	}

	// Usar um canal para esperar que todas as goroutines terminem
	var wg sync.WaitGroup

	// Iniciar goroutine para processar blocos
	go func() {
		defer close(blocoCh)

		var bloco []string
		for scanner.Scan() {
			linha := scanner.Text()

			// Adicionar linha ao bloco
			bloco = append(bloco, linha)

			// Se o tamanho do bloco atingir o máximo, enviar o bloco para processamento e limpar
			if len(bloco) >= maxLinhasPorBloco {
				blocoCh <- bloco
				bloco = nil
			}
		}

		// Se houver linhas restantes no último bloco, enviar o último bloco
		if len(bloco) > 0 {
			blocoCh <- bloco
		}
	}()

	// Processar os blocos recebidos
	for bloco := range blocoCh {
		wg.Add(1)
		go func(bloco []string) {
			defer wg.Done()
			processarBloco(bloco)
		}(bloco)
	}

	// Esperar até que todas as goroutines tenham terminado
	wg.Wait()

	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro ao ler o arquivo: %v\n", err)
		return err
	}

	log.Println("Inserção de dados concluída.")
	return nil
}

// Função auxiliar para converter string para sql.NullFloat64
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

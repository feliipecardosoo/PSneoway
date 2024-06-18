// Arquivo arquivo.go

package arquivo

import (
	"bufio"
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// LeituraDados realiza a leitura do arquivo e manipula os dados para depois inserir no banco de dados
func LeituraDados(db *sql.DB) error {
	var caminhoArquivo string = "base_teste.txt"

	file, err := os.Open(caminhoArquivo)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo: %v", err)
		return err
	}
	defer file.Close()

	// Criando um scanner para ler arquivo.
	scanner := bufio.NewScanner(file)

	// Enviar blocos para processamento
	// Escolhi 20 blocos pois a minha leitura do arquivo esta rapida e o meu processamento esta mais lento.
	blocoCh := make(chan []string, 20)

	// Goroutine para ler o arquivo
	go func() {
		defer close(blocoCh)

		var bloco []string

		// Loop que percorre cada linha do arquivo usando o scanner
		for scanner.Scan() {
			linha := scanner.Text()
			bloco = append(bloco, linha)

			// Se o bloco atingir o tamanho máximo definido por maxLinhasPorBloco
			if len(bloco) >= maxLinhasPorBloco {
				// Envia o bloco completo para o canal blocoCh
				blocoCh <- bloco
				// Reseta o bloco para iniciar um novo
				bloco = nil
			}
		}

		// Após a leitura de todas as linhas, se ainda houver linhas no bloco
		if len(bloco) > 0 {
			// Envia o bloco restante para o canal blocoCh
			blocoCh <- bloco
		}
	}()

	// Processar blocos
	ProcessarBlocos(db, blocoCh)

	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro ao ler o arquivo: %v\n", err)
		return err
	}

	log.Println("Inserção de dados concluída.")
	return nil
}

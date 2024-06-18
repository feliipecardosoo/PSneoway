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

	// Criando um buffer capacidade 64kb
	buf := make([]byte, 0, 64*1024)

	// Definindo buffer interno do scanner para 1mb
	scanner.Buffer(buf, 1024*1024)

	if scanner.Scan() {
		_ = scanner.Text()
	}

	blocoCh := make(chan []string, 10)

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

	// Processar blocos
	ProcessarBlocos(db, blocoCh)

	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro ao ler o arquivo: %v\n", err)
		return err
	}

	log.Println("Inserção de dados concluída.")
	return nil
}

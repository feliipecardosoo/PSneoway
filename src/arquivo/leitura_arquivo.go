package arquivo

import (
	"fmt"
	"log"
	"os"
)

func LeituraArquivo() {
	// Caminho do arquivo a ser lido
	var caminhoArquivo string = "base_teste.txt"

	// Leitura do arquivo selecionado
	file, err := os.Open(caminhoArquivo)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo: %v", err)
	}

	// Lendo os dados na fatia de bytes de 50k
	var quantidadeByte uint = 10000
	dados := make([]byte, quantidadeByte)
	count, err := file.Read(dados)
	if err != nil {
		log.Fatalf("Erro ao processar arquivo em byte: %v", err)
	}
	fmt.Println(count)
	defer file.Close()
}

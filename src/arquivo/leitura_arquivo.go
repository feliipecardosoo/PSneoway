package arquivo

import (
	"log"
	"os"
)

func LeituraPrimeiraLinha() {
	// Caminho do arquivo a ser lido
	var caminhoArquivo string = "base_teste.txt"

	// Leitura do arquivo selecionado
	file, err := os.Open(caminhoArquivo)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	// Lendo os dados na fatia de bytes de 50k
	var quantidadeByte uint = 10000
	dados := make([]byte, quantidadeByte)
	var n int
	n, err = file.Read(dados)
	if err != nil {
		log.Fatalf("Erro ao processar arquivo em byte: %v", err)
	}

	// n é necessário para evitar a variável anônima, mas não será usada
	_ = n
}

# PS Neoway

Este é um projeto desenvolvido como parte do Desafio Técnico da Neoway.

## Como executar

Para executar este projeto localmente, é necessário ter o Docker instalado.

1. Clone este repositório:


2. Execute o seguinte comando para iniciar o projeto:

```bash
docker-compose up
```


Este comando iniciará os containers necessários para executar a aplicação.

## Conexão com o Banco de Dados

Para conectar no banco de dados, utilize as seguintes configurações:

- **DB**: neoway
- **Usuário**: root
- **Senha**: 3103
- **Tabela**: pessoas

Certifique-se de que o container do banco de dados está em execução após o comando `docker-compose up`.

## Estratégia 

- **Leitura por Blocos:**
O arquivo é lido e armazenado em blocos com um máximo de 1500 linhas por bloco.
Utiliza a biblioteca padrão bufio para eficiência na leitura linha a linha.
O processo de leitura é feito de forma assíncrona usando goroutines para melhorar o desempenho.

- **Inserção no Banco de Dados:**
Após a leitura de cada bloco, os dados são enviados para uma função ProcessarBlocos que os manipula e insere no banco de dados.
Usa um canal (blocoCh) para comunicação assíncrona entre a leitura do arquivo e o processamento dos blocos.
É esperado que o banco de dados (db) seja passado como parâmetro para a função LeituraDados.
##

**Função ProcessarBlocos:**
Coordena o processamento dos blocos de linhas lidos do arquivo. Ela utiliza canais para receber os blocos de linhas (blocoCh) e inicia goroutines para processar cada bloco de forma assíncrona.

```bash
package arquivo

import (
	"database/sql"
	"log"
	"sync"
)

const maxLinhasPorBloco = 1500

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
```
##
**Função executarProcessamentoDeBloco:**
É responsável por processar cada bloco de linhas recebido pela função ProcessarBlocos. Ela realiza as seguintes operações:

- **Inicia uma transação no banco de dados para garantir a atomicidade das operações.**

- **Para cada linha no bloco, realiza as validações necessárias nos dados (CPF, CNPJ) antes de prepará-los para inserção.**

- **Prepara os valores a serem inseridos no banco de dados, formatando-os como strings e definindo os argumentos correspondentes.**

- **Utiliza a função InsertBloco para inserir os dados na transação atual.**


```bash
// executarProcessamentoDeBloco é responsável pelo processamento de um bloco lido
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
		// ... validações e preparação dos dados ...

		// Adiciona os valores formatados ao slice de valueStrings e valueArgs
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

	// Insere os dados no banco de dados usando a transação atual (tx)
	err = InsertBloco(tx, valueStrings, valueArgs)
	if err != nil {
		return
	}
}
```
##
**Observações:**

- **Atomicidade:** A transação é utilizada para garantir que todas as operações de inserção sejam realizadas com sucesso ou nenhuma delas seja aplicada, garantindo a integridade dos dados.

- **Validações:** São realizadas validações nos CPFs e CNPJs para assegurar que apenas dados válidos sejam inseridos no banco de dados.

- **Paralelismo:** O uso de goroutines permite processar múltiplos blocos de forma simultânea, otimizando o desempenho do processamento.

##
## Estrutura da Table:
```bash
CREATE TABLE pessoas (
		cpf VARCHAR(20) PRIMARY KEY,
		private BOOLEAN NOT NULL,
		incompleto BOOLEAN NOT NULL,
		data_ultima_compra DATE,
		ticket_medio DECIMAL(10,2),
		ticket_ultima_compra DECIMAL(10,2),
		loja_mais_frequente VARCHAR(255),
		loja_ultima_compra VARCHAR(255)
	);`
```
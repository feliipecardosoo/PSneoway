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

- Inicia uma transação no banco de dados para garantir a atomicidade das operações.

- Para cada linha no bloco, realiza as validações necessárias nos dados (CPF, CNPJ) antes de prepará-los para inserção.

- Prepara os valores a serem inseridos no banco de dados, formatando-os como strings e definindo os argumentos correspondentes.

- Utiliza a função InsertBloco para inserir os dados na transação atual.

```bash
// executarProcessamentoDeBloco é responsável pelo processamento de um bloco lido
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
		private BOOLEAN,
		incompleto BOOLEAN,
		data_ultima_compra DATE,
		ticket_medio DECIMAL(10,2),
		ticket_ultima_compra DECIMAL(10,2),
		loja_mais_frequente VARCHAR(255),
		loja_ultima_compra VARCHAR(255)
	);`
```
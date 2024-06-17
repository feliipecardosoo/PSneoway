package conn

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Credenciais armazena as informações de conexão com o banco de dados
type Credenciais struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// NovaCredencial inicializa um novo struct Credentials a partir das variáveis de ambiente
func NovaCredencial() *Credenciais {
	return &Credenciais{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}
}

// NovaConexao cria uma nova conexão com o banco de dados PostgreSQL
func NovaConexao(creds *Credenciais) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		creds.DBHost, creds.DBPort, creds.DBUser, creds.DBPassword, creds.DBName)

	var db *sql.DB
	var err error
	maxAttempts := 10
	attempt := 0

	// Loop para tentar conectar ao PostgreSQL várias vezes
	for attempt < maxAttempts {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			break
		}
		log.Printf("Erro ao conectar ao PostgreSQL (tentativa %d de %d): %v", attempt+1, maxAttempts, err)
		attempt++
		time.Sleep(5 * time.Second) // Espera 5 segundos antes de tentar novamente
	}

	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao PostgreSQL após %d tentativas: %v", maxAttempts, err)
	}

	// Verificar a conexão
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("erro ao pingar o PostgreSQL: %v", err)
	}

	fmt.Println("Conectado ao PostgreSQL!")
	return db, nil
}

// IniciarConexao inicializa as credenciais e cria a conexao com o banco de dados
func IniciarConexao() (*sql.DB, error) {
	creds := NovaCredencial()
	db, err := NovaConexao(creds)
	if err != nil {
		return nil, err
	}
	fmt.Println("Aplicação iniciada com sucesso!")
	return db, nil
}

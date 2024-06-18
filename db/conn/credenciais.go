package conn

// Credenciais armazena as informações de conexão com o banco de dados
type Credenciais struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

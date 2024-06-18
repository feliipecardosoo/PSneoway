package arquivo

import (
	"database/sql"
)

// Comprador representa os dados de um comprador conforme o arquivo base_teste2.txt
type Comprador struct {
	CPF                string
	Private            bool
	Incompleto         bool
	DataUltimaCompra   sql.NullString
	TicketMedio        sql.NullFloat64
	TicketUltimaCompra sql.NullFloat64
	LojaMaisFrequente  string
	LojaUltimaCompra   string
}

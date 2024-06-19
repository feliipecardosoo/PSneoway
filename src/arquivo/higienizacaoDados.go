package arquivo

import (
	"regexp"
	"strings"
)

// removerPontuacao é usado tanto no CPF quanto no CNPJ para tirar as pontuacoes.
func removerPontuacao(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '.' || r == '-' || r == '/' {
			return -1
		}
		return r
	}, s)
}

// validarCPF verifica se o CPF é valido, mecanismos: comprimento(11), digitos iguais e calculo do primeiro e segundo verificador.
func validarCPF(cpf string) bool {
	if cpf == "NULL" {
		return true
	}
	cpf = removerPontuacao(cpf)
	if len(cpf) != 11 {
		return false
	}

	if cpf == "00000000000" || cpf == "11111111111" || cpf == "22222222222" ||
		cpf == "33333333333" || cpf == "44444444444" || cpf == "55555555555" ||
		cpf == "66666666666" || cpf == "77777777777" || cpf == "88888888888" ||
		cpf == "99999999999" {
		return false
	}

	var soma int
	for i := 0; i < 9; i++ {
		soma += int(cpf[i]-'0') * (10 - i)
	}
	resto := soma % 11
	digitoVerificador1 := 11 - resto
	if digitoVerificador1 >= 10 {
		digitoVerificador1 = 0
	}

	soma = 0
	for i := 0; i < 10; i++ {
		soma += int(cpf[i]-'0') * (11 - i)
	}
	resto = soma % 11
	digitoVerificador2 := 11 - resto
	if digitoVerificador2 >= 10 {
		digitoVerificador2 = 0
	}

	return int(cpf[9]-'0') == digitoVerificador1 && int(cpf[10]-'0') == digitoVerificador2
}

// validarCNPJ verifica se o CNPJ é valido, mecanismo: compimento(14)
func validarCNPJ(cnpj string) bool {
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

// converterStringNula se a String for "NULL", retorna um sql.NullString, caso contrario, vai cair no segundo return, aonde contem a String de entrada e o Valid é definido como true
// vou usar isso no campo de Date.
// func converterStringNula(s string) sql.NullString {
// 	if s == "NULL" {
// 		return sql.NullString{}
// 	}
// 	return sql.NullString{
// 		String: s,
// 		Valid:  true,
// 	}
// }

// converterFloatNulo se a String de entrada for diferente de "NULL", tenta converter a string para float64
// vou usar isso no campo de tickets
// func converterFloatNulo(s string) sql.NullFloat64 {
// 	var nf sql.NullFloat64
// 	if s != "NULL" {
// 		var f float64
// 		_, err := fmt.Sscanf(s, "%f", &f)
// 		if err != nil {
// 			log.Printf("Erro ao converter para NullFloat64: %v\n", err)
// 			return nf
// 		}
// 		nf.Float64 = f
// 		nf.Valid = true
// 	}
// 	return nf
// }

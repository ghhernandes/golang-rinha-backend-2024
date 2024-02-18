package rinha

import "time"

type Transacao struct {
	ClienteId int       `json:"-"`
	Tipo      string    `json:"tipo"`
	Valor     int       `json:"valor"`
	Descricao string    `json:"descricao"`
	Data      time.Time `json:"realizada_em"`
}

type Saldo struct {
	Total       int       `json:"total"`
	DataExtrato time.Time `json:"data_extrato"`
	Limite      int       `json:"limite"`
}

type Extrato struct {
	Saldo     Saldo       `json:"saldo"`
	Historico []Transacao `json:"ultimas_transacoes"`
}

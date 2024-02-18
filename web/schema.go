package web

import "errors"

type TransacaoPostRequest struct {
	Valor     int    `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

func (t TransacaoPostRequest) Validate() error {
	if t.Tipo != "d" && t.Tipo != "c" {
		return errors.New("invalid tipo")
	}

	if t.Descricao == "" || len(t.Descricao) > 10 {
		return errors.New("invalid descricao")
	}

	if t.Valor == 0 {
		return errors.New("invalid valor")
	}
	return nil
}

type TransacaoPostResponse struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}

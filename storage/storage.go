package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghhernandes/golang-rinha-backend-2024"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(url string) (*Storage, error) {
	cfg, err := pgxpool.ParseConfig("host=/var/run/postgresql user=admin password=123 dbname=rinha")
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MinConns = 10

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &Storage{pool: pool}, nil
}

func (s Storage) Close() {
	s.pool.Close()
}

func (s Storage) CreateTransacao(ctx context.Context, t rinha.Transacao) (rinha.Saldo, error) {
	var (
		saldo      rinha.Saldo
		finalizado bool
	)

	conn, _ := s.pool.Acquire(ctx)
	defer conn.Release()

	conn.Exec(ctx, "SELECT pg_advisory_lock($1)", t.ClienteId)
	defer conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", t.ClienteId)

	if t.Tipo == "d" {
		t.Valor = t.Valor * -1
	}

	r := conn.QueryRow(
		context.Background(),
		"select limite_atual, novo_saldo, finalizado from create_transacao($1, $2, $3)",
		t.ClienteId, t.Valor, t.Descricao,
	)

	if err := r.Scan(&saldo.Limite, &saldo.Total, &finalizado); err != nil {
		return saldo, fmt.Errorf("create transaction: scan: %v", err)
	}

	if !finalizado {
		return saldo, errors.New("invalid operation")
	}

	return saldo, nil
}

func (s Storage) GetExtrato(ctx context.Context, clienteId int) (rinha.Extrato, error) {
	saldoCh := make(chan rinha.Saldo)
	errCh := make(chan error)
	go func() {
		saldo, err := s.getSaldo(ctx, clienteId)
		if err != nil {
			errCh <- err
		}
		saldoCh <- saldo
	}()

	rows, err := s.pool.Query(context.Background(), `
        select cliente_id, valor, descricao, data 
        from transacoes 
        where cliente_id = $1 
        order by data desc 
        limit 10`,
		clienteId,
	)
	if err != nil {
		return rinha.Extrato{}, err
	}

	var t rinha.Transacao
	var e rinha.Extrato

	_, err = pgx.ForEachRow(rows, []any{&t.ClienteId, &t.Valor, &t.Descricao, &t.Data}, func() error {
		t.Tipo, t.Valor = decodeTransacaoTipoValor(t.Valor)
		e.Historico = append(e.Historico, t)
		return nil
	})

	select {
	case saldo := <-saldoCh:
		e.Saldo = saldo
	case e := <-errCh:
		err = e
	case <-ctx.Done():
		return e, ctx.Err()
	}
	return e, err
}

func (s Storage) getSaldo(ctx context.Context, clienteId int) (rinha.Saldo, error) {
	var total, limite int
	row := s.pool.QueryRow(context.Background(),
		"select saldo, limite from clientes where id = $1",
		clienteId,
	)

	if err := row.Scan(&total, &limite); err != nil {
		return rinha.Saldo{}, nil
	}
	return rinha.Saldo{Total: total, Limite: limite, DataExtrato: time.Now().UTC()}, nil
}

func decodeTransacaoTipoValor(valor int) (string, int) {
	if valor < 0 {
		return "d", valor * -1
	}
	return "c", valor
}

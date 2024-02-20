package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	rinha "github.com/ghhernandes/golang-rinha-backend-2024"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	CREATE_TRANSACTION = "INSERT INTO transacoes(cliente_id, valor, descricao) values ($1, $2, $3)"
	UPDATE_SALDO       = "UPDATE clientes set saldo = saldo + $2 where id = $1"
	GET_SALDO          = "SELECT limite, saldo from clientes where id = $1"
	GET_EXTRATO        = "SELECT cliente_id, valor, descricao, data from transacoes where cliente_id = $1 order by data desc limit 10"
	GET_LOCK           = "SELECT pg_advisory_xact_lock($1)"
)

type Config struct {
	Host         string
	User         string
	Password     string
	DatabaseName string

	MinConns int32
	MaxConns int32
}

type Storage struct {
	pool *pgxpool.Pool
}

func New(config *Config) (*Storage, error) {
	if config == nil {
		return nil, errors.New("storage config not set")
	}

	cfg, err := pgxpool.ParseConfig(fmt.Sprintf("host=%s user=%s password=%s dbname=%s",
		config.Host,
		config.User,
		config.Password,
		config.DatabaseName,
	))

	if err != nil {
		return nil, err
	}

	cfg.MaxConns = config.MaxConns
	cfg.MinConns = config.MinConns

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
	var saldo rinha.Saldo

	if t.Tipo == "d" {
		t.Valor = t.Valor * -1
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return rinha.Saldo{}, err
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, GET_LOCK, t.ClienteId)

	q := tx.QueryRow(ctx, GET_SALDO, t.ClienteId)

	if err := q.Scan(&saldo.Limite, &saldo.Total); err != nil {
		return rinha.Saldo{}, err
	}

	if t.Valor < 0 && saldo.Total+t.Valor < (saldo.Limite*-1) {
		return rinha.Saldo{}, err
	}

	var batch pgx.Batch

	batch.Queue(CREATE_TRANSACTION, t.ClienteId, t.Valor, t.Descricao)
	batch.Queue(UPDATE_SALDO, t.ClienteId, t.Valor)

	r := tx.SendBatch(ctx, &batch)

	if err := r.Close(); err != nil {
		return rinha.Saldo{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return rinha.Saldo{}, err
	}

	saldo.Total += t.Valor

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

	rows, err := s.pool.Query(ctx, GET_EXTRATO, clienteId)
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
	row := s.pool.QueryRow(context.Background(), GET_SALDO, clienteId)

	if err := row.Scan(&limite, &total); err != nil {
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

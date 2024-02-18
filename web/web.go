package web

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/ghhernandes/rinha-backend-2024-q1"
	"github.com/ghhernandes/rinha-backend-2024-q1/storage"
)

type Handler struct {
	router  *http.ServeMux
	storage *storage.Storage

	quitCh chan struct{}
}

func New(storage *storage.Storage) *Handler {
	r := http.NewServeMux()

	r.HandleFunc("POST /clientes/{cliente_id}/transacoes", transacoesHandler(storage))
	r.HandleFunc("GET /clientes/{cliente_id}/extrato", extratoHandler(storage))

	return &Handler{
		router:  r,
		storage: storage,
		quitCh:  make(chan struct{}),
	}
}

func (h Handler) Listen() <-chan struct{} {
	go func() {
		defer close(h.quitCh)
		if err := http.ListenAndServe(":8080", h.router); err != nil {
			log.Println("ListenAndServe error:", err)
		}
	}()
	return h.quitCh
}

func (h Handler) Serve(l net.Listener) <-chan struct{} {
	go func() {
		defer close(h.quitCh)
		if err := http.Serve(l, h.router); err != nil {
			log.Println("Serve error:", err)
		}
	}()
	return h.quitCh
}

func transacoesHandler(storage *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TransacaoPostRequest
		ctx := context.Background()

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		clienteId, err := getClienteId(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		saldo, err := storage.CreateTransacao(ctx, rinha.Transacao{
			ClienteId: clienteId,
			Tipo:      req.Tipo,
			Valor:     req.Valor,
			Descricao: req.Descricao,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		response, _ := json.Marshal(TransacaoPostResponse{
			Limite: saldo.Limite,
			Saldo:  saldo.Total,
		})
		w.Write(response)
	}
}

func extratoHandler(storage *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clienteId, err := getClienteId(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		e, err := storage.GetExtrato(context.Background(), clienteId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		response, _ := json.Marshal(e)
		w.Write(response)
	}
}

func getClienteId(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("cliente_id"))
	if id < 0 || id > 5 {
		return -1, errors.New("cliente not found")
	}
	return id, err
}

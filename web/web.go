package web

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"

	rinha "github.com/ghhernandes/golang-rinha-backend-2024"
	"github.com/ghhernandes/golang-rinha-backend-2024/storage"
	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	router  *httprouter.Router
	storage *storage.Storage

	quitCh chan struct{}
}

func New(storage *storage.Storage) *Handler {
	r := httprouter.New()

	r.Handler(http.MethodPost, "/clientes/1/transacoes", transacoesHandler(storage, 1))
	r.Handler(http.MethodPost, "/clientes/2/transacoes", transacoesHandler(storage, 2))
	r.Handler(http.MethodPost, "/clientes/3/transacoes", transacoesHandler(storage, 3))
	r.Handler(http.MethodPost, "/clientes/4/transacoes", transacoesHandler(storage, 4))
	r.Handler(http.MethodPost, "/clientes/5/transacoes", transacoesHandler(storage, 5))

	r.Handler(http.MethodGet, "/clientes/1/extrato", extratoHandler(storage, 1))
	r.Handler(http.MethodGet, "/clientes/2/extrato", extratoHandler(storage, 2))
	r.Handler(http.MethodGet, "/clientes/3/extrato", extratoHandler(storage, 3))
	r.Handler(http.MethodGet, "/clientes/4/extrato", extratoHandler(storage, 4))
	r.Handler(http.MethodGet, "/clientes/5/extrato", extratoHandler(storage, 5))

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

func transacoesHandler(storage *storage.Storage, clienteId int) http.HandlerFunc {
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

func extratoHandler(storage *storage.Storage, clienteId int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		e, err := storage.GetExtrato(context.Background(), clienteId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		response, _ := json.Marshal(e)
		w.Write(response)
	}
}

package socket

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghhernandes/golang-rinha-backend-2024/web"
)

type Handler struct {
	addr    string
	handler *web.Handler

	quitCh   chan struct{}
	signalCh chan os.Signal

	socket net.Listener
}

func New(addr string, handler *web.Handler) *Handler {
	return &Handler{
		addr:    addr,
		handler: handler,

		quitCh:   make(chan struct{}),
		signalCh: make(chan os.Signal, 1),
	}
}

func (h *Handler) Listen() <-chan struct{} {
	var err error

	log.Println("listening to", h.addr)

	h.socket, err = net.Listen("unix", h.addr)
	if err != nil {
		panic(err)
	}

	signal.Notify(h.signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-h.signalCh
		if err := h.socket.Close(); err != nil {
			fmt.Println("socket close error:", err)
		}
	}()

	go func() {
		defer close(h.quitCh)
		<-h.handler.Serve(h.socket)
		os.Remove(h.addr)
	}()

	return h.quitCh
}

package main

import (
	"os"

	"github.com/ghhernandes/rinha-backend-2024-q1/pyroscope"
	"github.com/ghhernandes/rinha-backend-2024-q1/socket"
	"github.com/ghhernandes/rinha-backend-2024-q1/storage"
	"github.com/ghhernandes/rinha-backend-2024-q1/web"
)

func main() {
	if os.Getenv("PYROSCOPE_ENABLED") == "1" {
		pyroscope.Start("ghhernandes.rinha", os.Getenv("PYROSCOPE_ADDR"), os.Getenv("HOSTNAME"))
	}

	db, err := storage.New(os.Getenv("DB_URI"))
	if err != nil {
		panic(err)
	}

	handler := socket.New(os.Getenv("SOCKET_ADDR"), web.New(db))
	<-handler.Listen()
}

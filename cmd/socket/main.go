package main

import (
	"os"

	"github.com/ghhernandes/golang-rinha-backend-2024/socket"
	"github.com/ghhernandes/golang-rinha-backend-2024/storage"
	"github.com/ghhernandes/golang-rinha-backend-2024/web"
)

func main() {
	db, err := storage.New(os.Getenv("DB_URI"))
	if err != nil {
		panic(err)
	}

	handler := socket.New(os.Getenv("SOCKET_ADDR"), web.New(db))
	<-handler.Listen()
}

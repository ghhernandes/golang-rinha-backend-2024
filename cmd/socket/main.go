package main

import (
	"os"

	"github.com/ghhernandes/golang-rinha-backend-2024/socket"
	"github.com/ghhernandes/golang-rinha-backend-2024/storage"
	"github.com/ghhernandes/golang-rinha-backend-2024/web"
)

func main() {
	db, err := storage.New(&storage.Config{
		Host:         os.Getenv("DB_HOST"),
		User:         os.Getenv("DB_USER"),
		Password:     os.Getenv("DB_PASSWORD"),
		DatabaseName: os.Getenv("DB_NAME"),

		MaxConns: 10,
		MinConns: 10,
	})
	if err != nil {
		panic(err)
	}

	handler := socket.New(os.Getenv("SOCKET_ADDR"), web.New(db))
	<-handler.Listen()
}

package main

import (
	"os"

	"github.com/ghhernandes/rinha-backend-2024-q1/storage"
	"github.com/ghhernandes/rinha-backend-2024-q1/web"
)

func main() {
	db, err := storage.New(os.Getenv("DB_URI"))
	if err != nil {
		panic(err)
	}

	handler := web.New(db)
	<-handler.Listen()
}

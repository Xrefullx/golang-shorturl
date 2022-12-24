package main

import (
	"github.com/Xrefullx/golang-shorturl/internal/app"
	"github.com/Xrefullx/golang-shorturl/internal/handlers"
	"github.com/Xrefullx/golang-shorturl/internal/server"
	"github.com/Xrefullx/golang-shorturl/internal/storage/memory"
	"log"
	"os"
	"os/signal"
)

func main() {
	var config app.ServerConfig
	app.EnviromentConfig(&config)

	db := memory.NewStorage()
	sUrl, err := app.NewShort(db)
	if err != nil {
		log.Fatal(err.Error())
	}
	handler := handlers.CreateHandler(sUrl, config.Url)
	server := server.Createserver(config.Port, *handler)
	log.Fatal(server.Run())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}

package main

import (
	"context"
	"github.com/Xrefullx/golang-shorturl/internal/api"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/Xrefullx/golang-shorturl/internal/storage/file"
	"github.com/Xrefullx/golang-shorturl/internal/storage/postgres"
	"github.com/Xrefullx/golang-shorturl/pkg"
	"log"
	"os"
	"os/signal"
)

func main() {

	cfg, err := pkg.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := getDB(*cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	server, err := api.NewServer(cfg, db)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Fatal(server.Run())

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	<-sigc

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("error shutdown server: %s\n", err.Error())
	}
}

func getDB(cfg pkg.Config) (storage.Storage, error) {
	if cfg.DatabaseDSN != "" {
		db, err := postgres.NewStorage(cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		return db, nil
	}
	db, err := file.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

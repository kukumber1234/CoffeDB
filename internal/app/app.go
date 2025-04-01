package app

import (
	"flag"
	"frappuccino/config"
	"frappuccino/internal/dal"
	"frappuccino/internal/routes"
	"log"
	"net/http"
)

func Start() {
	dal.Create()
	config.Logger.Info("Starting server...")

	flag.Parse()

	if *config.Help {
		dal.Helper()
		return
	}
	config.Logger.Info("Parsed flags")

	config.Logger.Info("Trying to connect to DB")
	db, err := dal.ConnectionDB()
	if err != nil {
		config.Logger.Error("Error trying to connect to DB", err)
		log.Fatal(err)
	}
	defer db.Close()
	config.Logger.Info("Connected to DB")

	mux := http.NewServeMux()

	routes.Routes(mux, db)
	config.Logger.Info("DataBase connection established")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

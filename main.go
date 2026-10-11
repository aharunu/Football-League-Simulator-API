package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"league-simulator/handlers"
	"league-simulator/models"
	"league-simulator/services"
)

func main() {
	teams := []models.Team{
		{ID: 1, Name: "Manchester United", Strength: 5},
		{ID: 2, Name: "Manchester City", Strength: 7},
		{ID: 3, Name: "Chelsea", Strength: 6},
		{ID: 4, Name: "Liverpool", Strength: 8},
	}

	simulator := services.NewSimulator(teams)
	predictor := services.NewPredictor()
	api := handlers.NewAPI(simulator, predictor)

	r := http.NewServeMux()
	api.RegisterRoutes(r)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	log.Printf("Starting server on %s", addr)
	server := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Fatal(server.ListenAndServe())
}

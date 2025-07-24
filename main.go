package main

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"
    "league-simulator/models"
    "league-simulator/services"
    "league-simulator/handlers"
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

    r := mux.NewRouter()
    api.RegisterRoutes(r)

    log.Println("Server running at http://localhost:8080")
    http.ListenAndServe(":8080", r)
}

package handlers

import (
	"encoding/json"
	"math"
	"net/http"

	"league-simulator/models"
	"league-simulator/services"

	"github.com/gorilla/mux"
)

type API struct {
	Simulator services.LeagueSimulator
	Predictor services.Predictor
}

func NewAPI(sim services.LeagueSimulator, pred services.Predictor) *API {
	return &API{
		Simulator: sim,
		Predictor: pred,
	}
}

func (api *API) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", api.LandingPage).Methods("GET")
	router.HandleFunc("/simulate/week", api.SimulateWeek).Methods("POST")
	router.HandleFunc("/simulate/all", api.SimulateAll).Methods("POST")
	router.HandleFunc("/standings", api.GetStandings).Methods("GET")
	router.HandleFunc("/predict", api.PredictRemaining).Methods("GET")
	router.HandleFunc("/matches", api.Matches).Methods("GET")
	router.HandleFunc("/match/edit", api.EditMatchResult).Methods("POST")
	router.HandleFunc("/reset", api.Reset).Methods("POST")
}

func (api *API) LandingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Football League Simulator API"))
}

func (api *API) SimulateWeek(w http.ResponseWriter, r *http.Request) {
	played := api.Simulator.SimulateWeek()
	w.Header().Set("Content-Type", "application/json")

	if !played {
		w.WriteHeader(http.StatusGone) // 410: Artık oynanacak maç yok
		json.NewEncoder(w).Encode(map[string]string{
			"message": "No more matches left to simulate",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "One week simulated",
	})
}

func (api *API) SimulateAll(w http.ResponseWriter, r *http.Request) {
	api.Simulator.SimulateAll()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "All remaining matches simulated",
	})
}

func (api *API) GetStandings(w http.ResponseWriter, r *http.Request) {
	standings := api.Simulator.GetStandings()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standings)
}
func (api *API) PredictRemaining(w http.ResponseWriter, r *http.Request) {
	allMatches := api.Simulator.Matches()
	var flatMatches []models.Match
	for _, weekMatches := range allMatches {
		flatMatches = append(flatMatches, weekMatches...)
	}

	currentStandings := api.Simulator.StandingsCopy()

	// Check if at least 4 weeks have been played
	minGamesPlayed := 4
	hasEnoughGames := false
	for _, standing := range currentStandings {
		if standing.Played >= minGamesPlayed {
			hasEnoughGames = true
			break
		}
	}

	if !hasEnoughGames {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Predictions are only available from week 4 onwards. Please simulate more matches.",
		})
		return
	}

	// Calculate championship probabilities
	championshipProbabilities := calculateChampionshipProbabilities(flatMatches, currentStandings)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"championship_probabilities": championshipProbabilities,
		"message":                    "Championship winning probabilities based on current form",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// calculateChampionshipProbabilities calculates each team's probability of winning the league
func calculateChampionshipProbabilities(matches []models.Match, standings map[int]*models.Standing) []map[string]any {
	// Get predicted final standings
	predictedStandings := make(map[int]*models.Standing)
	for id, standing := range standings {
		predictedStandings[id] = &models.Standing{
			Team:         standing.Team,
			Played:       standing.Played,
			Won:          standing.Won,
			Drawn:        standing.Drawn,
			Lost:         standing.Lost,
			GoalsFor:     standing.GoalsFor,
			GoalsAgainst: standing.GoalsAgainst,
			GoalDiff:     standing.GoalDiff,
			Points:       standing.Points,
		}
	}

	// Simulate remaining matches
	for _, match := range matches {
		if !match.Played {
			homeStanding := predictedStandings[match.Home.ID]
			awayStanding := predictedStandings[match.Away.ID]

			homePointsPerGame := 0.0
			awayPointsPerGame := 0.0

			if homeStanding.Played > 0 {
				homePointsPerGame = float64(homeStanding.Points) / float64(homeStanding.Played)
			}
			if awayStanding.Played > 0 {
				awayPointsPerGame = float64(awayStanding.Points) / float64(awayStanding.Played)
			}

			formDifference := homePointsPerGame - awayPointsPerGame

			if formDifference > 0.5 {
				homeStanding.Won++
				homeStanding.Points += 3
				awayStanding.Lost++
			} else if formDifference < -0.5 {
				awayStanding.Won++
				awayStanding.Points += 3
				homeStanding.Lost++
			} else {
				homeStanding.Drawn++
				homeStanding.Points += 1
				awayStanding.Drawn++
				awayStanding.Points += 1
			}

			homeStanding.Played++
			awayStanding.Played++
		}
	}

	// Convert to slice and find max points
	var allStandings []models.Standing
	maxPoints := 0
	for _, standing := range predictedStandings {
		allStandings = append(allStandings, *standing)
		if standing.Points > maxPoints {
			maxPoints = standing.Points
		}
	}

	// Calculate probabilities based on points difference from leader
	var probabilities []map[string]any
	totalProbability := 0.0

	for _, standing := range allStandings {
		pointsDiff := maxPoints - standing.Points

		// Apply stricter rules based on week
		week := 0
		for _, s := range standings {
			if s.Played > week {
				week = s.Played
			}
		}

		var probability float64
		switch {
		case week >= 6:
			// In 6th week, only the leader can win
			if pointsDiff == 0 {
				probability = 100.0
			} else {
				probability = 0.0
			}
		case week >= 5:
			// After 5th week, >3 points behind can't win
			if pointsDiff == 0 {
				probability = 50.0
			} else if pointsDiff <= 3 {
				probability = 30.0
			} else {
				probability = 0.0
			}
		case week >= 4:
			// After 4th week, >6 points behind can't win
			if pointsDiff == 0 {
				probability = 50.0
			} else if pointsDiff <= 3 {
				probability = 30.0
			} else if pointsDiff <= 6 {
				probability = 15.0
			} else {
				probability = 0.0
			}
		default:
			// Default probabilities
			if pointsDiff == 0 {
				probability = 50.0
			} else if pointsDiff <= 3 {
				probability = 30.0
			} else if pointsDiff <= 6 {
				probability = 15.0
			} else if pointsDiff <= 9 {
				probability = 4.0
			} else {
				probability = 1.0
			}
		}

		probabilities = append(probabilities, map[string]any{
			"team_name":   standing.Team.Name,
			"probability": probability,
		})
		totalProbability += probability
	}

	// Normalize probabilities to sum to 100%
	for i := range probabilities {
		originalProb := probabilities[i]["probability"].(float64)
		normalizedProb := (originalProb / totalProbability) * 100
		probabilities[i]["probability"] = round(normalizedProb, 1)
	}

	return probabilities
}

// calculateMatchPredictions calculates win percentages for unplayed matches based on current standings
func calculateMatchPredictions(matches []models.Match, standings map[int]*models.Standing) []map[string]any {
	var predictions []map[string]any

	for _, match := range matches {
		if !match.Played {
			homeStanding := standings[match.Home.ID]
			awayStanding := standings[match.Away.ID]

			// Calculate points per game (avoid division by zero)
			homePointsPerGame := 0.0
			awayPointsPerGame := 0.0

			if homeStanding.Played > 0 {
				homePointsPerGame = float64(homeStanding.Points) / float64(homeStanding.Played)
			}
			if awayStanding.Played > 0 {
				awayPointsPerGame = float64(awayStanding.Points) / float64(awayStanding.Played)
			}

			// Calculate strength ratio (no home advantage)
			homeStrength := homePointsPerGame
			awayStrength := awayPointsPerGame
			totalStrength := homeStrength + awayStrength

			// Calculate percentages
			homeWinPercentage := 0.0
			awayWinPercentage := 0.0
			drawPercentage := 15.0 // Base draw percentage

			if totalStrength > 0 {
				homeWinPercentage = (homeStrength / totalStrength) * 85.0 // 85% for non-draw outcomes
				awayWinPercentage = (awayStrength / totalStrength) * 85.0
			} else {
				// If no games played, assume equal chances
				homeWinPercentage = 42.5
				awayWinPercentage = 42.5
				drawPercentage = 15.0
			}

			// Ensure percentages add up to 100%
			total := homeWinPercentage + awayWinPercentage + drawPercentage
			if total > 0 {
				homeWinPercentage = (homeWinPercentage / total) * 100
				awayWinPercentage = (awayWinPercentage / total) * 100
				drawPercentage = (drawPercentage / total) * 100
			}

			predictions = append(predictions, map[string]any{
				"match_id":            match.ID,
				"home_team":           match.Home.Name,
				"away_team":           match.Away.Name,
				"week":                match.Week,
				"home_win_percentage": round(homeWinPercentage, 1),
				"away_win_percentage": round(awayWinPercentage, 1),
				"draw_percentage":     round(drawPercentage, 1),
				"home_current_points": homeStanding.Points,
				"away_current_points": awayStanding.Points,
				"home_played":         homeStanding.Played,
				"away_played":         awayStanding.Played,
			})
		}
	}

	return predictions
}

// round rounds a float64 to specified decimal places
func round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func (api *API) Matches(w http.ResponseWriter, r *http.Request) {
	allMatches := api.Simulator.Matches()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(allMatches); err != nil {
		http.Error(w, "Failed to encode matches", http.StatusInternalServerError)
	}
}
func (api *API) EditMatchResult(w http.ResponseWriter, r *http.Request) {
	var result struct {
		MatchID   int `json:"match_id"`
		HomeGoals int `json:"home_goals"`
		AwayGoals int `json:"away_goals"`
	}

	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := api.Simulator.EditMatchResult(result.MatchID, result.HomeGoals, result.AwayGoals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.Simulator.RecalculateStandings()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Match result updated"})
}

func (api *API) Reset(w http.ResponseWriter, r *http.Request) {
	api.Simulator.Reset()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "League has been reset",
	})
}

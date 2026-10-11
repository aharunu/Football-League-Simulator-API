package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"sort"

	"league-simulator/models"
	"league-simulator/services"
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

func (api *API) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /{$}", api.LandingPage)
	router.HandleFunc("POST /simulate/week", api.SimulateWeek)
	router.HandleFunc("POST /simulate/all", api.SimulateAll)
	router.HandleFunc("GET /standings", api.GetStandings)
	router.HandleFunc("GET /predict", api.PredictRemaining)
	router.HandleFunc("GET /matches", api.Matches)
	router.HandleFunc("POST /match/edit", api.EditMatchResult)
	router.HandleFunc("POST /reset", api.Reset)
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
	allMatches, currentStandings := api.Simulator.Snapshot()
	var flatMatches []models.Match
	for _, weekMatches := range allMatches {
		flatMatches = append(flatMatches, weekMatches...)
	}

	// Require four complete rounds, not four arbitrary results for one team.
	requiredWeeks := 4
	if len(allMatches) < requiredWeeks {
		requiredWeeks = len(allMatches)
	}
	hasEnoughGames := requiredWeeks > 0
	for _, week := range allMatches[:requiredWeeks] {
		for _, match := range week {
			if !match.Played {
				hasEnoughGames = false
			}
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
	championshipProbabilities := calculateChampionshipProbabilities(flatMatches, currentStandings, api.Predictor)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"championship_probabilities": championshipProbabilities,
		"message":                    "Heuristic championship estimates based on current form; not calibrated probabilities",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// calculateChampionshipProbabilities calculates each team's probability of winning the league
func calculateChampionshipProbabilities(matches []models.Match, standings map[int]*models.Standing, predictor services.Predictor) []map[string]any {
	predicted := predictor.PredictFinalStandings(matches, standings)
	ids := make([]int, 0, len(standings))
	remaining := make(map[int]int)
	for _, match := range matches {
		if !match.Played {
			remaining[match.Home.ID]++
			remaining[match.Away.ID]++
		}
	}
	maxCurrent, maxPredicted := 0, 0
	for id, st := range standings {
		ids = append(ids, id)
		if st.Points > maxCurrent {
			maxCurrent = st.Points
		}
		if predicted[id].Points > maxPredicted {
			maxPredicted = predicted[id].Points
		}
	}
	sort.Ints(ids)
	finished := len(remaining) == 0
	var winner *models.Standing
	if finished {
		for _, id := range ids {
			st := standings[id]
			if winner == nil || st.Points > winner.Points || (st.Points == winner.Points && (st.GoalDiff > winner.GoalDiff || (st.GoalDiff == winner.GoalDiff && st.GoalsFor > winner.GoalsFor))) {
				winner = st
			}
		}
	}
	weights := make([]float64, len(ids))
	total := 0.0
	for i, id := range ids {
		st := standings[id]
		if finished {
			if st.Points == winner.Points && st.GoalDiff == winner.GoalDiff && st.GoalsFor == winner.GoalsFor {
				weights[i] = 1
			}
		} else if st.Points+3*remaining[id] >= maxCurrent {
			diff := maxPredicted - predicted[id].Points
			switch {
			case diff == 0:
				weights[i] = 50
			case diff <= 3:
				weights[i] = 30
			case diff <= 6:
				weights[i] = 15
			case diff <= 9:
				weights[i] = 4
			default:
				weights[i] = 1
			}
		}
		total += weights[i]
	}
	// Allocate tenths by largest remainder so the returned sum is exactly 100%.
	units := make([]int, len(ids))
	fractions := make([]float64, len(ids))
	order := make([]int, len(ids))
	used := 0
	for i := range ids {
		exact := weights[i] / total * 1000
		units[i] = int(math.Floor(exact))
		used += units[i]
		fractions[i] = exact - float64(units[i])
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return fractions[order[i]] > fractions[order[j]] })
	for i := 0; i < 1000-used && len(order) > 0; i++ {
		units[order[i%len(order)]]++
	}
	result := make([]map[string]any, 0, len(ids))
	for i, id := range ids {
		result = append(result, map[string]any{"team_name": standings[id].Team.Name, "probability": float64(units[i]) / 10})
	}
	return result
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
		MatchID   *int `json:"match_id"`
		HomeGoals *int `json:"home_goals"`
		AwayGoals *int `json:"away_goals"`
	}
	fail := func(status int, message string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": message})
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		fail(http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		fail(http.StatusBadRequest, "Expected one JSON object")
		return
	}
	if result.MatchID == nil || result.HomeGoals == nil || result.AwayGoals == nil || *result.MatchID <= 0 {
		fail(http.StatusBadRequest, "match_id, home_goals and away_goals are required; match_id must be positive")
		return
	}
	err := api.Simulator.EditMatchResult(*result.MatchID, *result.HomeGoals, *result.AwayGoals)
	if errors.Is(err, services.ErrInvalidScore) {
		fail(http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, services.ErrMatchNotFound) {
		fail(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		fail(http.StatusInternalServerError, "Unable to update match")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Match result updated"})
}

func (api *API) Reset(w http.ResponseWriter, r *http.Request) {
	api.Simulator.Reset()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "League has been reset",
	})
}

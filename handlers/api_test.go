package handlers

import (
	"encoding/json"
	"league-simulator/models"
	"league-simulator/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testAPI() (*API, *http.ServeMux) {
	sim := services.NewSimulator([]models.Team{{ID: 1, Name: "A", Strength: 5}, {ID: 2, Name: "B", Strength: 5}, {ID: 3, Name: "C", Strength: 5}, {ID: 4, Name: "D", Strength: 5}})
	api := NewAPI(sim, services.NewPredictor())
	r := http.NewServeMux()
	api.RegisterRoutes(r)
	return api, r
}

func TestEditValidation(t *testing.T) {
	api, r := testAPI()
	cases := []struct {
		body   string
		status int
	}{
		{`{"match_id":1,"home_goals":0,"away_goals":0}`, 200},
		{`{"match_id":999,"home_goals":0,"away_goals":0}`, 404},
		{`{"match_id":1,"home_goals":-1,"away_goals":0}`, 400},
		{`{"match_id":1,"home_goals":101,"away_goals":0}`, 400},
		{`{"match_id":1}`, 400}, {`null`, 400}, {`{}`, 400},
		{`{"match_id":1,"home_goals":null,"away_goals":0}`, 400},
		{`{"match_id":0,"home_goals":0,"away_goals":0}`, 400},
		{`{"match_id":1,"home_goals":0,"away_goals":0,"extra":1}`, 400},
		{`{"match_id":1,"home_goals":0,"away_goals":0} {}`, 400},
		{`{"match_id":1,"home_goals":1.5,"away_goals":0}`, 400},
		{strings.Repeat(" ", 4097) + `{}`, 400},
	}
	for _, c := range cases {
		t.Run(c.body[:min(len(c.body), 60)], func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/match/edit", strings.NewReader(c.body)))
			if w.Code != c.status {
				t.Fatalf("status %d, body %s", w.Code, w.Body.String())
			}
			if !json.Valid(w.Body.Bytes()) {
				t.Fatal("non-JSON response")
			}
		})
	}
	m, _ := api.Simulator.GetMatchByID(1)
	if m.HomeGoals != 0 || m.AwayGoals != 0 {
		t.Fatal("invalid request changed match")
	}
}

func TestPredictRequiresCompleteRounds(t *testing.T) {
	api, r := testAPI()
	for _, week := range api.Simulator.Matches()[:4] {
		for _, m := range week {
			if m.Home.ID == 1 || m.Away.ID == 1 {
				api.Simulator.EditMatchResult(m.ID, 0, 0)
			}
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/predict", nil))
	if w.Code != 400 {
		t.Fatalf("partial rounds accepted: %d", w.Code)
	}
	for i := 0; i < 4; i++ {
		api.Simulator.SimulateWeek()
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/predict", nil))
	if w.Code != 200 {
		t.Fatalf("complete rounds rejected %s", w.Body.String())
	}
}

func TestFinalPredictionUsesGoalDifference(t *testing.T) {
	standings := map[int]*models.Standing{
		1: {Team: models.Team{ID: 1, Name: "A"}, Points: 6, GoalDiff: 3, GoalsFor: 5},
		2: {Team: models.Team{ID: 2, Name: "B"}, Points: 6, GoalDiff: 1, GoalsFor: 5},
	}
	p := calculateChampionshipProbabilities(nil, standings, services.NewPredictor())
	if p[0]["probability"] != float64(100) || p[1]["probability"] != float64(0) {
		t.Fatalf("wrong winner %v", p)
	}
}

func TestPredictionRoundingAndSharedTitle(t *testing.T) {
	standings := map[int]*models.Standing{}
	for i := 1; i <= 3; i++ {
		standings[i] = &models.Standing{Team: models.Team{ID: i}, Points: 5}
	}
	p := calculateChampionshipProbabilities(nil, standings, services.NewPredictor())
	units := 0
	for _, row := range p {
		units += int(row["probability"].(float64)*10 + 0.5)
	}
	if units != 1000 {
		t.Fatalf("sum %d", units)
	}
	if p[0]["probability"] != 33.4 {
		t.Fatal("rounding not deterministic")
	}
}

func TestRouteMethodsAndSeasonLifecycle(t *testing.T) {
	api, router := testAPI()
	for _, c := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/", 200}, {"GET", "/unknown", 404},
		{"GET", "/simulate/week", 405}, {"POST", "/simulate/all", 200},
		{"POST", "/simulate/week", 410}, {"GET", "/standings", 200},
		{"GET", "/matches", 200}, {"GET", "/predict", 200}, {"POST", "/reset", 200},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(c.method, c.path, nil))
		if w.Code != c.status {
			t.Fatalf("%s %s: got %d, expected %d", c.method, c.path, w.Code, c.status)
		}
	}
	for _, st := range api.Simulator.GetStandings() {
		if st.Played != 0 {
			t.Fatal("HTTP reset failed")
		}
	}
}

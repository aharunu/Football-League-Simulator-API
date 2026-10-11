package services

import (
	"errors"
	"league-simulator/models"
	"reflect"
	"sync"
	"testing"
)

func testTeams(n int) []models.Team {
	teams := make([]models.Team, n)
	for i := range teams {
		teams[i] = models.Team{ID: i + 1, Strength: 5}
	}
	return teams
}

func TestEditDrawDoesNotDoubleCount(t *testing.T) {
	s := NewSimulator(testTeams(4))
	for _, score := range [][2]int{{0, 0}, {2, 1}, {0, 0}, {0, 0}} {
		if err := s.EditMatchResult(1, score[0], score[1]); err != nil {
			t.Fatal(err)
		}
		m, _ := s.GetMatchByID(1)
		st := s.StandingsCopy()[m.Home.ID]
		if st.Played != 1 || st.Played != st.Won+st.Drawn+st.Lost {
			t.Fatalf("double count: %+v", st)
		}
	}
	m, _ := s.GetMatchByID(1)
	st := s.StandingsCopy()[m.Home.ID]
	if st.Drawn != 1 || st.Points != 1 || st.GoalsFor != 0 {
		t.Fatalf("incorrect draw: %+v", st)
	}
}

func TestSimulateAllSkipsEditedRound(t *testing.T) {
	s := NewSimulator(testTeams(4))
	for _, m := range s.Matches()[0] {
		s.EditMatchResult(m.ID, 0, 0)
	}
	s.SimulateAll()
	for _, week := range s.Matches() {
		for _, m := range week {
			if !m.Played {
				t.Fatalf("unplayed %d", m.ID)
			}
		}
	}
	for _, st := range s.GetStandings() {
		if st.Played != 6 {
			t.Fatalf("played %d", st.Played)
		}
	}
	if s.SimulateWeek() {
		t.Fatal("finished season advanced")
	}
}

func TestInvalidScoreLeavesStateUnchanged(t *testing.T) {
	s := NewSimulator(testTeams(4))
	before := s.Matches()
	for _, score := range [][2]int{{-1, 0}, {0, -1}, {101, 0}, {0, 101}} {
		if err := s.EditMatchResult(1, score[0], score[1]); !errors.Is(err, ErrInvalidScore) {
			t.Fatalf("error %v", err)
		}
	}
	if !reflect.DeepEqual(before, s.Matches()) {
		t.Fatal("invalid score mutated fixtures")
	}
	if err := s.EditMatchResult(999, 0, 0); !errors.Is(err, ErrMatchNotFound) {
		t.Fatal(err)
	}
}

func TestSnapshotsAreDetached(t *testing.T) {
	s := NewSimulator(testTeams(4))
	matches, standings := s.Snapshot()
	matches[0][0].HomeGoals = 99
	standings[1].Points = 999
	m, _ := s.GetMatchByID(1)
	m.HomeGoals = 88
	current, _ := s.GetMatchByID(1)
	if current.HomeGoals != 0 || s.StandingsCopy()[1].Points != 0 {
		t.Fatal("getter leaked internal state")
	}
}

func TestFixtureCoverage(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 4, 5, 6} {
		fixtures := NewSimulator(testTeams(n)).Matches()
		pairs := map[[2]int]int{}
		for _, week := range fixtures {
			seen := map[int]bool{}
			for _, m := range week {
				if seen[m.Home.ID] || seen[m.Away.ID] || m.Home.ID == m.Away.ID {
					t.Fatalf("invalid round for %d teams", n)
				}
				seen[m.Home.ID] = true
				seen[m.Away.ID] = true
				pairs[[2]int{m.Home.ID, m.Away.ID}]++
			}
		}
		if len(pairs) != n*(n-1) {
			t.Fatalf("teams %d: %d ordered pairs", n, len(pairs))
		}
		for _, count := range pairs {
			if count != 1 {
				t.Fatal("repeated pairing")
			}
		}
	}
}

func TestResetAndConcurrentAccess(t *testing.T) {
	s := NewSimulator(testTeams(4))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.SimulateWeek()
				s.GetStandings()
				s.Snapshot()
				s.EditMatchResult(1, j%5, 0)
				s.GetMatchByID(1)
				s.RecalculateStandings()
				s.Reset()
			}
		}()
	}
	wg.Wait()
	s.Reset()
	for _, st := range s.GetStandings() {
		if st.Played != 0 || st.Points != 0 {
			t.Fatal("reset retained standings")
		}
	}
	if !s.SimulateWeek() {
		t.Fatal("reset did not rewind season")
	}
}

func TestGoalsNeverNegative(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if goals := randomGoals(-100); goals < 0 || goals > 5 {
			t.Fatalf("goals %d", goals)
		}
	}
}

func TestPredictionIndependentOfFixtureOrder(t *testing.T) {
	s := NewSimulator(testTeams(4))
	s.EditMatchResult(1, 2, 0)
	s.EditMatchResult(2, 0, 0)
	standings := s.StandingsCopy()
	before := s.StandingsCopy()
	var matches []models.Match
	for _, week := range s.Matches() {
		matches = append(matches, week...)
	}
	p := NewPredictor()
	a := p.PredictFinalStandings(matches, standings)
	for i, j := 0, len(matches)-1; i < j; i, j = i+1, j-1 {
		matches[i], matches[j] = matches[j], matches[i]
	}
	b := p.PredictFinalStandings(matches, standings)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("fixture order changed prediction")
	}
	if !reflect.DeepEqual(before, standings) {
		t.Fatal("prediction changed input")
	}
	for _, st := range a {
		if st.Played != 6 || st.GoalDiff != st.GoalsFor-st.GoalsAgainst || st.Points != 3*st.Won+st.Drawn {
			t.Fatalf("invalid prediction %+v", st)
		}
	}
}

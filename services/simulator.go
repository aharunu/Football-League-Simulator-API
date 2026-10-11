package services

import (
	"errors"
	"fmt"
	"league-simulator/models"
	"math/rand"
	"sort"
	"sync"
)

// LeagueSimulator defines the interface for the league simulation.
var ErrMatchNotFound = errors.New("match not found")
var ErrInvalidScore = errors.New("scores must be between 0 and 100")

type LeagueSimulator interface {
	Snapshot() ([][]models.Match, map[int]*models.Standing)
	SimulateWeek() bool
	SimulateAll()
	GetStandings() []models.Standing
	Matches() [][]models.Match
	StandingsCopy() map[int]*models.Standing
	EditMatchResult(matchID, homeGoals, awayGoals int) error
	RecalculateStandings()
	GetMatchByID(matchID int) (*models.Match, error)
	Reset()
}

// SimulatorImpl implements the LeagueSimulator interface.

type SimulatorImpl struct {
	mu          sync.RWMutex
	teams       []models.Team
	matches     [][]models.Match
	standings   map[int]*models.Standing
	currentWeek int
}

func NewSimulator(teams []models.Team) LeagueSimulator {
	standings := make(map[int]*models.Standing)
	for _, team := range teams {
		standings[team.ID] = &models.Standing{Team: team}
	}

	return &SimulatorImpl{
		teams:       append([]models.Team(nil), teams...),
		standings:   standings,
		matches:     generateFixtures(teams),
		currentWeek: 0,
	}
}
func generateFixtures(teams []models.Team) [][]models.Match {
	numTeams := len(teams)
	if numTeams < 2 {
		return [][]models.Match{}
	}
	// A nil entry is a bye for leagues with an odd number of teams.
	slots := make([]*models.Team, numTeams)
	for i := range teams {
		slots[i] = &teams[i]
	}
	if numTeams%2 != 0 {
		slots = append(slots, nil)
		numTeams++
	}
	weeks := (numTeams - 1) * 2
	fixtures := make([][]models.Match, weeks)
	matchID := 1

	// Takımların sırasını değiştirmemek için kopyasını al
	teamList := append([]*models.Team(nil), slots...)

	// İlk devre
	for w := 0; w < numTeams-1; w++ {
		var weekMatches []models.Match
		for i := 0; i < numTeams/2; i++ {
			home := teamList[i]
			away := teamList[numTeams-1-i]
			if home == nil || away == nil {
				continue
			}
			weekMatches = append(weekMatches, models.Match{
				ID:     matchID,
				Home:   *home,
				Away:   *away,
				Week:   w + 1,
				Played: false,
			})
			matchID++
		}
		fixtures[w] = weekMatches
		teamList = append([]*models.Team{teamList[0]}, append(teamList[numTeams-1:], teamList[1:numTeams-1]...)...)
	}

	// İkinci devre (ev-deplasman değişimi)
	teamList2 := append([]*models.Team(nil), slots...)
	for w := 0; w < numTeams-1; w++ {
		var weekMatches []models.Match
		for i := 0; i < numTeams/2; i++ {
			home := teamList2[numTeams-1-i]
			away := teamList2[i]
			if home == nil || away == nil {
				continue
			}
			weekMatches = append(weekMatches, models.Match{
				ID:     matchID,
				Home:   *home,
				Away:   *away,
				Week:   w + numTeams,
				Played: false,
			})
			matchID++
		}
		fixtures[w+numTeams-1] = weekMatches
		teamList2 = append([]*models.Team{teamList2[0]}, append(teamList2[numTeams-1:], teamList2[1:numTeams-1]...)...)
	}

	return fixtures
}

func (s *SimulatorImpl) SimulateWeek() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.simulateWeek()
}

func (s *SimulatorImpl) simulateWeek() bool {
	for s.currentWeek < len(s.matches) {
		played := false
		for i := range s.matches[s.currentWeek] {
			match := &s.matches[s.currentWeek][i]
			if match.Played {
				continue
			}
			match.HomeGoals, match.AwayGoals = simulateMatch(match.Home.Strength, match.Away.Strength)
			match.Played = true
			updateStandings(s.standings, *match)
			played = true
		}
		s.currentWeek++
		if played {
			return true
		}
	}
	return false
}

func (s *SimulatorImpl) SimulateAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for s.simulateWeek() {
	}
}

func simulateMatch(homeStrength, awayStrength int) (int, int) {
	homeGoals := randomGoals(homeStrength)
	awayGoals := randomGoals(awayStrength)

	return homeGoals, awayGoals
}

func randomGoals(strength int) int {
	base := float64(strength) / 5
	noise := rand.NormFloat64() * 0.5 // normal noise
	goals := int(base + noise + 1.5)
	if goals < 0 {
		return 0
	}
	if goals > 5 {
		return 5
	}
	return goals
}

func updateStandings(standings map[int]*models.Standing, match models.Match) {
	homeStanding := standings[match.Home.ID]
	awayStanding := standings[match.Away.ID]

	homeStanding.Played++
	awayStanding.Played++

	homeStanding.GoalsFor += match.HomeGoals
	homeStanding.GoalsAgainst += match.AwayGoals
	awayStanding.GoalsFor += match.AwayGoals
	awayStanding.GoalsAgainst += match.HomeGoals

	if match.HomeGoals > match.AwayGoals {
		homeStanding.Won++
		awayStanding.Lost++
	} else if match.HomeGoals < match.AwayGoals {
		homeStanding.Lost++
		awayStanding.Won++
	} else {
		homeStanding.Drawn++
		awayStanding.Drawn++
	}

	homeStanding.GoalDiff = homeStanding.GoalsFor - homeStanding.GoalsAgainst
	awayStanding.GoalDiff = awayStanding.GoalsFor - awayStanding.GoalsAgainst

	homeStanding.Points = homeStanding.Won*3 + homeStanding.Drawn
	awayStanding.Points = awayStanding.Won*3 + awayStanding.Drawn
}

// EditMatchResult allows editing the result of a specific match by ID
func (s *SimulatorImpl) EditMatchResult(matchID, homeGoals, awayGoals int) error {
	if homeGoals < 0 || awayGoals < 0 || homeGoals > 100 || awayGoals > 100 {
		return ErrInvalidScore
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for wi := range s.matches {
		for mi := range s.matches[wi] {
			match := &s.matches[wi][mi]
			if match.ID != matchID {
				continue
			}
			if match.Played {
				reverseStandings(s.standings, *match)
			}
			match.HomeGoals, match.AwayGoals, match.Played = homeGoals, awayGoals, true
			updateStandings(s.standings, *match)
			return nil
		}
	}
	return fmt.Errorf("%w: %d", ErrMatchNotFound, matchID)
}

// RecalculateStandings recalculates all standings from scratch based on played matches
func (s *SimulatorImpl) RecalculateStandings() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Reset all standings
	for _, standing := range s.standings {
		standing.Played = 0
		standing.Won = 0
		standing.Drawn = 0
		standing.Lost = 0
		standing.GoalsFor = 0
		standing.GoalsAgainst = 0
		standing.GoalDiff = 0
		standing.Points = 0
	}

	// Recalculate from all played matches
	for _, weekMatches := range s.matches {
		for _, match := range weekMatches {
			if match.Played {
				updateStandings(s.standings, match)
			}
		}
	}
}

// reverseStandings reverses the effect of a match on standings
func reverseStandings(standings map[int]*models.Standing, match models.Match) {
	homeStanding := standings[match.Home.ID]
	awayStanding := standings[match.Away.ID]

	homeStanding.Played--
	awayStanding.Played--

	homeStanding.GoalsFor -= match.HomeGoals
	homeStanding.GoalsAgainst -= match.AwayGoals
	awayStanding.GoalsFor -= match.AwayGoals
	awayStanding.GoalsAgainst -= match.HomeGoals

	if match.HomeGoals > match.AwayGoals {
		homeStanding.Won--
		awayStanding.Lost--
	} else if match.HomeGoals < match.AwayGoals {
		homeStanding.Lost--
		awayStanding.Won--
	} else {
		homeStanding.Drawn--
		awayStanding.Drawn--
	}

	homeStanding.GoalDiff = homeStanding.GoalsFor - homeStanding.GoalsAgainst
	awayStanding.GoalDiff = awayStanding.GoalsFor - awayStanding.GoalsAgainst

	homeStanding.Points = homeStanding.Won*3 + homeStanding.Drawn
	awayStanding.Points = awayStanding.Won*3 + awayStanding.Drawn
}

// GetMatchByID finds and returns a match by its ID
func (s *SimulatorImpl) GetMatchByID(matchID int) (*models.Match, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for weekIdx := range s.matches {
		for matchIdx := range s.matches[weekIdx] {
			if s.matches[weekIdx][matchIdx].ID == matchID {
				match := s.matches[weekIdx][matchIdx]
				return &match, nil
			}
		}
	}
	return nil, fmt.Errorf("%w: %d", ErrMatchNotFound, matchID)
}

// GetStandings returns the current standings of the league.
func (s *SimulatorImpl) GetStandings() []models.Standing {
	s.mu.RLock()
	defer s.mu.RUnlock()
	standings := make([]models.Standing, 0, len(s.standings))
	for _, standing := range s.standings {
		standings = append(standings, *standing)
	}

	// Sort standings by points, then goal difference, then goals for
	sort.Slice(standings, func(i, j int) bool {
		if standings[i].Points != standings[j].Points {
			return standings[i].Points > standings[j].Points
		}
		if standings[i].GoalDiff != standings[j].GoalDiff {
			return standings[i].GoalDiff > standings[j].GoalDiff
		}
		if standings[i].GoalsFor != standings[j].GoalsFor {
			return standings[i].GoalsFor > standings[j].GoalsFor
		}
		return standings[i].Team.ID < standings[j].Team.ID
	})

	return standings
}

// Snapshot returns detached copies of fixtures and standings from one instant.
func (s *SimulatorImpl) Snapshot() ([][]models.Match, map[int]*models.Standing) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	matches := make([][]models.Match, len(s.matches))
	for i, week := range s.matches {
		matches[i] = append([]models.Match(nil), week...)
	}
	standings := make(map[int]*models.Standing, len(s.standings))
	for id, standing := range s.standings {
		copy := *standing
		standings[id] = &copy
	}
	return matches, standings
}

func (s *SimulatorImpl) Matches() [][]models.Match { matches, _ := s.Snapshot(); return matches }
func (s *SimulatorImpl) StandingsCopy() map[int]*models.Standing {
	_, standings := s.Snapshot()
	return standings
}

// Reset resets the league simulator to the initial state
func (s *SimulatorImpl) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Reset all matches
	for weekIdx := range s.matches {
		for matchIdx := range s.matches[weekIdx] {
			match := &s.matches[weekIdx][matchIdx]
			match.HomeGoals = 0
			match.AwayGoals = 0
			match.Played = false
		}
	}
	// Reset standings
	for _, standing := range s.standings {
		standing.Played = 0
		standing.Won = 0
		standing.Drawn = 0
		standing.Lost = 0
		standing.GoalsFor = 0
		standing.GoalsAgainst = 0
		standing.GoalDiff = 0
		standing.Points = 0
	}
	s.currentWeek = 0
}

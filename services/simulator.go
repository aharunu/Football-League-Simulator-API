package services

import (
	"fmt"
	"league-simulator/models"
	"math/rand"
	"sort"
)

// LeagueSimulator defines the interface for the league simulation.
type LeagueSimulator interface {
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
		teams:       teams,
		standings:   standings,
		matches:     generateFixtures(teams),
		currentWeek: 0,
	}
}
func generateFixtures(teams []models.Team) [][]models.Match {
	numTeams := len(teams)
	weeks := (numTeams - 1) * 2
	fixtures := make([][]models.Match, weeks)
	matchID := 1

	// Takımların sırasını değiştirmemek için kopyasını al
	teamList := make([]models.Team, numTeams)
	copy(teamList, teams)

	// İlk devre
	for w := 0; w < numTeams-1; w++ {
		var weekMatches []models.Match
		for i := 0; i < numTeams/2; i++ {
			home := teamList[i]
			away := teamList[numTeams-1-i]
			weekMatches = append(weekMatches, models.Match{
				ID:     matchID,
				Home:   home,
				Away:   away,
				Week:   w + 1,
				Played: false,
			})
			matchID++
		}
		fixtures[w] = weekMatches
		teamList = append([]models.Team{teamList[0]}, append(teamList[numTeams-1:], teamList[1:numTeams-1]...)...)
	}

	// İkinci devre (ev-deplasman değişimi)
	teamList2 := make([]models.Team, numTeams)
	copy(teamList2, teams)
	for w := 0; w < numTeams-1; w++ {
		var weekMatches []models.Match
		for i := 0; i < numTeams/2; i++ {
			home := teamList2[numTeams-1-i]
			away := teamList2[i]
			weekMatches = append(weekMatches, models.Match{
				ID:     matchID,
				Home:   home,
				Away:   away,
				Week:   w + numTeams,
				Played: false,
			})
			matchID++
		}
		fixtures[w+numTeams-1] = weekMatches
		teamList2 = append([]models.Team{teamList2[0]}, append(teamList2[numTeams-1:], teamList2[1:numTeams-1]...)...)
	}

	return fixtures
}

func (s *SimulatorImpl) SimulateWeek() bool {
	if s.currentWeek >= len(s.matches) {
		return false // Tüm maçlar oynandı
	}

	weekMatches := s.matches[s.currentWeek]
	allPlayed := true

	for i := range weekMatches {
		match := &weekMatches[i]
		if !match.Played {
			homeGoals, awayGoals := simulateMatch(match.Home.Strength, match.Away.Strength)

			match.HomeGoals = homeGoals
			match.AwayGoals = awayGoals
			match.Played = true

			updateStandings(s.standings, *match)
			allPlayed = false
		}
	}

	s.matches[s.currentWeek] = weekMatches
	s.currentWeek++
	return !allPlayed
}

// SimulateAll simulates all remaining weeks until no more matches can be played.

func (s *SimulatorImpl) SimulateAll() {
	for s.SimulateWeek() {
		// Simulate all weeks until no more matches can be played
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
	// Find the match in all weeks
	for weekIdx := range s.matches {
		for matchIdx := range s.matches[weekIdx] {
			match := &s.matches[weekIdx][matchIdx]
			if match.ID == matchID {
				// Store old result for standings update
				oldHomeGoals := match.HomeGoals
				oldAwayGoals := match.AwayGoals

				// Update match result
				match.HomeGoals = homeGoals
				match.AwayGoals = awayGoals
				match.Played = true

				// If match was already played, reverse old standings first
				if oldHomeGoals != 0 || oldAwayGoals != 0 {
					reverseStandings(s.standings, models.Match{
						ID:        match.ID,
						Home:      match.Home,
						Away:      match.Away,
						HomeGoals: oldHomeGoals,
						AwayGoals: oldAwayGoals,
						Week:      match.Week,
						Played:    true,
					})
				}

				// Apply new standings
				updateStandings(s.standings, *match)
				return nil
			}
		}
	}
	return fmt.Errorf("match with ID %d not found", matchID)
}

// RecalculateStandings recalculates all standings from scratch based on played matches
func (s *SimulatorImpl) RecalculateStandings() {
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
	for weekIdx := range s.matches {
		for matchIdx := range s.matches[weekIdx] {
			if s.matches[weekIdx][matchIdx].ID == matchID {
				return &s.matches[weekIdx][matchIdx], nil
			}
		}
	}
	return nil, fmt.Errorf("match with ID %d not found", matchID)
}

// GetStandings returns the current standings of the league.
func (s *SimulatorImpl) GetStandings() []models.Standing {
	var standings []models.Standing
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
		return standings[i].GoalsFor > standings[j].GoalsFor
	})

	return standings
}
func (s *SimulatorImpl) Matches() [][]models.Match {
	return s.matches
}
func (s *SimulatorImpl) StandingsCopy() map[int]*models.Standing {
	copied := make(map[int]*models.Standing)
	for id, st := range s.standings {
		copy := *st
		copied[id] = &copy
	}
	return copied
}

// Reset resets the league simulator to the initial state
func (s *SimulatorImpl) Reset() {
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

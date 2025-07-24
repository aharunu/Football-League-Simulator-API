package services

import "league-simulator/models"

type Predictor interface {
	PredictFinalStandings(currentMatches []models.Match, standings map[int]*models.Standing) map[int]*models.Standing
}

type predictorImpl struct{}

func NewPredictor() Predictor {
	return &predictorImpl{}
}

func (p *predictorImpl) PredictFinalStandings(currentMatches []models.Match, standings map[int]*models.Standing) map[int]*models.Standing {
	// Create a copy of current standings for prediction
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

	// Simulate remaining matches based on current form
	for _, match := range currentMatches {
		if !match.Played {
			homeStanding := predictedStandings[match.Home.ID]
			awayStanding := predictedStandings[match.Away.ID]

			// Calculate expected points based on current form
			homePointsPerGame := 0.0
			awayPointsPerGame := 0.0

			if homeStanding.Played > 0 {
				homePointsPerGame = float64(homeStanding.Points) / float64(homeStanding.Played)
			}
			if awayStanding.Played > 0 {
				awayPointsPerGame = float64(awayStanding.Points) / float64(awayStanding.Played)
			}

			// Determine likely result based on form difference (no home advantage)
			formDifference := homePointsPerGame - awayPointsPerGame

			if formDifference > 0.5 {
				// Home team likely to win
				homeStanding.Won++
				homeStanding.Points += 3
				awayStanding.Lost++
			} else if formDifference < -0.5 {
				// Away team likely to win
				awayStanding.Won++
				awayStanding.Points += 3
				homeStanding.Lost++
			} else {
				// Draw likely
				homeStanding.Drawn++
				homeStanding.Points += 1
				awayStanding.Drawn++
				awayStanding.Points += 1
			}

			homeStanding.Played++
			awayStanding.Played++
		}
	}

	return predictedStandings
}

package services

import "league-simulator/models"

type Predictor interface {
	PredictFinalStandings([]models.Match, map[int]*models.Standing) map[int]*models.Standing
}
type predictorImpl struct{}

func NewPredictor() Predictor { return &predictorImpl{} }

// Freeze current form so fixture order does not influence the prediction.
// A predicted win is represented as 1-0 and a draw as 0-0.
func (p *predictorImpl) PredictFinalStandings(matches []models.Match, standings map[int]*models.Standing) map[int]*models.Standing {
	predicted := make(map[int]*models.Standing, len(standings))
	form := make(map[int]float64, len(standings))
	for id, standing := range standings {
		copy := *standing
		predicted[id] = &copy
		if standing.Played > 0 {
			form[id] = float64(standing.Points) / float64(standing.Played)
		}
	}
	for _, match := range matches {
		if match.Played {
			continue
		}
		match.HomeGoals, match.AwayGoals = 0, 0
		difference := form[match.Home.ID] - form[match.Away.ID]
		if difference > 0.5 {
			match.HomeGoals = 1
		} else if difference < -0.5 {
			match.AwayGoals = 1
		}
		updateStandings(predicted, match)
	}
	return predicted
}

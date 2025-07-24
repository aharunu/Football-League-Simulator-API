package models

type Standing struct {
    Team       Team
    Played     int
    Won        int
    Drawn      int
    Lost       int
    GoalsFor   int
    GoalsAgainst int
    GoalDiff   int
    Points     int
}
// Standing represents the standing of a team in the league.
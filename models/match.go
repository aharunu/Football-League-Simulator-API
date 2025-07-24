package models

type Match struct {
    ID       int
    Home     Team
    Away     Team
    HomeGoals int
    AwayGoals int
    Played    bool
    Week      int
}
// Match represents a football match between two teams.
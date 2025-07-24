-- Takımlar tablosu
CREATE TABLE teams (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    strength INTEGER NOT NULL
);

-- Maçlar tablosu
CREATE TABLE matches (
    id INTEGER PRIMARY KEY,
    home_team_id INTEGER,
    away_team_id INTEGER,
    home_goals INTEGER,
    away_goals INTEGER,
    played BOOLEAN,
    week INTEGER,
    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id)
);

-- Puan durumu tablosu
CREATE TABLE standings (
    team_id INTEGER PRIMARY KEY,
    played INTEGER,
    won INTEGER,
    drawn INTEGER,
    lost INTEGER,
    goals_for INTEGER,
    goals_against INTEGER,
    goal_diff INTEGER,
    points INTEGER,
    FOREIGN KEY (team_id) REFERENCES teams(id)
);

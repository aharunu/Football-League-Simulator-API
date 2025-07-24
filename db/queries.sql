-- Get all teams with their strength
SELECT id, name, strength FROM teams ORDER BY name;

-- Get specific team by ID
SELECT * FROM teams WHERE id = 1;

-- Insert new team
INSERT INTO teams (name, strength) VALUES ('Arsenal', 8);

-- Update team strength
UPDATE teams SET strength = 9 WHERE id = 1;

-- Get all matches
SELECT * FROM matches;

-- Get matches for week 1
SELECT * FROM matches WHERE week = 1;

-- Get played matches only
SELECT * FROM matches WHERE played = 1;

-- Update match result
UPDATE matches SET home_goals = 2, away_goals = 1, played = 1 WHERE id = 1;

-- Get all standings
SELECT * FROM standings ORDER BY points DESC;

-- Get top 3 teams
SELECT * FROM standings ORDER BY points DESC LIMIT 3;

-- Get one team's standing
SELECT * FROM standings WHERE team_id = 1;

-- Update team points
UPDATE standings SET points = points + 3 WHERE team_id = 1;

-- Matches with team names
SELECT 
    m.id,
    t1.name AS home_team,
    t2.name AS away_team,
    m.home_goals,
    m.away_goals
FROM matches m
JOIN teams t1 ON m.home_team_id = t1.id
JOIN teams t2 ON m.away_team_id = t2.id;

-- Standings with team names
SELECT 
    t.name,
    s.points,
    s.played
FROM standings s
JOIN teams t ON s.team_id = t.id
ORDER BY s.points DESC;

-- How many teams?
SELECT COUNT(*) FROM teams;

-- How many matches played?
SELECT COUNT(*) FROM matches WHERE played = 1;

-- How many wins for team 1?
SELECT won FROM standings WHERE team_id = 1;
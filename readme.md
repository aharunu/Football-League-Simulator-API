# Football League Simulator API

A comprehensive football league simulation API built with Go that allows you to simulate matches, track standings, predict championship probabilities, and manage league data

## Tech Stack

- **Backend**: Go 1.24+
- **HTTP Router**: gorilla/mux
- **Containerization**: Docker
- **Cloud Deployment**: Google Cloud Run
- **API Testing**: Postman Collection included

## Project Structure

```
Football-League-Simulator-API/
├── handlers/
│   └── api.go              # HTTP handlers and routes
├── models/
│   └── models.go           # Data structures (Team, Match, Standing)
├── services/
│   ├── simulator.go        # Core simulation logic
│   └── predictor.go        # Championship prediction algorithms
├── db/
│   └── schema.sql          # Database schema
│   └── queries.sql         # Database queries
├── collection.json         # Postman collection for API testing
├── Dockerfile             # Docker configuration
├── go.mod                 # Go dependencies
└── main.go               # Application entry point
```

## 🚀 API Endpoints & Implementation

### Core Simulation Engine
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/simulate/week` | POST | Simulate one week of matches
| `/simulate/all` | POST | Simulate entire remaining season
| `/reset` | POST | Reset league to initial state

### Data Retrieval  
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/standings` | GET | Current league table |
| `/matches` | GET | All fixtures (played/unplayed) |
| `/predict` | GET | Championship probabilities |

### Match Management
| Endpoint | Method | Description | 
|----------|--------|-------------|
| `/match/edit` | POST | Edit specific match result |

**Request Format:**
```json
{
  "match_id": 1,
  "home_goals": 2,
  "away_goals": 1
}
```

## 🧪 API Testing & Validation

### Production Postman Collection

A dedicated Postman collection is included for the live production instance. All endpoints are pre-configured with the production base URL:

- **Base URL:** `https://league-simulator-282922766146.europe-west1.run.app`
- **Usage:** Import `collection.json` into Postman to test the live API endpoints directly. All requests are already configured to use the production base URL, so no additional environment setup is required.

**Quick Test Endpoints:**
```bash
# Get current standings
curl https://league-simulator-282922766146.europe-west1.run.app/standings

# Simulate one week
curl -X POST https://league-simulator-282922766146.europe-west1.run.app/simulate/week

# Get championship predictions (after week 4)
curl https://league-simulator-282922766146.europe-west1.run.app/predict
```

### Local Development Setup

1. **Clone and Initialize**
   ```bash
   git clone https://github.com/aharunu/Football-League-Simulator-API.git
   cd Football-League-Simulator-API
   go mod download
   ```

2. **Run Application**
   ```bash
   go run .
   # Server starts at http://localhost:8080
   ```

3. **Verify Installation**
   ```bash
   curl http://localhost:8080/standings
   # Should return standings
   ```

4. **Test All Functionalities**
   ```bash
   # 1. View landing page
   curl http://localhost:8080/
   
   # 2. Check initial empty standings
   curl http://localhost:8080/standings
   
   # 3. View all unplayed matches
   curl http://localhost:8080/matches
   
   # 4. Simulate first week
   curl -X POST http://localhost:8080/simulate/week
   
   # 5. Check updated standings after week 1
   curl http://localhost:8080/standings
   
   # 6. Simulate 3 more weeks (total 4 weeks for predictions)
   curl -X POST http://localhost:8080/simulate/week
   curl -X POST http://localhost:8080/simulate/week
   curl -X POST http://localhost:8080/simulate/week
   
   # 7. Get championship predictions (available after week 4)
   curl http://localhost:8080/predict
   
   # 8. Edit a match result
   curl -X POST http://localhost:8080/match/edit \
     -H "Content-Type: application/json" \
     -d '{"match_id":1,"home_goals":5,"away_goals":0}'
   
   # 9. Check standings after manual edit
   curl http://localhost:8080/standings
   
   # 10. Simulate all remaining matches
   curl -X POST http://localhost:8080/simulate/all
   
   # 11. View final standings
   curl http://localhost:8080/standings
   
   # 12. Reset league to start over
   curl -X POST http://localhost:8080/reset
   ```
## 🐳 Run with Docker

You can build and run the API using Docker without installing Go locally:

### 1. Build Docker Image

```bash
docker build -t football-league-simulator .
```
### 2. Run The Container
```bash
docker run -p 8080:8080 football-league-simulator
```
The API will now be accessible at:
http://localhost:8080

### 3. Test Endpoints
```bash
# Get current standings
curl http://localhost:8080/standings

# Simulate one week of matches
curl -X POST http://localhost:8080/simulate/week

# View all matches
curl http://localhost:8080/matches
```




## 📊 API Response Examples

### Standings Response (Sorted by Performance)
```json
[
  {
    "team": {
      "id": 4,
      "name": "Liverpool", 
      "strength": 8
    },
    "played": 3,
    "won": 3,
    "drawn": 0,
    "lost": 0,
    "goals_for": 9,
    "goals_against": 2,
    "goal_diff": 7,
    "points": 9
  },
  {
    "team": {
      "id": 2,
      "name": "Manchester City",
      "strength": 7  
    },
    "played": 3,
    "won": 2,
    "drawn": 0,
    "lost": 1,
    "goals_for": 6,
    "goals_against": 4,
    "goal_diff": 2,
    "points": 6
  }
]
```

### Championship Predictions (Week 4+ Only)
```json
{
  "championship_probabilities": [
    {
      "team_name": "Liverpool",
      "probability": 68.5
    },
    {
      "team_name": "Manchester City", 
      "probability": 22.3
    },
    {
      "team_name": "Chelsea",
      "probability": 9.2
    },
    {
      "team_name": "Manchester United",
      "probability": 0.0
    }
  ],
  "message": "Championship winning probabilities based on current form"
}
```

### Match Fixtures Response
```json
[
  [
    {
      "id": 1,
      "home": {"id": 1, "name": "Manchester United", "strength": 5},
      "away": {"id": 2, "name": "Manchester City", "strength": 7},
      "home_goals": 1,
      "away_goals": 3,
      "played": true,
      "week": 1
    }
  ],
  [
    {
      "id": 7,
      "home": {"id": 2, "name": "Manchester City", "strength": 7},
      "away": {"id": 1, "name": "Manchester United", "strength": 5},
      "home_goals": 0,
      "away_goals": 0,
      "played": false,
      "week": 4
    }
  ]
]
```
## 💾 Database Schema Design

```sql
-- Team entity with strength rating
CREATE TABLE teams (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    strength INTEGER NOT NULL  -- Used in simulation algorithm
);

-- Match entity with complete game state
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

-- League standings with comprehensive statistics
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
```


## 📄 License

This project is open source and available under the [MIT License](LICENSE).

---

**🏆 Technical Case Study - Football League Simulator API**  
*Demonstrating Go backend development excellence with clean architecture principles*

For technical questions or code review feedback, please open an issue in this repository.
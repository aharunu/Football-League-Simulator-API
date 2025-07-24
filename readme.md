# Football League Simulator API

A comprehensive football league simulation API built with Go that allows you to simulate matches, track standings, predict championship probabilities, and manage league data

## Tech Stack

- **Backend**: Go 1.24+
- **HTTP Router**: Gorilla Mux
- **Containerization**: Docker
- **Cloud Deployment**: Google Cloud Run
- **API Testing**: Postman Collection included

## Project Structure

```
League Simulator/
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

## 🚀 Getting Started

### Prerequisites
- **Go 1.24+** (Uses latest language features)
- **Docker** (Optional, for containerization)
- **Google Cloud CLI** (For production deployment)

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
   # Should return empty standings array: []
   ```

4. **Test Core Functionality**
   ```bash
   # Simulate first week
   curl -X POST http://localhost:8080/simulate/week
   
   # Check updated standings
   curl http://localhost:8080/standings
   ```

## 🧪 API Testing & Validation

### Production Postman Collection

A dedicated Postman collection is included for the live production instance. All endpoints are pre-configured to use the production base URL:

- **Base URL:** `https://league-simulator-282922766146.europe-west1.run.app`
- **Usage:** Import `collection.json` and select the "Production" environment to test live API endpoints.

**Quick Test Endpoints:**
```bash
# Get current standings
curl https://league-simulator-282922766146.europe-west1.run.app/standings

# Simulate one week
curl -X POST https://league-simulator-282922766146.europe-west1.run.app/simulate/week

# Get championship predictions (after week 4)
curl https://league-simulator-282922766146.europe-west1.run.app/predict
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

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

---

**🏆 Technical Case Study - Football League Simulator API**  
*Demonstrating Go backend development excellence with clean architecture principles*

For technical questions or code review feedback, please open an issue in this repository.
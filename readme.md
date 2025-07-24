# League Simulator

A comprehensive football league simulation API built with Go that allows you to simulate matches, track standings, predict championship probabilities, and manage league data.

## Features

- **Match Simulation**: Simulate individual weeks or entire seasons
- **Standings Management**: Real-time league table with points, goals, and statistics
- **Championship Predictions**: AI-powered predictions based on current form (available from week 4)
- **Match Result Editing**: Manually edit match results and recalculate standings
- **League Reset**: Reset the entire league to start fresh
- **RESTful API**: Clean HTTP endpoints for all operations

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
├── collection.json         # Postman collection for API testing
├── Dockerfile             # Docker configuration
├── go.mod                 # Go dependencies
└── main.go               # Application entry point
```

## API Endpoints

### Core Simulation
- `POST /simulate/week` - Simulate one week of matches
- `POST /simulate/all` - Simulate all remaining matches
- `POST /reset` - Reset league to initial state

### Data Access
- `GET /standings` - Get current league standings
- `GET /matches` - Get all matches (played and unplayed)
- `GET /predict` - Get championship predictions (available from week 4+)

### Match Management
- `POST /match/edit` - Edit a specific match result
  ```json
  {
    "match_id": 1,
    "home_goals": 2,
    "away_goals": 1
  }
  ```

## Getting Started

### Prerequisites
- Go 1.24+
- Docker (optional)
- Google Cloud CLI (for deployment)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd "League Simulator"
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the application**
   ```bash
   go run .
   ```

4. **Test the API**
   ```bash
   curl http://localhost:8080/standings
   ```

### Docker Deployment

1. **Build Docker image**
   ```bash
   docker build -t league-simulator .
   ```

2. **Run container**
   ```bash
   docker run -p 8080:8080 league-simulator
   ```

### Cloud Deployment (Google Cloud Run)

1. **Create Google Cloud project**
   ```bash
   gcloud projects create league-sim-project --set-as-default
   ```

2. **Enable required services**
   ```bash
   gcloud services enable run.googleapis.com artifactregistry.googleapis.com
   ```

3. **Create Artifact Registry repository**
   ```bash
   gcloud artifacts repositories create league-sim-repo --repository-format=docker --location=europe-west1
   ```

4. **Build and push image**
   ```bash
   docker build -t europe-west1-docker.pkg.dev/league-sim-project/league-sim-repo/league-simulator:latest .
   docker push europe-west1-docker.pkg.dev/league-sim-project/league-sim-repo/league-simulator:latest
   ```

5. **Deploy to Cloud Run**
   ```bash
   gcloud run deploy league-simulator \
     --image europe-west1-docker.pkg.dev/league-sim-project/league-sim-repo/league-simulator:latest \
     --platform managed \
     --region europe-west1 \
     --allow-unauthenticated \
     --port 8080
   ```

## Live Demo

🌐 **API Base URL**: `https://league-simulator-282922766146.europe-west1.run.app`

### Try it out:
- **Get Standings**: `GET https://league-simulator-282922766146.europe-west1.run.app/standings`
- **Simulate Week**: `POST https://league-simulator-282922766146.europe-west1.run.app/simulate/week`

## API Testing

Import the included `collection.json` file into Postman to test all endpoints with pre-configured requests.

### Example Requests

**Simulate one week:**
```bash
curl -X POST https://league-simulator-282922766146.europe-west1.run.app/simulate/week
```

**Get current standings:**
```bash
curl https://league-simulator-282922766146.europe-west1.run.app/standings
```

**Edit match result:**
```bash
curl -X POST https://league-simulator-282922766146.europe-west1.run.app/match/edit \
  -H "Content-Type: application/json" \
  -d '{"match_id":1,"home_goals":2,"away_goals":1}'
```

**Reset league:**
```bash
curl -X POST https://league-simulator-282922766146.europe-west1.run.app/reset
```

## Simulation Logic

### Match Simulation
- Teams have strength ratings that influence match outcomes
- Goals are randomly generated based on team strength
- Results affect standings automatically

### Championship Predictions
- Available from week 4 onwards
- Based on current points per game and form
- Considers remaining fixtures and strength differences
- Provides percentage probability for each team
- Week-based probability rules:
  - **Week 4-5**: Teams >6 points behind eliminated
  - **Week 6+**: Teams >3 points behind eliminated
  - **Advanced weeks**: Only leaders have winning chances

### Standings Calculation
- 3 points for win, 1 for draw, 0 for loss
- Sorted by: Points → Goal Difference → Goals For

## Database Schema

The project uses a simple SQL schema with three main tables:
- `teams` - Team information and strength ratings
- `matches` - Match fixtures and results
- `standings` - Current league standings

## API Response Examples

### Get Standings Response
```json
[
  {
    "team": {
      "id": 1,
      "name": "Manchester City",
      "strength": 90
    },
    "played": 5,
    "won": 4,
    "drawn": 1,
    "lost": 0,
    "goals_for": 12,
    "goals_against": 3,
    "goal_diff": 9,
    "points": 13
  }
]
```

### Championship Predictions Response
```json
{
  "championship_probabilities": [
    {
      "team_name": "Manchester City",
      "probability": 65.2
    },
    {
      "team_name": "Arsenal",
      "probability": 23.8
    }
  ],
  "message": "Championship winning probabilities based on current form"
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source and available under the MIT License.

---

**Built with ❤️ in Go**

For questions or support, please open an issue in the repository.
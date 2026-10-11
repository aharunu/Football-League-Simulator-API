# Football League Simulator API

A local Go demo for a four-team, six-round football league. Simulate matches, edit scores, view standings and estimate championship chances.

## Run locally

Requires Go 1.24.5 or newer. Uses only the Go standard library; no third-party dependencies.

```powershell
go run .
```

Open http://localhost:8080. The server listens on 127.0.0.1:8080. Stop it with Ctrl+C. If the port is occupied, startup reports the error.

All state is held in memory and resets on restart. The SQL files under `db/` are historical examples; the application does not connect to a database. No cloud service is required. `LISTEN_ADDR` can override the listening address. The retained Dockerfile uses the container interface; map its port to `127.0.0.1:8080:8080` for local use.

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/` | API name |
| GET | `/standings` | Table sorted by points, goal difference, goals scored, then team ID for stable display |
| GET | `/matches` | Fixtures grouped by round |
| POST | `/simulate/week` | Play the next unfinished round, skipping fully edited rounds; 410 when finished |
| POST | `/simulate/all` | Play all remaining matches |
| POST | `/match/edit` | Set or replace a result and update standings atomically |
| GET | `/predict` | Championship estimates after the first four complete rounds |
| POST | `/reset` | Clear scores and standings |

Edit requests require all three integer fields. Scores must be between 0 and 100. Missing/null fields, unknown fields, trailing JSON and invalid scores return 400; an unknown match returns 404. Error responses contain an `error` field.

```json
{"match_id": 1, "home_goals": 2, "away_goals": 1}
```

On Windows, use these examples:

```powershell
Invoke-RestMethod http://localhost:8080/standings
Invoke-RestMethod http://localhost:8080/simulate/week -Method Post
Invoke-RestMethod http://localhost:8080/match/edit -Method Post -ContentType 'application/json' -Body '{"match_id":1,"home_goals":2,"away_goals":1}'
```

Import `collection.json` into Postman. Its `baseUrl` variable defaults to `http://localhost:8080`; no former cloud URL is used. GET responses preserve the original Go field names (`Team`, `Points`, `HomeGoals`, etc.). Edit request fields use snake_case.

## Prediction limitations

Estimates are a simple heuristic, not calibrated probabilities. Current points per game are fixed at the time of prediction; a difference over 0.5 predicts a 1-0 win, otherwise a 0-0 draw. The shared predictor updates all table statistics consistently. Teams that cannot reach the current leader's points receive zero weight. Other teams receive weights based on predicted point gaps, normalized to exactly 100% in tenths.

At season end, points, goal difference and goals scored determine the winner. Teams tied on all three share the estimate equally; ID only determines display order. For leagues shorter than four rounds, prediction becomes available when every round is complete.

## Check changes

```powershell
go test ./...
go vet ./...
```

The service tests cover edited draws, skipped rounds, score validation, detached snapshots, fixture coverage, reset and concurrent reads/writes. HTTP tests cover malformed edit requests, response codes and prediction eligibility.

## Layout

- `main.go`: local HTTP server and initial teams
- `handlers/api.go`: endpoints and championship estimates
- `models/`: team, match and standing data
- `services/simulator.go`: fixtures and league state
- `services/predictor.go`: deterministic form-based predictions
- `collection.json`: local Postman collection
- `db/`, `Dockerfile`: retained historical/supporting files

# CLAUDE.md

## Project Overview

Slack Export is a Go utility that exports Slack workspace chat history into text files (optionally JSON and file attachments). It maintains state between runs so subsequent exports only retrieve new messages.

## Build & Run

```bash
make build               # Build binary
make test                # Run tests
make vet                 # Static analysis
make clean               # Remove binary
make docker              # Build Docker image
./slack-export           # Run (requires .env or env vars)
docker-compose up        # Run via Docker
```

## Configuration

Set environment variables in a `.env` file (see `sample.env`):
- `SLACK_API_TOKEN` (required) - Slack Bot User OAuth Token
- `EXPORT_ROOT` (required) - target output folder
- `EXPORT_JSON` - also export raw JSON payloads
- `EXPORT_FILES` - download file attachments
- `MIN_DATE_ISO` / `MAX_DATE_ISO` - date range filters (RFC3339 or YYYY-MM-DD)
- `RESET_CONF` - clear persisted state to force full re-export

## Project Structure

- `main.go` - Entry point, env var loading, validation
- `service.go` - `APIService` struct with all Slack API and export logic
- `model.go` - Data structs (`SimpleMsg`)
- `state.go` - JSON file-based state persistence (cross-platform via `os.UserConfigDir()`)

## Dependencies

- `github.com/slack-go/slack` - Slack API client
- `github.com/joho/godotenv` - .env file loading

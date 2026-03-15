# CLAUDE.md

## Project Overview

Slack Export is a TypeScript/Node.js utility that exports Slack workspace chat history into text files (optionally JSON and file attachments). It maintains state between runs so subsequent exports only retrieve new messages.

## Build & Run

```bash
npm run build        # Compile TypeScript to ./build
npm run prod         # Compile TypeScript to ./dist-prod (production)
npm run start:dev    # Dev mode with hot reload (nodemon + ts-node)
npm start            # Run production build
docker-compose up    # Run via Docker
```

## Lint

```bash
npm run lint         # ESLint on TypeScript files
```

There are no automated tests.

## Configuration

Set environment variables in a `.env` file (see `sample.env`):
- `SLACK_API_TOKEN` (required) - Slack Bot User OAuth Token
- `EXPORT_ROOT` (required) - target output folder
- `EXPORT_JSON` - also export raw JSON payloads
- `EXPORT_FILES` - download file attachments
- `MIN_DATE_ISO` / `MAX_DATE_ISO` - date range filters
- `RESET_CONF` - clear persisted state to force full re-export

## Project Structure

- `src/index.ts` - Entry point, config loading, initialization
- `src/api-service.ts` - Core `APIService` class for all Slack API interactions
- `src/model.ts` - TypeScript interfaces (User, Conversation, SlackMessage, etc.)
- `dist-prod/` - Checked-in production build output

## Code Style

- 4-space indentation, single quotes, semicolons required
- ESLint + Prettier enforced

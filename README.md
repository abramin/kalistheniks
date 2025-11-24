# Training App Backend — Stage 1

This repository contains the Stage 1 backend for a simple training and progression app.
The goal of this stage is to build a clean, minimal service that captures workouts, stores them in Postgres, and returns basic next-workout suggestions.

## Features

* User signup and login with JWT authentication
* Create training sessions
* Add sets (exercise, reps, weight) to a session
* List all sessions for the authenticated user
* Basic progression logic for the next workout (V1 rules)
* Structured Postgres schema with enums for body parts and muscles
* Local development environment using Docker Compose
* CI/CD preparation for later stages

## API Endpoints (Stage 1)

[Open API Documentation](https://kalistheniksio.stoplight.io/docs/kalistheniks-app/b92dfc4634655-training-api)

* `POST /signup`
* `POST /login`
* `POST /sessions`
* `POST /sessions/{id}/sets`
* `GET /sessions`
* `GET /plan/next`

These follow the project’s OpenAPI specification.

## Tech Stack

* **Language**: Go
* **Framework**: chi
* **Database**: PostgreSQL
* **Auth**: JWT-based access tokens

## Local Development

1. Start Postgres:

```bash
docker compose up -d
```

2. Run migrations:

```bash
make migrate-up
```
(Assumes you have golang-migrate installed)

3. Start the API:

```bash
go run ./cmd/api
```

4. Use the API with Postman, curl, or your preferred client.

## Seeding Exercises

A small set of barbell and foundational bodyweight exercises is included in `migrations/0004_seed_exercises.up.sql`.
These provide enough data to test the session and progression logic.

## Progression Logic (V1)

The initial progression rule is intentionally simple:

* If the user hits the upper end of the rep range, suggest a small load increase
* If they fail early, keep the load and reduce reps
* Alternate upper and lower sessions to maintain balance

This will be replaced by a more advanced engine in later stages.

## Next Steps

* Full CI/CD pipeline
* Terraform for AWS infrastructure
* API deployment on ECS Fargate
* Expanded exercise data
* Improved rule engine

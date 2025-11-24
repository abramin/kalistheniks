# Changelog

## v0.0.0

- Scaffolded the Go service with chi-based routing, JWT auth, sessions/sets endpoints, and a minimal progression plan service, generated with cursor-assisted workflows.
- Added PostgreSQL docker-compose setup and migrations for users, exercises, sessions, sets, enum columns, and placeholder seed migration.
- Introduced repositories/services layers with bcrypt authentication and simple plan suggestion logic.
- Scaffolded test suites (unit stubs with Testify and BDD feature files with Godog); test implementations will be written and reviewed manually.
- Rules engine will be authored solely by the maintainer; current code leaves it untouched for future work.

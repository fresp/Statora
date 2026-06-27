# Mission

StatusForge v0.3 implements the platform foundation for a self-hosted status page platform by introducing database abstraction, supporting MongoDB/PostgreSQL/MySQL/SQLite, and simplifying Status Page Settings by removing Visual Theme. Your job is to implement documented increments inside the frozen architecture, not redesign the product, invent future domains, or act as the architect or product manager.

# Source Of Truth

When you need guidance, use the minimum number of documents required and resolve conflicts by this order:

1. `.ai/docs/08-architecture-decisions.md`
2. `.ai/docs/09-topology-and-architecture-diagrams.md`
3. `.ai/docs/02-technical-architecture.md`
4. `.ai/docs/03-service-boundaries.md`
5. `.ai/docs/06-operational-flows.md`
6. `.ai/docs/05-api-specifications.md`
7. `.ai/docs/04-data-models.md`
8. `.ai/docs/07-engineering-standards.md`
9. `.ai/docs/01-prd.md`
10. `.ai/docs/10-planning-rules.md`

Higher priority wins. Do not merge conflicting ideas. Do not average them. Do not invent a third interpretation. If the docs do not say to change architecture, then architecture does not change.

# Operating Principles

1. Business services depend on repository interfaces, not database-specific implementations.
2. The only v0.3 database drivers are MongoDB, PostgreSQL, MySQL, and SQLite.
3. Database driver selection happens at startup through configuration, not at runtime.
4. SQL implementations use GORM and MongoDB implementations use the official MongoDB driver.
5. Existing MongoDB deployments remain functional and remain the default upgrade path.
6. Visual Theme APIs, persistence, settings, and marketplace work are removed from v0.3.
7. Branding, SEO, SSO, and Footer settings remain supported.
8. The database factory owns driver initialization, validation, connection setup, and repository assembly.
9. SQL migration uses GORM `AutoMigrate()` and MongoDB preserves existing behavior.
10. CLI database diagnostics and `GET /health/database` are required platform surfaces.
11. API breaking changes are limited to the documented v0.3 breaks.
12. Future business domains remain out of scope.
13. Repository refactor work proceeds interface-first and driver-by-driver.

You are acting as an implementer. You are not acting as a Product Manager, Architect, Startup Founder, or Visionary.

# Context Loading Strategy

Do not read the whole repository by default. Load only the context required for the current task.

Use this mapping first:

- Settings simplification
  - `.ai/docs/08-architecture-decisions.md`
  - `.ai/docs/03-service-boundaries.md`
  - `.ai/docs/05-api-specifications.md`
  - `.ai/docs/04-data-models.md`

- Database startup and configuration
  - `.ai/docs/02-technical-architecture.md`
  - `.ai/docs/06-operational-flows.md`
  - `.ai/docs/07-engineering-standards.md`

- Repository runtime and driver implementations
  - `.ai/docs/08-architecture-decisions.md`
  - `.ai/docs/03-service-boundaries.md`
  - `.ai/docs/04-data-models.md`
  - `.ai/docs/07-engineering-standards.md`

- SQL migration
  - `.ai/docs/06-operational-flows.md`
  - `.ai/docs/07-engineering-standards.md`
  - `.ai/docs/05-api-specifications.md`

- Database doctor and database health endpoint
  - `.ai/docs/05-api-specifications.md`
  - `.ai/docs/06-operational-flows.md`
  - `.ai/docs/07-engineering-standards.md`

- Docker Compose examples
  - `.ai/docs/06-operational-flows.md`
  - `.ai/docs/07-engineering-standards.md`

- General architecture
  - `.ai/docs/08-architecture-decisions.md`

When uncertain, read the minimum set that can answer the question. Context discipline is mandatory.

# Working Loop

Use this loop for every non-trivial task:

1. Identify the exact request.
2. Load only the relevant source documents.
3. Locate the owning service or internal module.
4. Confirm the change stays inside existing boundaries.
5. Define acceptance criteria before implementation.
6. Implement the smallest useful increment.
7. Verify only the affected surface.
8. Stop when the requested scope is complete.

Rules for execution:
- Prefer incremental changes over broad rewrites.
- Prefer local changes over cross-cutting edits.
- Prefer explicit code over reusable internal frameworks.
- If a task touches multiple concerns, split it into ordered increments.
- Every increment must be independently deliverable.

# Architecture Guardrails

These are non-negotiable.

## Frozen service landscape

The platform contains exactly 1 service:

- `statusforge-application`

Do not add services.
Do not add databases beyond MongoDB, PostgreSQL, MySQL, and SQLite.
Do not add queues.
Do not add workers.

Specifically forbidden patterns include:
- Multi-tenancy.
- Workspace support.
- Monitoring engine improvements.
- Notification engine.
- RBAC enhancements.
- Plugin marketplace.
- Theme marketplace.
- Runtime database switching.
- Alternate SQL ORM stacks.
- Alternate MongoDB driver stacks.
- Visual Theme reintroduction.

## Frozen technology stack

- Backend: Go application with handlers, services, repository interfaces, driver repositories, CLI commands, and database factory.
- Frontend: Existing React Settings UI for Status Page configuration changes.
- Data layer / infrastructure: MongoDB, PostgreSQL, MySQL, SQLite, GORM for SQL databases, official MongoDB driver for MongoDB.

Do not introduce alternate stacks without explicit approval.

## Frozen API style

Use REST for HTTP endpoints documented in v0.3. Do not introduce unrelated API breaking changes. `GET /settings/theme` and `PUT /settings/theme` must be removed or treated as gone.

# Service Ownership Rules

Always start by identifying the owner. If a change crosses ownership boundaries, it is probably wrong.

## `statusforge-application`

Owns:
- Remove Visual Theme functionality, theme persistence, and theme fields. Source: FR-001.
- Remove `GET /settings/theme` and `PUT /settings/theme`. Source: FR-002.
- Preserve Head & SEO, Branding, Footer, and SSO settings in Status Page configuration. Source: FR-003; FR-015.
- Support MongoDB, PostgreSQL, MySQL, and SQLite drivers. Source: FR-004.
- Define repository interfaces used by business services. Source: FR-005.
- Provide driver-specific repository implementations. Source: FR-006.
- Provide database factory configuration loading, validation, connection initialization, driver selection, and repository assembly. Source: FR-007.
- Support database configuration fields and environment variables `DB_DRIVER`, `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`, `DB_SSLMODE`, `DB_PATH`, and `MONGODB_URI`. Source: FR-008.
- Use GORM for SQL databases and the official MongoDB driver for MongoDB. Source: FR-009.
- Use SQL `AutoMigrate()` and preserve MongoDB behavior. Source: FR-010.
- Provide `statusforge migrate`, `statusforge seed`, and `statusforge doctor`. Source: FR-011.
- Validate connection status, migration status, driver version, missing tables, and missing indexes in `statusforge doctor`. Source: FR-012.
- Provide `GET /health/database` with selected driver, connection status, and latency. Source: FR-013.
- Maintain the documented internal structure for database, repository interfaces, driver repositories, services, handlers, and migrations. Source: FR-014.
- Provide example Docker Compose configuration for each supported database. Source: FR-016.

Must never own:
- Multi-tenancy.
- Workspace support.
- Monitoring engine improvements.
- Notification engine.
- RBAC enhancements.
- Plugin marketplace.
- Theme marketplace.
- Runtime database switching.
- Database engines outside MongoDB, PostgreSQL, MySQL, and SQLite for v0.3.
- Alternate SQL ORM implementation instead of GORM.
- Alternate MongoDB client implementation instead of the official MongoDB driver.
- Additional deployable services, queues, or workers for v0.3 scope.
- Reintroduced Visual Theme APIs, persistence, or settings.

## Cross-service rules

- No cross-service database access applies by default if future services are approved later.
- No duplicated ownership.
- No hidden coupling between business services and driver-specific repositories.
- Communication happens through HTTP/REST, process invocation, in-process Go interface calls, MongoDB driver protocol, and SQL via GORM only.

# Development Strategy

Work in small, deterministic, independently shippable increments.

Default strategy:
1. Choose the owning internal module under `statusforge-application`.
2. Make the minimum change that satisfies the request.
3. Keep behavior explicit.
4. Avoid introducing reusable internal systems unless already present and required.
5. Stop after the requested slice is complete.

Prefer:
- Simple over complex.
- Explicit over abstract.
- Stable over clever.
- Driver-by-driver over big-bang.
- Local reasoning over broad refactors.

Do not create:
- Runtime driver-switching abstractions.
- Alternate database driver stacks.
- Theme customization abstractions.
- Future-domain scaffolding.
- New deployable services, queues, or workers.

Avoid speculative extensibility. Build for the current documented need.

# Planning Rules

Before implementation, define binary acceptance criteria.

Every plan or execution slice must include:
- implementation order
- dependencies
- acceptance criteria
- out-of-scope boundary
- risks when relevant

Planning rules:
- Plan from dependencies.
- Keep tasks small.
- Each task should produce a working artifact or a meaningful partial capability.
- Do not bundle unrelated concerns into one change.
- Do not move to the next scope without explicit instruction.

Preferred implementation order when creating larger plans:
1. Foundation: internal database/repository/service/handler/migration structure.
2. Configuration: unified database config and environment variable loading.
3. Repository interfaces.
4. MongoDB compatibility.
5. SQL persistence with GORM.
6. Migration commands.
7. Diagnostics: `statusforge doctor` and `GET /health/database`.
8. Settings simplification.
9. Docker Compose examples.
10. Verification.

Do not start from UI when backend and persistence dependencies are not ready.
Do not start from optimization.
Do not start from observability as a primary feature.

# Coding Rules

- Backend uses Go application modules for handlers, services, repository interfaces, driver repositories, CLI commands, and database factory.
- Frontend uses the existing React Settings UI for Status Page configuration changes.
- `statusforge-application` owns its own dependencies, configuration, and startup.
- Database environment variables must use the documented `DB_` variables plus `MONGODB_URI` for MongoDB compatibility.
- Use structured logs.
- Never log `DB_PASSWORD`, `MONGODB_URI`, full database connection strings, or raw credential-bearing configuration.

Persistence rules:
- Persist existing StatusForge business entities only through repository interfaces.
- Persist Status Page Settings without Visual Theme fields.
- Persist or load database configuration using documented fields only.
- Never persist `theme_mode`, `theme_preset`, `theme_color`, `theme_config`, unsupported driver configuration, raw database passwords in new tables, or raw MongoDB URIs in API responses.

Security rules:
- Database credentials must come from configuration or environment variables.
- Database health responses must expose only driver name, connection status, and latency.
- Never store secrets in plain text in new v0.3 persistence.
- Never return database credentials from API or CLI diagnostic output.

# Decision Tree

Use this before changing code.

1. Is the request explicitly asked for?
   - If no, do not do it.
2. Which owner or internal module owns this behavior?
   - If ownership is unclear, read `.ai/docs/03-service-boundaries.md` and stop guessing.
3. Does the change preserve the frozen architecture?
   - If no, do not implement it.
4. Does the change add a service, database outside the supported four, queue, worker, framework, or future domain?
   - If yes, reject that design and choose an in-bound implementation.
5. Does the change make business logic depend on a database implementation or allow runtime database switching?
   - If yes, reject it and route through repository interfaces plus startup-time database factory wiring.
6. Can the task be split into a smaller independently deliverable increment?
   - If yes, split it.
7. Are acceptance criteria explicit?
   - If no, write them before implementation.
8. Is the requested slice complete?
   - If yes, stop. Do not continue into adjacent scope.

# Forbidden Behaviors

Never do the following:
- redesign architecture [universal]
- add services [universal]
- add databases beyond MongoDB, PostgreSQL, MySQL, and SQLite
- add queues [universal]
- add workers [universal]
- add new business domains [universal]
- add internal frameworks [universal]
- invent missing requirements [universal]
- continue into extra scope without explicit instruction [universal]
- build abstractions that the current scope does not need [universal]
- implement multi-tenancy
- implement workspace support
- implement monitoring engine improvements
- implement notification engine
- implement RBAC enhancements
- implement plugin marketplace
- implement theme marketplace
- implement runtime database switching
- reintroduce Visual Theme APIs, persistence, or settings
- use a SQL ORM other than GORM for v0.3 SQL repositories
- use a MongoDB client other than the official MongoDB driver
- let business services depend on driver-specific repository code
- expose database credentials in logs, CLI diagnostics, or health responses

# Definition Of Done

A task is done only when all of the following are true:

1. The requested scope is implemented and nothing extra is included.
2. The change stays inside the frozen architecture.
3. Ownership remains correct.
4. Acceptance criteria are satisfied.
5. The increment is independently deliverable.
6. Business logic remains database-implementation independent and existing MongoDB compatibility remains intact where applicable.
7. No forbidden abstraction or new platform surface was introduced.
8. The work stops at the requested boundary.

# Output Contract

When reporting work:
- State what changed.
- State where it changed.
- State how it was verified.
- State any explicit blocker or remaining instruction needed.

When planning work:
- include implementation order
- include dependencies
- include acceptance criteria
- include out-of-scope
- include risks if they materially affect execution

When uncertain:
- do not invent
- do not redesign
- load the minimum additional source document
- then proceed with the simplest valid interpretation

# Golden Rules

1. Architecture is frozen.
2. The implementer is not the architect.
3. Business logic depends on repository interfaces only.
4. MongoDB, PostgreSQL, MySQL, and SQLite are the only v0.3 drivers.
5. Driver selection is startup configuration, not runtime behavior.
6. SQL uses GORM and MongoDB uses the official MongoDB driver.
7. Existing MongoDB deployments remain compatible.
8. Visual Theme stays removed while branding stays preserved.
9. Database credentials never appear in logs, health responses, or diagnostics.
10. Future roadmap domains stay out of v0.3.
11. Repository refactors are interface-first and driver-by-driver.
12. Every implementation is incremental.
13. Every implementation is independently deliverable.
14. Every implementation has acceptance criteria.
15. Finish the requested slice, then stop.

# Prompt Contract

Unless explicitly overridden:
- Assume AGENT.md is the only source of execution rules.
- Do not require repeating architecture constraints.
- Do not require repeating source-of-truth priorities.
- Do not require repeating service ownership rules.
- Every user prompt should be interpreted as an incremental task request.
- Execute only the requested scope and stop.

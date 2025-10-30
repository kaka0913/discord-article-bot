<!--
Sync Impact Report:
Version: 0.0.0 → 1.0.0
Rationale: Initial constitution adoption for discord-article-bot project

Modified Principles:
- All principles newly defined (initial version)

Added Sections:
- Core Principles: 6 principles covering bot architecture, testing, deployment, reliability, security, and simplicity
- Development Standards: Error handling, logging, and code quality requirements
- Operational Requirements: Monitoring, performance, and deployment practices
- Governance: Constitution versioning and compliance process

Removed Sections: None (initial version)

Templates Status:
- ✅ plan-template.md: Constitution Check section reviewed - compatible
- ✅ spec-template.md: User scenarios and requirements sections align with principle-driven development
- ✅ tasks-template.md: Task organization supports bot architecture and testing principles
- ⚠ commands/*.md: No command files found yet - will need alignment when created

Follow-up TODOs:
- RATIFICATION_DATE set to today (2025-10-27) as initial adoption
- Consider adding command files for workflow automation in future
-->

# Discord Article Bot Constitution

## Core Principles

### I. Bot-First Architecture

The Discord article bot is designed as a single-purpose service with clear separation of concerns:

- **Command handlers** MUST be isolated, independently testable modules
- **Discord gateway interactions** MUST be abstracted behind interfaces to enable testing without live connections
- **Article processing logic** MUST be library-style components that can be tested without Discord infrastructure
- **State management** MUST be explicit - no hidden global state except for properly managed caches
- **Configuration** MUST be externalized via environment variables or config files, never hardcoded

**Rationale**: Bot architecture directly impacts testability, reliability, and maintainability. Clear boundaries prevent the "big ball of mud" pattern common in bot projects.

### II. Contract-Driven Integration Testing

Integration with external services (Discord API, article sources, databases) requires contract testing:

- **Discord API interactions** MUST have contract tests verifying request/response shapes
- **Article fetching** MUST have integration tests against real or recorded HTTP responses
- **Database operations** MUST have integration tests using real database engines (not mocks)
- **Rate limiting** MUST be testable via time-controlled test scenarios
- **Webhook/callback handlers** MUST have contract tests verifying payload handling

**Rationale**: Discord bots fail in production due to API changes, malformed payloads, and race conditions. Contract tests catch these before deployment.

### III. Graceful Deployment & Zero-Downtime Updates

Bot deployment must not disrupt active user interactions:

- **Command processing** MUST complete in-flight requests before shutdown (graceful drain)
- **State persistence** MUST survive restarts - ephemeral state documented and justified
- **Version migrations** MUST be backward compatible for at least one version overlap
- **Health checks** MUST report ready/live status accurately for orchestration
- **Rollback capability** MUST be maintained - database migrations reversible, feature flags available

**Rationale**: Discord users expect reliability. Ungraceful restarts create poor UX and data loss.

### IV. Reliability Under Failure

External dependencies (Discord API, article sources, databases) will fail - the bot must handle this gracefully:

- **Retry logic** MUST use exponential backoff with jitter for transient failures
- **Circuit breakers** MUST be implemented for failing downstream services
- **Fallback behavior** MUST be defined for degraded modes (e.g., cached responses, error messages)
- **Timeout enforcement** MUST prevent hanging requests from blocking the bot
- **Error messages to users** MUST be informative without exposing internal details

**Rationale**: Production reliability requires defensive programming against external failures.

### V. Security & Rate Limiting

Discord bots are exposed to untrusted user input and must protect against abuse:

- **Input validation** MUST sanitize all user commands, preventing injection attacks
- **Rate limiting** MUST be enforced per-user and per-guild to prevent abuse
- **Secrets management** MUST use environment variables or secure vaults, never committed to git
- **Permission checking** MUST verify user/role permissions before privileged operations
- **Audit logging** MUST record security-relevant events (auth failures, rate limit hits, permission denials)

**Rationale**: Discord bots are attack surfaces. Security must be built in from day one.

### VI. Simplicity & Maintainability

Bot projects tend toward complexity - actively resist this:

- **Start simple**: Implement minimal viable features first, add complexity only when justified
- **Avoid premature abstraction**: Don't create frameworks until patterns emerge from 3+ use cases
- **Prefer Go standard library**: Use third-party dependencies only when significant value added
- **YAGNI principle**: "You Aren't Gonna Need It" - don't build for hypothetical future requirements
- **Code comments** MUST explain "why" not "what" - the code itself should be self-documenting

**Rationale**: Overengineering kills bot projects. Keep it simple until proven complexity is needed.

## Development Standards

### Error Handling

- All errors MUST be returned and handled explicitly (no silent failures)
- Error messages MUST include context for debugging (what operation failed, with what inputs)
- Errors exposed to users MUST be friendly and actionable
- Panic/recover MUST only be used for truly unrecoverable errors (not control flow)

### Logging

- Structured logging MUST be used for machine-readable output (JSON format)
- Log levels MUST be semantic: DEBUG (development), INFO (state changes), WARN (recoverable issues), ERROR (failures)
- Sensitive data (tokens, user IDs) MUST be redacted or excluded from logs
- Logs MUST include correlation IDs for tracing requests across components

### Code Quality

- Go code MUST pass `go vet`, `golint`, and `staticcheck` without warnings
- Test coverage MUST be tracked, with minimum 70% coverage for core business logic
- Code reviews MUST verify: security, error handling, testing, documentation
- Breaking changes MUST be documented and approved before merge

## Operational Requirements

### Monitoring & Observability

- Metrics MUST be exposed for: command latency, error rates, active connections, rate limit hits
- Health check endpoint MUST report bot readiness (connected to Discord, dependencies healthy)
- Alerting MUST be configured for: error rate spikes, connectivity loss, resource exhaustion

### Performance

- Command responses MUST be sent within 3 seconds or use "thinking" indicators for longer operations
- Memory usage MUST be bounded - no unbounded caches or queues
- Database queries MUST use connection pooling and prepared statements
- API rate limits MUST be respected with appropriate backoff

### Deployment

- Deployment MUST be automated via CI/CD pipeline
- Rollback MUST be executable within 5 minutes
- Configuration MUST be versioned and reviewed like code
- Production deployment MUST follow canary or blue-green pattern for risk mitigation

## Governance

This constitution supersedes all other development practices and must be referenced during:

- Feature specification reviews
- Code reviews and pull request approvals
- Architecture decision records
- Post-incident retrospectives

### Amendment Process

1. Propose amendment with rationale and impact analysis
2. Document in constitution with version bump (see versioning below)
3. Update affected templates, specs, and documentation for consistency
4. Obtain team approval before adoption
5. Communicate changes to all stakeholders

### Versioning

This constitution follows semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Backward-incompatible governance changes (e.g., removing a principle, redefining core requirements)
- **MINOR**: New principles added or existing principles materially expanded
- **PATCH**: Clarifications, typo fixes, non-semantic improvements

### Compliance

- All pull requests MUST verify compliance with this constitution
- Complexity additions MUST be justified in the plan.md "Complexity Tracking" section
- Constitution violations MUST block PR approval unless explicitly approved with documented rationale

**Version**: 1.0.0 | **Ratified**: 2025-10-27 | **Last Amended**: 2025-10-27

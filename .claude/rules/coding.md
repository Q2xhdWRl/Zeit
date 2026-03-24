# Coding Principles & Structure

## Core Principles

### SOLID
- **Single Responsibility**: Each function/class has one well-defined purpose. If a function exceeds ~40 lines, decompose it.
- **Open/Closed**: Design modules to be extendable without modifying existing code.
- **Liskov Substitution**: Subtypes must be substitutable for their base types without breaking behavior.
- **Interface Segregation**: Prefer small, specific interfaces over large, general ones.
- **Dependency Inversion**: Depend on abstractions, not concrete implementations.

### DRY (Don't Repeat Yourself)
- Extract duplicated logic into shared utilities. Centralize security controls (validation, sanitization, auth).
- But avoid premature abstraction — require at least two concrete use cases before extracting.

### YAGNI (You Aren't Gonna Need It)
- Only implement what is currently needed. Do not build for hypothetical future requirements.
- Remove dead code, unused imports, and commented-out blocks.

## Code Quality

- **Readability first**: Code must be understandable by someone who didn't write it.
- **Explicit over implicit**: Name things so their purpose is evident without context.
- **Guard clauses**: Use early returns to reduce nesting (max 3 levels deep).
- **No boolean flag params**: Use separate, clearly named functions instead.
- **No clever one-liners**: Avoid unreadable regex, ternary chains, or bitwise tricks without justification.

## Naming

- Descriptive, unambiguous names. Follow the project's existing convention.
- Constants: `UPPER_SNAKE_CASE`, centrally defined.
- Security functions named clearly: `validateAuthToken`, `sanitizeUserInput`.

## Error Handling

- Never swallow errors silently — no empty `catch` blocks.
- Log errors with context, recover gracefully or propagate with clear messages.
- Security errors (auth failures, validation) must always be logged.

## Defensive Coding

- Set timeouts on all external calls. No unbounded waits, loops, or memory allocations.
- Resource cleanup in `finally`/`defer`/`using` blocks.
- Handle partial failures — leave system in consistent state.

## Type Safety

- Full type annotations in static languages. Type hints/JSDoc in dynamic languages.
- Avoid `any` — if unavoidable, comment why.

## File Structure

- Respect existing project layout. Ask if unsure where a file belongs.
- One module/class per file. Separate concerns (logic, data access, API, config).
- Security logic centralized in designated directories.

## Frontend Accessibility (WCAG 2.1 AA)

- Semantic HTML (`<button>`, `<nav>`, `<main>`, `<label>`), not `<div>` with click handlers.
- ARIA attributes where semantic HTML is insufficient.
- Keyboard-accessible interactive elements. Color never the sole indicator.
- `alt` text on images. `<label>` on form inputs.

## Workflow

- Plan before coding (for non-trivial tasks). State requirements, approach, risks, assumptions.
- Generate code in small, reviewable increments.
- If ambiguous, stop and ask — don't guess.
- Match existing codebase style, patterns, and conventions.

# Testing

## Framework: Vitest

- Use **Vitest** as the testing framework for all frontend/TypeScript code.
- Follow existing test conventions (`.test.ts` / `.spec.ts`). Mirror source file structure in test directories.

## Integration Testing (Priority)

- Prefer **integration tests** over unit tests. Test real interactions between components, services, and data layers.
- Use real database connections and API calls in integration tests where feasible — avoid mocks for core behavior.
- Mock only external third-party services and non-deterministic dependencies (time, randomness).

## Test Priorities

1. **Security tests**: Auth enforcement, authorization, input validation, no data leaks.
2. **Stability tests**: Failure conditions (timeouts, unavailable services, malformed input), resource cleanup, error messages.
3. **Functional tests**: Happy path, business logic, edge cases.

## Test Quality

- Every test must have meaningful assertions. No tests that just call a function and assert `true`.
- Tests must not be tightly coupled to implementation details.
- Never skip, disable, or weaken existing tests.
- Every new source file must have a corresponding test file.
- Generate tests alongside production code, not as an afterthought.

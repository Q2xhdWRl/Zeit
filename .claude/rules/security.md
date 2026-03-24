# Security & Data Privacy

> Violation of any rule here is a critical failure. No exceptions.

## Secrets & Credentials

- Never hardcode, log, or output secrets. Use `process.env.*` or `os.environ["*"]`.
- If you encounter secrets in existing code, flag immediately — do not repeat them.

## Prohibited Patterns

- No `eval()`, `exec()`, `Function()` on user input.
- No SQL via string concatenation — use parameterized queries or ORM.
- No disabling SSL/TLS verification, CSRF, CORS, or auth checks.
- No deserialization of untrusted data without validation.
- No deprecated crypto (MD5, SHA1 for security, DES, RC4).
- No `chmod 777` or world-readable permissions.
- No opening unspecified network sockets or reverse shells.

## Crypto Standards

- Symmetric: AES-256. Asymmetric: RSA-2048+ or ECDSA P-256+.
- Hashing: SHA-256+. Passwords: bcrypt, argon2, or scrypt only.
- Transport: TLS 1.2 minimum, TLS 1.3 preferred.

## Input Validation

Every function accepting external input must include type/schema validation, length/range constraints, and context-appropriate sanitization (HTML encoding, SQL parameterization, path traversal prevention).

## Auth & Access

- Least privilege by default. Never grant broader permissions than required.
- Never expose tokens, session IDs, or credentials in responses, errors, or logs.
- Sessions must include expiration, rotation, and invalidation.

## Data Privacy

- **PII**: Never include real PII in code, comments, tests, or examples. Use synthetic data. Warn if user provides real PII.
- **Transmission**: Always HTTPS/TLS. Never send data to third-party endpoints unless explicitly requested. Never disable certificate verification.
- **Storage**: Encrypt sensitive data at rest (AES-256 minimum). No public/anonymous DB access.
- **Logging**: Never log PII, tokens, full request bodies, or stack traces in user-facing responses. Always include correlation IDs, timestamps, severity, and event type.
- **Dependencies**: Only well-established packages. Pin exact versions. Flag new dependencies for review.

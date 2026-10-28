# Automated Code Review Checklist

**Purpose**: AI runs these checks automatically before requesting human review. This catches 70% of common issues.

---

## Pre-Review Requirements

- [ ] Branch is up to date with main
- [ ] All files are saved
- [ ] No syntax errors

---

## Build & Test Quality

### Compilation
- [ ] Code compiles without errors
- [ ] Code compiles without warnings
- [ ] No deprecated API usage

### Testing
- [ ] All existing tests pass
- [ ] New tests added for new code
- [ ] Test coverage ≥ 80% for changed files
- [ ] Edge cases have tests
- [ ] Error paths have tests

### Linting
- [ ] Passes linter (golangci-lint / flake8)
- [ ] No spelling errors in comments
- [ ] No commented-out code

---

## Code Standards

Reference: `standards/` directory for detailed conventions.

### Naming
- [ ] Variables: clear, descriptive names
- [ ] Functions: verb phrases (e.g., `getUserById`)
- [ ] Constants: descriptive, not magic numbers
- [ ] No single-letter names (except i, j in loops)

### Error Handling
- [ ] All errors are checked (no `_` discards)
- [ ] Error messages are descriptive
- [ ] Errors wrapped with context
- [ ] No bare `panic()` or uncaught exceptions

### Comments
- [ ] Complex logic has comments explaining WHY
- [ ] Public functions have documentation
- [ ] No obvious/redundant comments
- [ ] TODOs have issue numbers

### Formatting
- [ ] Consistent indentation
- [ ] No trailing whitespace
- [ ] Line length reasonable (<120 chars)
- [ ] Imports organized correctly

---

## Security

### Secrets & Credentials
- [ ] No hardcoded passwords/tokens/keys
- [ ] No credentials in comments
- [ ] Sensitive config in environment variables

### Input Validation
- [ ] All user input is validated
- [ ] SQL queries use parameterized statements
- [ ] No direct string concatenation in queries
- [ ] File paths are sanitized

### Data Safety
- [ ] No passwords/tokens in logs
- [ ] PII is not logged in plain text
- [ ] Sensitive data encrypted at rest/transit

---

## Best Practices

### Code Quality
- [ ] No code duplication (DRY principle)
- [ ] Functions are focused (single responsibility)
- [ ] No deeply nested logic (max 3-4 levels)
- [ ] No functions >50 lines (consider refactoring)

### Dependencies
- [ ] No unnecessary dependencies added
- [ ] Dependencies are pinned to versions
- [ ] Lock files updated (go.sum / requirements.txt)

### Production Readiness
- [ ] No TODO/FIXME/HACK in production code
- [ ] No debug print statements
- [ ] No disabled tests
- [ ] Configuration is externalized

---

## Language-Specific Checks

### Go
- [ ] `defer` used for cleanup (Close, Unlock)
- [ ] No goroutine leaks
- [ ] Context passed to long-running operations
- [ ] Errors use `%w` for wrapping

### Python
- [ ] Type hints present on public functions
- [ ] No mutable default arguments
- [ ] Context managers used (`with` statement)
- [ ] Async functions use `await` properly

---

## Output Format

### If all checks pass:
```
✅ Automated review passed. Ready for human review.

Summary:
- Build: ✅ Compiled successfully
- Tests: ✅ All tests pass (coverage 87%)
- Lint: ✅ No issues
- Security: ✅ No secrets detected
- Standards: ✅ Follows coding conventions
```

### If checks fail:
```
❌ Automated review found issues:

## Build Errors (P0)
- file.go:42: syntax error

## Test Failures (P0)
- TestUserLogin failed: expected 200, got 500

## Lint Issues (P1)
- file.go:15: variable 'userId' should be 'userID'

## Security Issues (P0)
- config.go:10: hardcoded password detected

⚠️ Fix these issues before requesting human review.
```

---

## Automation

This checklist should be run:
1. **On save** - syntax checks
2. **Before commit** - full checks via pre-commit hook
3. **In PR** - CI/CD runs complete suite

**Command:**
```bash
make lint && make test
```

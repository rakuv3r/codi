# Human Code Review Guide

**Purpose**: Deep manual review after AI automated checks pass. Focus on what AI cannot judge.

---

## Review Process

### Step 1: Understand Context (5 min)
- Read the issue/requirement
- Understand the problem being solved
- Check the proposed solution approach
- Review any design docs

### Step 2: High-Level Review (10 min)
- Does the solution fit the architecture?
- Are there better approaches?
- Is the scope appropriate (not over-engineered)?
- Will this scale?

### Step 3: Line-by-Line Review (30-60 min)

**For each file:**

1. **Read entire file first** - understand structure and flow
2. **Line by line** - scrutinize each line:
   - Does this line make sense?
   - Any edge cases missed?
   - Performance implications?
   - Security risks?
   - Concurrency issues?
3. **Check interactions** - how do functions call each other?
4. **Verify error paths** - what happens when things fail?

**Per-line questions:**
- Variable naming: accurate and clear?
- Logic: correct for ALL inputs?
- Error handling: complete and informative?
- Resource management: acquired and released properly?
- Concurrency: safe access to shared state?

### Step 4: Provide Feedback
- Comment on **specific lines** with line numbers
- Explain **WHY** it's a problem (not just "this is wrong")
- Suggest **concrete fix** (not vague "improve this")
- Use severity: P0 (must fix), P1 (should fix), P2 (nice to have)

---

## P0 Critical Items (Must Check Every Line)

### 1. Concurrency Safety

**Look for:**
- Shared data accessed without locks
- Race conditions
- Deadlocks (lock ordering)
- Goroutine/thread leaks

**Example:**
```go
// ❌ REJECT - race condition
type Counter struct {
    count int
}
func (c *Counter) Inc() {
    c.count++  // unprotected write
}

// ✅ APPROVE - properly synchronized
type Counter struct {
    mu    sync.Mutex
    count int
}
func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

### 2. Resource Leaks

**Look for:**
- Files not closed
- Connections not released
- Goroutines not stopped
- Memory not freed

**Example:**
```go
// ❌ REJECT - file leaked on error
func process(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    if err := validate(f); err != nil {
        return err  // f never closed!
    }
    f.Close()
    return nil
}

// ✅ APPROVE - defer ensures cleanup
func process(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    return validate(f)
}
```

### 3. Security Vulnerabilities

**Look for:**
- SQL injection
- Command injection
- Path traversal
- XSS vulnerabilities
- Unvalidated input
- Exposed secrets

**Example:**
```go
// ❌ REJECT - SQL injection
query := "SELECT * FROM users WHERE id=" + userId
rows, _ := db.Query(query)

// ✅ APPROVE - parameterized query
query := "SELECT * FROM users WHERE id=?"
rows, _ := db.Query(query, userId)
```

### 4. Logic Errors

**Look for:**
- Off-by-one errors
- Wrong comparison operators
- Missing nil checks
- Incorrect default values
- Unreachable code

**Example:**
```go
// ❌ REJECT - off-by-one
for i := 0; i <= len(items); i++ {
    process(items[i])  // panic on last iteration!
}

// ✅ APPROVE
for i := 0; i < len(items); i++ {
    process(items[i])
}
```

---

## P1 Strongly Recommended

### Architecture & Design
- Does this fit the system design?
- Is abstraction level appropriate?
- Are responsibilities clear?
- Is it maintainable?

### Performance
- N+1 database queries?
- Unnecessary allocations in hot path?
- Blocking operations in critical path?
- Will this scale to production load?

### Error Handling
- All error paths covered?
- Error messages helpful for debugging?
- Errors propagated correctly?
- Logging appropriate?

### Testing
- Right tests for the change?
- Edge cases covered?
- Error scenarios tested?
- Test quality good (not just coverage)?

---

## Line-by-Line Checklist

### For Every Function
- [ ] Name describes what it does
- [ ] All parameters validated
- [ ] All errors handled
- [ ] All resources cleaned up (defer)
- [ ] No data races
- [ ] Edge cases covered
- [ ] Can be tested

### For Every if/for/switch
- [ ] All branches handled
- [ ] No off-by-one errors
- [ ] Loop termination guaranteed
- [ ] Default case present (switch)

### For Every Variable
- [ ] Name is clear and accurate
- [ ] Type is appropriate
- [ ] Initialization is correct
- [ ] Lifetime is managed (no use-after-free)

---

## Common Issues to Catch

### Logic Errors
- Off-by-one in loops (`<=` vs `<`)
- Wrong comparison operator
- Missing nil/null checks
- Incorrect boolean logic
- Wrong default values

### Performance
- N+1 queries (loop with query inside)
- Unnecessary allocations in loops
- Blocking calls in hot path
- Missing indexes on queries

### Maintainability
- Magic numbers (use named constants)
- Duplicated code (DRY)
- Too many nested levels (>3-4)
- Function too long (>50 lines)
- Poor naming

---

## Review Mindset

### Think Like an Attacker
- What if input is empty/negative/huge?
- What if this fails halfway through?
- What if two requests run concurrently?
- What if user is malicious?

### Think Like a Maintainer
- Will I understand this in 6 months?
- Can I debug this easily?
- Is this testable?
- Will this break when requirements change?

### Think Like a User
- What happens when this fails?
- Is error message helpful?
- Will this be fast enough?
- Is data safe?

---

## When to REJECT

**Must reject if:**
- Any P0 issue present
- Logic clearly wrong
- Security vulnerability exists
- Will cause production incident
- Insufficient test coverage

**Consider rejecting if:**
- Multiple P1 issues
- Code is unmaintainable
- Performance is unacceptable
- Design doesn't fit architecture

---

## When to APPROVE

**Approve when:**
- All P0 items checked and pass
- P1 issues are minor or documented
- No obvious bugs
- Test coverage adequate
- You're confident it won't break production
- Error handling is complete

**Don't approve if:**
- You don't understand the code
- You haven't checked P0 items
- You have concerns not addressed
- Tests don't cover the change

---

## Review Examples

### Good Feedback
```
file.go:42 - P0: Race condition
The `counter` field is accessed without a lock. This will race
when multiple goroutines call Inc() concurrently.

Fix: Add mutex protection:
  mu.Lock()
  defer mu.Unlock()
  s.counter++
```

### Bad Feedback
```
file.go:42 - This is wrong, fix it.
```

---

## Review Efficiency

### Focus Your Time
- **Most time**: P0 critical paths (concurrency, security, resources)
- **Medium time**: Business logic, error handling
- **Least time**: Style, comments (AI already checked)

### Skip These (AI Checked)
- Code formatting
- Import organization
- Simple naming issues
- Spelling in comments

### Don't Nitpick
- Personal style preferences
- Minor refactorings
- "I would've done it differently" (if current is correct)

---

## Final Checklist

Before approving:
- [ ] Understood the requirement
- [ ] Reviewed all changed files line-by-line
- [ ] Checked all P0 items
- [ ] No obvious bugs
- [ ] Error handling complete
- [ ] Tests cover the changes
- [ ] Performance acceptable
- [ ] Security verified
- [ ] Confident this won't break production

Before rejecting:
- [ ] Provided specific, actionable feedback
- [ ] Explained WHY issues are problems
- [ ] Suggested concrete fixes
- [ ] Marked severity (P0/P1/P2)

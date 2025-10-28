# Testing Guide for StratusShell UI

## Running Tests

### Unit Tests
All existing unit tests continue to pass:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Test specific package
go test ./internal/server/
go test ./internal/provision/
```

### Expected Output
```
✅ internal/provision: PASS
✅ internal/server: PASS  
✅ internal/service: PASS
✅ internal/validation: PASS
```

## Manual Testing

### 1. Build and Run

```bash
# Generate templ files
templ generate

# Build
go build -o stratusshell main.go

# Run server
./stratusshell serve --port=8080
```

### 2. Access UI

Navigate to: `http://localhost:8080/login?user=testuser`

### 3. Test Scenarios

#### Scenario 1: Terminal Creation
1. Click "Terminal → New Terminal" in navbar
2. Verify new tab appears with "Terminal 1"
3. Verify tab is active (highlighted)
4. Verify terminal iframe loads
5. Repeat up to 10 terminals

**Expected**: 
- Each terminal gets unique ID
- Tab becomes active automatically
- GoTTY loads in iframe

#### Scenario 2: Tab Switching
1. Create 3 terminals
2. Click on Tab 1
3. Verify Tab 1 becomes active
4. Click on Tab 2
5. Verify Tab 2 becomes active

**Expected**:
- Active tab has distinct styling
- Terminal iframe swaps correctly
- No page reload occurs

#### Scenario 3: Tab Renaming
1. Create terminal
2. Click on tab name ("Terminal 1")
3. Type new name ("Python Dev")
4. Click outside the input
5. Verify name persists

**Expected**:
- Input becomes editable
- Name saves on blur
- Database updated

#### Scenario 4: Terminal Closing
1. Create 3 terminals
2. Click × on Tab 2
3. Verify Tab 2 disappears
4. Verify remaining tabs still work

**Expected**:
- Terminal closes immediately
- GoTTY process killed
- Port released
- Database updated

#### Scenario 5: Max Capacity
1. Create 10 terminals
2. Verify + button becomes disabled
3. Hover over + button
4. Verify tooltip shows "Maximum terminals reached"

**Expected**:
- Cannot create 11th terminal
- Button visually disabled
- Tooltip provides feedback

#### Scenario 6: Save Session
1. Create 3 terminals with custom names
2. Click "Sessions → Save Session..."
3. Enter name: "Dev Setup"
4. Enter description: "Python and Node"
5. Click Save
6. Verify success message

**Expected**:
- Modal opens
- Form submits
- Success confirmation
- Database stores session

#### Scenario 7: Load Session
1. Save a session (from Scenario 6)
2. Close all terminals
3. Click "Sessions → Load Session..."
4. Select "Dev Setup"
5. Click Load
6. Verify terminals restore

**Expected**:
- Modal shows saved sessions
- Terminals recreate with saved names
- Configuration restores
- Active terminal set

#### Scenario 8: Empty State
1. Close all terminals
2. Verify empty state appears

**Expected**:
- Terminal icon displays
- "No terminals open" message
- "Create Terminal" button shows
- Clicking button creates terminal

### 4. Browser Compatibility

Test on multiple browsers:

#### Chrome/Edge
```bash
# Open in Chrome
google-chrome http://localhost:8080/login?user=testuser
```

#### Firefox
```bash
# Open in Firefox
firefox http://localhost:8080/login?user=testuser
```

#### Safari
```bash
# Open in Safari (macOS)
open -a Safari http://localhost:8080/login?user=testuser
```

### 5. Responsive Testing

Test at different viewport sizes:
- Desktop: 1920×1080
- Laptop: 1366×768
- Tablet: 768×1024

**Verify**:
- Layout doesn't break
- Tabs scroll horizontally if needed
- Buttons remain clickable
- Text remains readable

### 6. Performance Testing

#### Load Testing
```bash
# Create maximum terminals
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/terminals/add
  sleep 1
done
```

**Verify**:
- All 10 terminals spawn
- UI remains responsive
- Memory usage acceptable (~500MB total)

#### Stress Testing
- Rapidly create/close terminals
- Quickly switch between tabs
- Rename multiple tabs rapidly

**Verify**:
- No memory leaks
- No race conditions
- UI remains stable

### 7. Security Testing

#### CSRF Protection
```bash
# Attempt request without CSRF token
curl -X POST http://localhost:8080/api/terminals/add
```

**Expected**: Request blocked

#### Rate Limiting
```bash
# Send 150 requests rapidly
for i in {1..150}; do
  curl http://localhost:8080/ &
done
wait
```

**Expected**: Some requests return 429 Too Many Requests

#### Session Validation
```bash
# Access without session
curl http://localhost:8080/
```

**Expected**: Redirect to /login

### 8. Accessibility Testing

#### Keyboard Navigation
1. Tab through all interactive elements
2. Verify focus visible
3. Test Enter/Space on buttons
4. Test Escape on modals

**Expected**:
- All elements reachable via keyboard
- Focus indicators visible
- Actions work with keyboard

#### Screen Reader
1. Enable screen reader
2. Navigate through UI
3. Verify labels are read
4. Test form interactions

**Expected**:
- Semantic HTML provides context
- ARIA labels where needed
- Form fields properly labeled

## Debugging

### Check Logs
```bash
# Server logs
./stratusshell serve --port=8080

# System logs (if systemd)
sudo journalctl -u stratusshell-developer -f
```

### Browser DevTools

#### Console
- Check for JavaScript errors
- Verify HTMX requests
- Monitor WebSocket connections

#### Network Tab
- Verify API calls succeed
- Check response times
- Monitor WebSocket traffic

#### Elements Tab
- Inspect DOM structure
- Verify CSS classes applied
- Check computed styles

### Common Issues

#### Tabs Don't Load
**Check**:
- HTMX loaded? (Network tab)
- `/api/tabs` responding? (Network tab)
- JavaScript errors? (Console)

**Fix**:
- Clear browser cache
- Hard refresh (Ctrl+Shift+R)
- Check server logs

#### Styling Missing
**Check**:
- Tailwind CSS loaded? (Network tab)
- DaisyUI loaded? (Network tab)
- Custom CSS loaded? (Network tab)

**Fix**:
- Check CDN accessibility
- Verify static files serving
- Clear browser cache

#### Terminal Not Interactive
**Check**:
- WebSocket connected? (Network → WS tab)
- GoTTY process running? (`ps aux | grep gotty`)
- Port accessible? (`netstat -tulpn | grep :808`)

**Fix**:
- Check reverse proxy
- Verify GoTTY spawned
- Check credentials

## CI/CD Testing

### GitHub Actions
Tests run automatically on push:

```yaml
- name: Run tests
  run: go test ./...
```

### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit

templ generate
go test ./...
go build -o stratusshell main.go
```

## Test Coverage

### Current Coverage
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out
```

### Target Coverage
- Unit tests: 80%+
- Integration tests: 60%+
- Critical paths: 100%

## Performance Benchmarks

### Response Times
- API calls: < 100ms
- Terminal spawn: < 2s
- Tab switch: < 50ms
- Modal open: < 30ms

### Resource Usage
- Memory per terminal: ~50MB
- CPU at idle: < 1%
- CPU during spawn: < 10%
- Disk I/O: Minimal

## Continuous Testing

### Daily
- Smoke tests
- Basic functionality
- UI loads correctly

### Weekly
- Full test suite
- Performance tests
- Browser compatibility

### Before Release
- Complete manual testing
- Stress testing
- Security audit
- Accessibility audit

## Test Automation

Future improvements:
- [ ] E2E tests with Playwright
- [ ] Visual regression tests
- [ ] Load testing automation
- [ ] Performance monitoring
- [ ] Automated accessibility checks

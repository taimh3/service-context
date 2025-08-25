# K6 Load Testing for ScyllaDB Tasks API

This directory contains k6 load tests for the ScyllaDB tasks API endpoints.

## Files

- `test-get-task.js` - Main k6 test script for tasks API
- `run-tests.sh` - Bash script to run different test scenarios
- `README.md` - This file

## Prerequisites

1. **Install k6**:

   ```bash
   # macOS
   brew install k6

   # Ubuntu/Debian
   sudo apt update && sudo apt install k6

   # Or download from https://k6.io/docs/getting-started/installation/
   ```

2. **Start the API server**:
   ```bash
   cd /path/to/golang-clean-architecture
   make run
   # or
   go run main.go
   ```

## Running Tests

### Quick Start

```bash
# Make the runner script executable (if not already done)
chmod +x run-tests.sh

# Run all test scenarios
./run-tests.sh

# Run specific test scenario
./run-tests.sh smoke    # Quick smoke test
./run-tests.sh stress   # Stress test with 20 users
./run-tests.sh spike    # Spike test with varying load
```

### Direct k6 Commands

```bash
# Basic test run
k6 run test-get-task.js

# Run with custom base URL
k6 run --env BASE_URL=http://localhost:8080 test-get-task.js

# Smoke test (1 user for 30 seconds)
k6 run --vus 1 --duration 30s test-get-task.js

# Load test (10 users for 1 minute)
k6 run --vus 10 --duration 1m test-get-task.js

# Stress test (20 users for 2 minutes)
k6 run --vus 20 --duration 2m test-get-task.js
```

## Test Scenarios

The test script includes comprehensive testing of all tasks API endpoints:

### 1. Health Check (`/ping`)

- Verifies API is accessible
- Checks response time < 200ms
- Validates expected response format

### 2. Create Task (`POST /v1/scylla/tasks`)

- Creates new tasks with valid data
- Tests response time < 1000ms
- Validates success response format

### 3. Get Task (`GET /v1/scylla/tasks/{id}`)

- Retrieves specific tasks by ID
- Tests response time < 500ms
- Validates task data structure

### 4. Update Task (`PATCH /v1/scylla/tasks/{id}`)

- Updates existing tasks
- Tests partial updates
- Validates update responses

### 5. List Tasks (`GET /v1/scylla/tasks`)

- Lists all tasks
- Tests filtering by status
- Tests pagination parameters
- Validates response structure

### 6. Delete Task (`DELETE /v1/scylla/tasks/{id}`)

- Deletes tasks by ID
- Tests response time < 1000ms
- Validates deletion responses

### 7. Error Scenarios

- Tests invalid data handling
- Tests non-existent resource requests
- Validates error response formats

## Load Test Patterns

### Default Load Pattern

```
Ramp up:   5 users over 30s
Sustain:   10 users for 1m
Ramp down: 0 users over 30s
```

### Smoke Test

- 1 user for 30 seconds
- Quick verification that API works

### Stress Test

- 20 users for 2 minutes
- Tests performance under higher load

### Spike Test

- Sudden load increase to test scalability
- Pattern: 1→20→20→1 users

## Performance Thresholds

The tests include the following performance assertions:

- **Response Time**: 95% of requests < 500ms
- **Error Rate**: < 10% of requests should fail
- **HTTP Errors**: < 10% HTTP error rate
- **Health Check**: < 200ms response time

## Metrics

K6 provides detailed metrics including:

- **http_req_duration**: Request duration percentiles
- **http_req_failed**: Rate of failed HTTP requests
- **http_reqs**: Total number of HTTP requests
- **vus**: Number of active virtual users
- **iteration_duration**: Time for complete test iteration

## Customization

### Environment Variables

- `BASE_URL`: API base URL (default: http://localhost:8080)

### Test Data

Modify the test data in `test-get-task.js`:

```javascript
const TEST_TASK = {
  title: "Your Custom Task Title",
  description: "Your custom description",
  status: "doing",
};
```

### Load Patterns

Adjust the `options.stages` in `test-get-task.js` for custom load patterns:

```javascript
export const options = {
  stages: [
    { duration: "1m", target: 10 }, // Custom ramp up
    { duration: "5m", target: 10 }, // Custom sustain
    { duration: "1m", target: 0 }, // Custom ramp down
  ],
};
```

## Troubleshooting

### API Not Accessible

```
❌ API is not accessible at http://localhost:8080
```

- Ensure the API server is running
- Check if the port is correct
- Verify the base URL

### High Error Rates

- Check API server logs for errors
- Ensure database (ScyllaDB) is running
- Verify API endpoints are implemented correctly

### Performance Issues

- Monitor system resources during tests
- Check ScyllaDB performance metrics
- Consider database connection pool settings

## Integration with CI/CD

Example GitHub Actions workflow:

```yaml
name: Load Tests
on: [push, pull_request]

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Start API
        run: |
          # Start your API server
          make run &
          sleep 10

      - name: Install k6
        run: |
          sudo apt update
          sudo apt install k6

      - name: Run Load Tests
        run: |
          cd test/k6
          ./run-tests.sh smoke
```

## References

- [K6 Documentation](https://k6.io/docs/)
- [K6 JavaScript API](https://k6.io/docs/javascript-api/)
- [ScyllaDB Best Practices](https://docs.scylladb.com/stable/using-scylla/)

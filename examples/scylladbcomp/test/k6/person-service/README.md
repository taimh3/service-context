# K6 Load Testing for Person Service

This directory contains k6 load testing scripts for the ScyllaDB Person API endpoints.

## Overview

The Person Service provides the following endpoints:

- `POST /v1/scylla/persons` - Create a new person
- `GET /v1/scylla/persons` - List persons with optional filtering

## Files

- `test-person.js` - Main k6 test script with comprehensive test scenarios
- `run-tests.sh` - Shell script to run different test scenarios
- `validate-setup.sh` - Validation script to check test setup
- `config.json` - Configuration file with test scenarios and test data
- `README.md` - This documentation file

## Prerequisites

1. **Install k6**

   ```bash
   # macOS
   brew install k6

   # Ubuntu/Debian
   sudo apt update && sudo apt install k6

   # Or download from https://k6.io/docs/getting-started/installation/
   ```

2. **Start the API Server**

   ```bash
   cd ../../
   go run main.go
   # or
   make run
   ```

3. **Verify Setup**
   ```bash
   ./validate-setup.sh
   ```

## Quick Start

### 1. Run All Test Scenarios

```bash
./run-tests.sh
```

### 2. Run Specific Test Scenarios

```bash
# Quick smoke test (1 user, 30 seconds)
./run-tests.sh smoke

# Stress test (20 users, 2 minutes)
./run-tests.sh stress

# Spike test (varying load)
./run-tests.sh spike
```

### 3. Run with Custom Settings

```bash
# Direct k6 execution
k6 run test-person.js

# Custom virtual users and duration
k6 run --vus 5 --duration 1m test-person.js

# With custom base URL
k6 run --env BASE_URL=http://api.example.com test-person.js
```

## Test Scenarios

### Default Load Test

- **Duration**: 2 minutes
- **Pattern**: Ramp up to 5 users (30s) → 10 users (1m) → Ramp down (30s)
- **Tests**: All CRUD operations with various data combinations

### Smoke Test

- **Duration**: 30 seconds
- **Users**: 1
- **Purpose**: Quick validation of API functionality

### Stress Test

- **Duration**: 2 minutes
- **Users**: 20
- **Purpose**: Test API under high load

### Spike Test

- **Duration**: 1 minute
- **Pattern**: 1 user → 20 users → 20 users → 1 user
- **Purpose**: Test API response to sudden load changes

## Test Coverage

### API Endpoints

- ✅ Health check (`/ping`)
- ✅ Create person (`POST /v1/scylla/persons`)
- ✅ List all persons (`GET /v1/scylla/persons`)
- ✅ List persons by first name (`GET /v1/scylla/persons?first_name=John`)
- ✅ List persons by full name (`GET /v1/scylla/persons?first_name=John&last_name=Doe`)

### Test Cases

- ✅ Valid person creation
- ✅ Person listing without filters
- ✅ Person listing with first_name filter (partition key)
- ✅ Person listing with both first_name and last_name filters
- ✅ Error handling for invalid data
- ✅ Error handling for missing required fields
- ✅ Error handling for malformed JSON
- ✅ Performance under various load conditions

### Data Validation

- ✅ Required field validation (first_name, last_name)
- ✅ Email format handling
- ✅ Empty and null value handling
- ✅ Long string handling

## Performance Thresholds

### Default Thresholds

- 95th percentile response time: < 500ms
- Error rate: < 10%
- HTTP failure rate: < 10%

### Stress Test Thresholds

- 95th percentile response time: < 1000ms
- Error rate: < 15%

### Spike Test Thresholds

- 95th percentile response time: < 2000ms
- Error rate: < 20%

## Environment Variables

- `BASE_URL` - API base URL (default: `http://localhost:8080`)

## Example Output

```bash
$ ./run-tests.sh smoke

🚀 Starting K6 Load Tests for Persons API
Base URL: http://localhost:8080
✅ k6 found: k6 v0.47.0

🔍 Checking if API is accessible...
✅ API is accessible at http://localhost:8080

📊 Running: Smoke Test
----------------------------------------

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: test-person.js
     output: -

  scenarios: (100.00%) 1 scenario, 1 max VUs, 1m0s max duration (incl. graceful stop):
           * default: 1 looping VUs for 30s (gracefulStop: 30s)

✅ Smoke Test completed successfully
```

## Troubleshooting

### Common Issues

1. **API not accessible**

   ```
   ❌ API is not accessible at http://localhost:8080
   ```

   **Solution**: Make sure your API server is running on the correct port.

2. **k6 not found**

   ```
   ❌ k6 is not installed
   ```

   **Solution**: Install k6 following the prerequisites section.

3. **Test file syntax errors**
   ```
   ❌ Test file has syntax errors
   ```
   **Solution**: Run `k6 inspect test-person.js` to see detailed syntax errors.

### Debugging

1. **Verbose output**

   ```bash
   k6 run --verbose test-person.js
   ```

2. **HTTP debugging**

   ```bash
   k6 run --http-debug test-person.js
   ```

3. **Custom log output**
   ```bash
   k6 run --console-output=person-test.log test-person.js
   ```

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Run Person Service Load Tests
  run: |
    cd test/k6/person-service
    ./run-tests.sh smoke
```

### Docker Example

```bash
docker run --rm -v $(pwd):/scripts grafana/k6 run /scripts/test-person.js
```

## Contributing

When adding new test scenarios:

1. Follow the existing test structure
2. Add appropriate checks and thresholds
3. Update this README with new test coverage
4. Test both success and failure scenarios
5. Consider ScyllaDB-specific characteristics (partition keys, clustering keys)

## Related Documentation

- [K6 Documentation](https://k6.io/docs/)
- [ScyllaDB Documentation](https://docs.scylladb.com/)
- [Person API OpenAPI Specification](../../../api/openapi-spec/golang-clean-architecture.yaml)

# Film API Testing Guide

This document provides information about the test suite for the Film API project.

## Test Structure

The test suite is organized to match the structure of the application:

1. **Model Tests**: Tests for the data models and database operations
   - `internals/models/*.go`

2. **Handler Tests**: Tests for the HTTP handlers
   - `cmd/api/handlers_test.go`

3. **Middleware Tests**: Tests for the middleware functions
   - `cmd/api/middleware_test.go`

4. **Helper Tests**: Tests for the helper functions
   - `cmd/api/helpers_test.go`

5. **Validator Tests**: Tests for the validation functions
   - `internals/validator/validator_test.go`

6. **Logger Tests**: Tests for the JSON logger
   - `internals/jsonlog/jsonlog_test.go`

## Running Tests

You can run the tests using the provided script:

```bash
./run_tests.sh
```

This script will:
1. Run all tests with verbose output
2. Run tests with coverage information
3. Generate an HTML coverage report

Alternatively, you can run tests manually:

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internals/models

# Run tests with verbose output
go test -v ./...

# Run tests with coverage information
go test -cover ./...
```

## Test Coverage

The test suite aims to provide comprehensive coverage of the application code. The coverage report generated by `run_tests.sh` will show which parts of the code are covered by tests and which are not.

## Mock Objects

The tests use mock objects to simulate database operations and other external dependencies. This allows the tests to run quickly and reliably without requiring a real database or external services.

The main mock objects are:
- `MockDB`: A mock database connection
- `MockFilmModel`: A mock film model
- `MockUserModel`: A mock user model
- `MockTokenModel`: A mock token model
- `MockPermissionModel`: A mock permission model

## Test Data

The tests use a variety of test data to cover different scenarios:
- Valid and invalid films
- Valid and invalid users
- Valid and invalid tokens
- Various HTTP requests and responses


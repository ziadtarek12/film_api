#!/bin/bash

# Run all tests with verbose output
go test -v ./...

# Run tests with coverage report
echo -e "\nRunning tests with coverage report..."
go test -cover ./...

# Generate HTML coverage report
echo -e "\nGenerating detailed coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
echo "Coverage report generated at coverage.html"

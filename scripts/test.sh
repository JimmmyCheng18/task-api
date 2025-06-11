#!/bin/bash

# Test script for Task API
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_DIR="./coverage"
COVERAGE_FILE="${COVERAGE_DIR}/coverage.out"
COVERAGE_HTML="${COVERAGE_DIR}/coverage.html"

echo -e "${GREEN}🧪 Running Task API Tests${NC}"

# Create coverage directory
mkdir -p ${COVERAGE_DIR}

# Function to run tests with coverage
run_tests_with_coverage() {
    echo -e "${BLUE}📊 Running tests with coverage...${NC}"
    go test -v -race -coverprofile=${COVERAGE_FILE} -covermode=atomic ./...
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
    else
        echo -e "${RED}❌ Some tests failed!${NC}"
        exit 1
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    echo -e "${BLUE}📈 Generating coverage report...${NC}"
    go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_HTML}
    
    # Display coverage percentage
    COVERAGE_PERCENT=$(go tool cover -func=${COVERAGE_FILE} | grep total | awk '{print $3}')
    echo -e "${YELLOW}📊 Total coverage: ${COVERAGE_PERCENT}${NC}"
    
    # Check if coverage meets minimum threshold
    THRESHOLD=${MIN_COVERAGE:-70}
    COVERAGE_NUM=$(echo ${COVERAGE_PERCENT} | sed 's/%//')
    
    if (( $(echo "${COVERAGE_NUM} >= ${THRESHOLD}" | bc -l) )); then
        echo -e "${GREEN}✅ Coverage meets minimum threshold of ${THRESHOLD}%${NC}"
    else
        echo -e "${RED}❌ Coverage ${COVERAGE_PERCENT} is below minimum threshold of ${THRESHOLD}%${NC}"
        if [ "${STRICT_COVERAGE}" = "true" ]; then
            exit 1
        fi
    fi
}

# Function to run specific package tests
run_package_tests() {
    local package=$1
    echo -e "${BLUE}🔍 Running tests for package: ${package}${NC}"
    go test -v -race ${package}
}

# Function to run benchmark tests
run_benchmarks() {
    echo -e "${BLUE}⚡ Running benchmark tests...${NC}"
    go test -bench=. -benchmem ./...
}

# Function to run integration tests
run_integration_tests() {
    echo -e "${BLUE}🔗 Running integration tests...${NC}"
    if [ -d "./test/integration" ]; then
        go test -v -race -tags=integration ./test/integration/...
    else
        echo -e "${YELLOW}⚠️  No integration tests found${NC}"
    fi
}

# Function to run e2e tests
run_e2e_tests() {
    echo -e "${BLUE}🌐 Running end-to-end tests...${NC}"
    if [ -d "./test/e2e" ]; then
        go test -v -race -tags=e2e ./test/e2e/...
    else
        echo -e "${YELLOW}⚠️  No e2e tests found${NC}"
    fi
}

# Function to check test dependencies
check_dependencies() {
    echo -e "${BLUE}🔍 Checking test dependencies...${NC}"
    
    # Check if required tools are installed
    if ! command -v bc &> /dev/null; then
        echo -e "${YELLOW}⚠️  bc calculator not found, coverage threshold check may not work${NC}"
    fi
    
    # Check if Go modules are up to date
    go mod verify
    go mod tidy
}

# Function to clean test artifacts
clean_test_artifacts() {
    echo -e "${BLUE}🧹 Cleaning test artifacts...${NC}"
    rm -rf ${COVERAGE_DIR}
    go clean -testcache
}

# Function to lint code
run_linting() {
    echo -e "${BLUE}🔍 Running code linting...${NC}"
    
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
    else
        echo -e "${YELLOW}⚠️  golangci-lint not found, skipping linting${NC}"
        echo -e "${YELLOW}    Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
    fi
}

# Function to check code formatting
check_formatting() {
    echo -e "${BLUE}📝 Checking code formatting...${NC}"
    
    UNFORMATTED=$(gofmt -l .)
    if [ -n "${UNFORMATTED}" ]; then
        echo -e "${RED}❌ The following files are not properly formatted:${NC}"
        echo "${UNFORMATTED}"
        echo -e "${YELLOW}Run 'gofmt -w .' to fix formatting${NC}"
        if [ "${STRICT_FORMAT}" = "true" ]; then
            exit 1
        fi
    else
        echo -e "${GREEN}✅ All files are properly formatted${NC}"
    fi
}

# Function to run security checks
run_security_checks() {
    echo -e "${BLUE}🔒 Running security checks...${NC}"
    
    if command -v gosec &> /dev/null; then
        gosec ./...
    else
        echo -e "${YELLOW}⚠️  gosec not found, skipping security checks${NC}"
        echo -e "${YELLOW}    Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest${NC}"
    fi
}

# Main execution based on arguments
case "$1" in
    "clean")
        clean_test_artifacts
        ;;
    "deps")
        check_dependencies
        ;;
    "unit")
        run_tests_with_coverage
        generate_coverage_report
        ;;
    "integration")
        run_integration_tests
        ;;
    "e2e")
        run_e2e_tests
        ;;
    "benchmark")
        run_benchmarks
        ;;
    "lint")
        run_linting
        ;;
    "format")
        check_formatting
        ;;
    "security")
        run_security_checks
        ;;
    "package")
        if [ -z "$2" ]; then
            echo -e "${RED}❌ Package name required${NC}"
            echo "Usage: $0 package <package_name>"
            exit 1
        fi
        run_package_tests $2
        ;;
    "all")
        check_dependencies
        check_formatting
        run_linting
        run_tests_with_coverage
        generate_coverage_report
        run_integration_tests
        run_e2e_tests
        run_security_checks
        ;;
    *)
        echo -e "${GREEN}🧪 Task API Test Runner${NC}"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  unit        Run unit tests with coverage"
        echo "  integration Run integration tests"
        echo "  e2e         Run end-to-end tests"
        echo "  benchmark   Run benchmark tests"
        echo "  lint        Run code linting"
        echo "  format      Check code formatting"
        echo "  security    Run security checks"
        echo "  package     Run tests for specific package"
        echo "  deps        Check test dependencies"
        echo "  clean       Clean test artifacts"
        echo "  all         Run all tests and checks"
        echo ""
        echo "Environment variables:"
        echo "  MIN_COVERAGE    Minimum coverage percentage (default: 70)"
        echo "  STRICT_COVERAGE Exit on coverage failure (default: false)"
        echo "  STRICT_FORMAT   Exit on format issues (default: false)"
        echo ""
        
        # Run default test suite
        run_tests_with_coverage
        generate_coverage_report
        ;;
esac

echo -e "${GREEN}🎉 Test execution completed!${NC}"
#!/bin/bash

# Build script for Task API
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="task-api"
BUILD_DIR="./bin"
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${GREEN}🚀 Building Task API${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Build Time: ${BUILD_TIME}${NC}"
echo -e "${YELLOW}Git Commit: ${GIT_COMMIT}${NC}"

# Create build directory
mkdir -p ${BUILD_DIR}

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Build for current platform
echo -e "${GREEN}📦 Building for current platform...${NC}"
go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME} ./cmd/server

# Build for multiple platforms if MULTI_PLATFORM is set
if [ "${MULTI_PLATFORM}" = "true" ]; then
    echo -e "${GREEN}📦 Building for multiple platforms...${NC}"
    
    # Linux AMD64
    echo -e "${YELLOW}Building for Linux AMD64...${NC}"
    GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-linux-amd64 ./cmd/server
    
    # Linux ARM64
    echo -e "${YELLOW}Building for Linux ARM64...${NC}"
    GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-linux-arm64 ./cmd/server
    
    # Darwin AMD64 (Intel Mac)
    echo -e "${YELLOW}Building for Darwin AMD64...${NC}"
    GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-darwin-amd64 ./cmd/server
    
    # Darwin ARM64 (Apple Silicon)
    echo -e "${YELLOW}Building for Darwin ARM64...${NC}"
    GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-darwin-arm64 ./cmd/server
    
    # Windows AMD64
    echo -e "${YELLOW}Building for Windows AMD64...${NC}"
    GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-windows-amd64.exe ./cmd/server
fi

# Make binaries executable
chmod +x ${BUILD_DIR}/${APP_NAME}*

# Display build results
echo -e "${GREEN}✅ Build completed successfully!${NC}"
echo -e "${YELLOW}Built binaries:${NC}"
ls -la ${BUILD_DIR}

# Optional: Create checksums
if [ "${CREATE_CHECKSUMS}" = "true" ]; then
    echo -e "${GREEN}📝 Creating checksums...${NC}"
    cd ${BUILD_DIR}
    sha256sum ${APP_NAME}* > checksums.txt
    echo -e "${YELLOW}Checksums created in ${BUILD_DIR}/checksums.txt${NC}"
    cd ..
fi

echo -e "${GREEN}🎉 Build process completed!${NC}"
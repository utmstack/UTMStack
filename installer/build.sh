#!/bin/bash

set -e  # Exit on error

# ============================================
# CONFIGURATION - Edit these values
# ============================================
DEFAULT_BRANCH="prod"
INSTALLER_VERSION="v11.0.0-dev.1"
CM_ENCRYPT_SALT="your-encryption-salt-here"
CM_SIGN_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
your-public-key-here
-----END PUBLIC KEY-----"

# ============================================
# Build Process
# ============================================

# Change to installer directory
cd "$(dirname "${BASH_SOURCE[0]}")"

# Set Go environment variables (same as GitHub Actions)
export GOOS=linux
export GOARCH=amd64
export GOPRIVATE=github.com/utmstack
export GONOPROXY=github.com/utmstack
export GONOSUMDB=github.com/utmstack

echo "Building V11 Installer for production release"

# Execute build with ldflags
go build -o installer -v -ldflags "\
-X 'github.com/utmstack/UTMStack/installer/config.DEFAULT_BRANCH=${DEFAULT_BRANCH}' \
-X 'github.com/utmstack/UTMStack/installer/config.INSTALLER_VERSION=${INSTALLER_VERSION}' \
-X 'github.com/utmstack/UTMStack/installer/config.REPLACE=${CM_ENCRYPT_SALT}' \
-X 'github.com/utmstack/UTMStack/installer/config.PUBLIC_KEY=${CM_SIGN_PUBLIC_KEY}'" \
.

echo "✅ Build completed: ./installer"

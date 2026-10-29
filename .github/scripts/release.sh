#!/bin/bash

# Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
#
# This software is the property of WSO2 LLC. and its suppliers, if any.
# Dissemination of any information or reproduction of any material contained
# herein in any form is strictly forbidden, unless permitted by WSO2 expressly.
# You may not alter or remove any copyright or other notice from copies of this content.
#

# Exit the script on any command with non-zero exit status.
set -e
set -o pipefail

UPSTREAM_BRANCH="main"

# Assign command line arguments to variables.
GIT_TOKEN=$1
WORK_DIR=$2
VERSION_TYPE=$3 # possible values: major, minor, patch
SERVER=$4 # possible values: asgardeo-mcp, self-service-mcp

# Check if GIT_TOKEN is empty
if [ -z "$GIT_TOKEN" ]; then
  echo "❌ Error: GIT_TOKEN is not set."
  exit 1
fi

# Check if WORK_DIR is empty
if [ -z "$WORK_DIR" ]; then
  echo "❌ Error: WORK_DIR is not set."
  exit 1
fi

# Validate VERSION_TYPE
if [[ "$VERSION_TYPE" != "major" && "$VERSION_TYPE" != "minor" && "$VERSION_TYPE" != "patch" ]]; then
  echo "❌ Error: VERSION_TYPE must be one of: major, minor, or patch."
  exit 1
fi

# Each server is released with its own tags and packages.
case "$SERVER" in
  asgardeo-mcp)
    TAG_PREFIX="asgardeo-mcp-v"
    # Releases before the self-service server were tagged vX.Y.Z.
    LEGACY_TAG_PREFIX="v"
    PACKAGE_PREFIX="asgardeo_mcp_server"
    SERVER_NAME="Asgardeo MCP Server"
    MARK_LATEST="true"
    ;;
  self-service-mcp)
    TAG_PREFIX="self-service-mcp-v"
    PACKAGE_PREFIX="self_service_mcp_server"
    SERVER_NAME="Self Service MCP Server"
    MARK_LATEST="false"
    ;;
  *)
    echo "❌ Error: SERVER must be one of: asgardeo-mcp or self-service-mcp."
    exit 1
    ;;
esac

BUILD_DIRECTORY="$WORK_DIR/build"
RELEASE_DIRECTORY="$BUILD_DIRECTORY/releases"

# Navigate to the working directory.
cd "${WORK_DIR}"

# Create the release directory.
if [ ! -d "$RELEASE_DIRECTORY" ]; then
  mkdir -p "$RELEASE_DIRECTORY"
else
  rm -rf "$RELEASE_DIRECTORY"/*
fi

# Extract the current version of this server only.
CURRENT_VERSION=$(
  for prefix in "${TAG_PREFIX}" ${LEGACY_TAG_PREFIX:+"$LEGACY_TAG_PREFIX"}; do
    git tag -l "${prefix}*" | grep -E "^${prefix}[0-9]+\.[0-9]+\.[0-9]+$" | sed "s/^${prefix}//" || true
  done | sort -t. -k1,1n -k2,2n -k3,3n | tail -n 1
)
CURRENT_VERSION=${CURRENT_VERSION:-0.0.0}
echo "Current Version: ${CURRENT_VERSION}"
IFS='.' read -r MAJOR MINOR PATCH <<< "${CURRENT_VERSION}"

# Determine which part to increment
case "$VERSION_TYPE" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch|*)
    PATCH=$((PATCH + 1))
    ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "Building ${SERVER} ${NEW_VERSION}..."
make build-linux build-linux-arm build-darwin SERVER="${SERVER}" BUILD_VERSION="${NEW_VERSION}"

echo "Creating release packages for version $NEW_VERSION..."

# List of supported OSes.
oses=("linux" "linux-arm" "darwin")

# Navigate to the release directory.
cd "${RELEASE_DIRECTORY}"

for os in "${oses[@]}"; do
  os_dir="../$os"
  
  if [ -d "$os_dir" ]; then
    release_artifact_folder="${PACKAGE_PREFIX}_${os}-v${NEW_VERSION}"
    mkdir -p "$release_artifact_folder"

    cp -r $os_dir/* "$release_artifact_folder"

    # Zip the release package.
    zip_file="$release_artifact_folder.zip"
    echo "Creating $zip_file..."
    zip -r "$zip_file" "$release_artifact_folder"

    # Delete the folder after zipping.
    rm -rf "$release_artifact_folder"

    # Generate checksum file.
    sha256sum "$zip_file" | sed "s|target/releases/||" > "$zip_file.sha256"
    echo "Checksum generated for the $os package."

    echo "Release packages created successfully for $os."
  else
    echo "Skipping $os release package creation as the build artifacts are not available."
  fi
done

echo "Release packages created successfully in $RELEASE_DIRECTORY."

# Navigate back to the project root directory.
cd "${WORK_DIR}"

# Collect all ZIP and .sha256 files in the target/releases directory.
FILES_TO_UPLOAD=$(find build/releases -type f \( -name "*.zip" -o -name "*.sha256" \))

# Create a release with the current version.
TAG_NAME="${TAG_PREFIX}${NEW_VERSION}"
export GITHUB_TOKEN="${GIT_TOKEN}"
gh release create "${TAG_NAME}" ${FILES_TO_UPLOAD} --title "${TAG_NAME}" --notes "${SERVER_NAME} - ${TAG_NAME}" --target "${UPSTREAM_BRANCH}" --latest="${MARK_LATEST}" || { echo "Failed to create release"; exit 1; }


echo "Release ${TAG_NAME} created successfully."

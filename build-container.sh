#!/bin/bash
# build-container.sh - Podman build script for OpenShift MCP

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
CONTAINER_NAME="openshift-mcp"
REGISTRY="quay.io"
NAMESPACE="openshift-community"
TAG="latest"
PLATFORM="linux/amd64"
BUILD_CONTEXT="."
CONTAINERFILE="Containerfile"

# Build information
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}
DATE=${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}
BUILD_USER=${BUILD_USER:-$(whoami)}

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"; }
info() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"; }

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build OpenShift MCP container image using Podman

OPTIONS:
    -n, --name NAME         Container name (default: $CONTAINER_NAME)
    -r, --registry REGISTRY Registry URL (default: $REGISTRY)
    -ns, --namespace NS     Registry namespace (default: $NAMESPACE)
    -t, --tag TAG          Image tag (default: $TAG)
    -p, --platform ARCH    Target platform (default: $PLATFORM)
    -f, --file FILE        Containerfile path (default: $CONTAINERFILE)
    -c, --context DIR      Build context (default: $BUILD_CONTEXT)
    --no-cache            Don't use build cache
    --push                Push image after build
    --scan                Scan image for vulnerabilities
    -h, --help            Show this help

EXAMPLES:
    $0                                          # Basic build
    $0 --tag v1.0.0 --push                    # Build and push with tag
    $0 --registry docker.io --namespace myorg  # Custom registry
    $0 --platform linux/arm64                 # ARM64 build
    $0 --scan                                  # Build with security scan

ENVIRONMENT VARIABLES:
    VERSION        Image version (default: git describe)
    COMMIT         Git commit hash (default: git rev-parse HEAD)
    DATE           Build date (default: current UTC)
    BUILD_USER     Build user (default: whoami)
    REGISTRY_USER  Registry username for push
    REGISTRY_PASS  Registry password for push

EOF
}

# Parse command line arguments
NO_CACHE=""
PUSH=false
SCAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            CONTAINER_NAME="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -ns|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -f|--file)
            CONTAINERFILE="$2"
            shift 2
            ;;
        -c|--context)
            BUILD_CONTEXT="$2"
            shift 2
            ;;
        --no-cache)
            NO_CACHE="--no-cache"
            shift
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --scan)
            SCAN=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate requirements
if ! command -v podman &> /dev/null; then
    error "Podman is not installed or not in PATH"
    exit 1
fi

if ! command -v git &> /dev/null; then
    warn "Git is not available, using default version info"
fi

# Build image tags
FULL_IMAGE_NAME="${REGISTRY}/${NAMESPACE}/${CONTAINER_NAME}"
VERSION_TAG="${FULL_IMAGE_NAME}:${TAG}"
LATEST_TAG="${FULL_IMAGE_NAME}:latest"

# Display build information
info "Build Configuration:"
echo "  Container Name: ${CONTAINER_NAME}"
echo "  Registry: ${REGISTRY}"
echo "  Namespace: ${NAMESPACE}"
echo "  Full Image: ${FULL_IMAGE_NAME}"
echo "  Version Tag: ${VERSION_TAG}"
echo "  Platform: ${PLATFORM}"
echo "  Containerfile: ${CONTAINERFILE}"
echo "  Build Context: ${BUILD_CONTEXT}"
echo "  Version: ${VERSION}"
echo "  Commit: ${COMMIT}"
echo "  Date: ${DATE}"
echo "  Build User: ${BUILD_USER}"
echo ""

# Check if Containerfile exists
if [[ ! -f "${CONTAINERFILE}" ]]; then
    error "Containerfile not found: ${CONTAINERFILE}"
    exit 1
fi

# Build the image
log "Building container image..."
BUILD_CMD="podman build \
    --platform ${PLATFORM} \
    --file ${CONTAINERFILE} \
    --build-arg VERSION=${VERSION} \
    --build-arg COMMIT=${COMMIT} \
    --build-arg DATE=${DATE} \
    --build-arg BUILD_USER=${BUILD_USER} \
    --tag ${VERSION_TAG} \
    --tag ${LATEST_TAG} \
    ${NO_CACHE} \
    ${BUILD_CONTEXT}"

info "Executing: ${BUILD_CMD}"
eval "${BUILD_CMD}"

if [[ $? -eq 0 ]]; then
    log "Container image built successfully!"
else
    error "Container build failed!"
    exit 1
fi

# List the built images
info "Built images:"
podman images | grep "${CONTAINER_NAME}" | head -5

# Get image size
IMAGE_SIZE=$(podman inspect "${VERSION_TAG}" --format "{{.Size}}" 2>/dev/null | numfmt --to=iec || echo "unknown")
info "Image size: ${IMAGE_SIZE}"

# Security scan if requested
if [[ "${SCAN}" == "true" ]]; then
    log "Scanning image for vulnerabilities..."
    if command -v podman &> /dev/null; then
        # Use Podman's built-in scanner if available
        podman scan "${VERSION_TAG}" || warn "Vulnerability scan completed with findings"
    elif command -v skopeo &> /dev/null && command -v trivy &> /dev/null; then
        # Use Trivy if available
        trivy image "${VERSION_TAG}" || warn "Trivy scan completed with findings"
    else
        warn "No vulnerability scanner available (install trivy for security scanning)"
    fi
fi

# Push if requested
if [[ "${PUSH}" == "true" ]]; then
    log "Pushing images to registry..."
    
    # Login if credentials are provided
    if [[ -n "${REGISTRY_USER:-}" ]] && [[ -n "${REGISTRY_PASS:-}" ]]; then
        echo "${REGISTRY_PASS}" | podman login --username "${REGISTRY_USER}" --password-stdin "${REGISTRY}"
    fi
    
    # Push version tag
    podman push "${VERSION_TAG}"
    
    # Push latest tag only if it's not the same as version tag
    if [[ "${TAG}" != "latest" ]]; then
        podman push "${LATEST_TAG}"
    fi
    
    log "Images pushed successfully!"
fi

# Final information
log "Build completed successfully!"
info "To run the container:"
echo "  podman run --rm -p 8080:8080 -e GEMINI_API_KEY=your_key ${VERSION_TAG}"
echo ""
info "To push manually:"
echo "  podman push ${VERSION_TAG}"
echo "  podman push ${LATEST_TAG}"
echo ""
info "To run with OpenShift:"
echo "  oc new-app ${VERSION_TAG}"

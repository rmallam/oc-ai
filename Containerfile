# Containerfile for Podman/Buildah
# This is optimized for Red Hat UBI and OpenShift environments

# Build stage using Red Hat UBI Go toolset
FROM registry.access.redhat.com/ubi9/go-toolset:1.21 AS builder

# Switch to root to install dependencies
USER 0

# Set working directory
WORKDIR /opt/app-root/src

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY --chown=1001:0 . .

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
ARG BUILD_USER=podman

# Build the application with optimization flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static' \
             -X main.version=${VERSION} \
             -X main.commit=${COMMIT} \
             -X main.date=${DATE} \
             -X main.buildUser=${BUILD_USER}" \
    -o openshift-mcp ./cmd/openshift-mcp

# Verify the binary
RUN file openshift-mcp && \
    ldd openshift-mcp 2>&1 | grep -q "not a dynamic executable" || echo "Warning: Binary has dynamic dependencies"

# Final stage using minimal UBI
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

# Add metadata labels (OpenShift/Kubernetes standard)
LABEL name="openshift-mcp" \
      vendor="OpenShift Community" \
      version="${VERSION}" \
      release="${COMMIT}" \
      summary="AI-powered OpenShift SRE Assistant" \
      description="OpenShift MCP provides intelligent cluster management, diagnostics, and automation through a conversational AI interface using advanced LLM technology." \
      maintainer="OpenShift MCP Team" \
      io.k8s.description="OpenShift MCP - AI SRE Assistant for cluster management and diagnostics" \
      io.k8s.display-name="OpenShift MCP" \
      io.openshift.tags="openshift,ai,sre,assistant,go,llm,kubernetes" \
      io.openshift.wants="openshift" \
      com.redhat.component="openshift-mcp" \
      usage="podman run -p 8080:8080 -e GEMINI_API_KEY=your_key openshift-mcp"

# Install minimal runtime dependencies
RUN microdnf update -y && \
    microdnf install -y \
        ca-certificates \
        tzdata \
        curl \
        shadow-utils && \
    microdnf clean all && \
    rm -rf /var/cache/yum

# Create application user with OpenShift-compatible UID/GID
RUN useradd -r -u 1001 -g 0 \
    -d /home/openshift-mcp \
    -s /sbin/nologin \
    -c "OpenShift MCP Application User" \
    openshift-mcp

# Create necessary directories with proper permissions
RUN mkdir -p /home/openshift-mcp/.config/openshift-mcp && \
    mkdir -p /var/log/openshift-mcp && \
    mkdir -p /tmp/openshift-mcp && \
    chown -R 1001:0 /home/openshift-mcp && \
    chown -R 1001:0 /var/log/openshift-mcp && \
    chown -R 1001:0 /tmp/openshift-mcp && \
    chmod -R g=u /home/openshift-mcp && \
    chmod -R g=u /var/log/openshift-mcp && \
    chmod -R g=u /tmp/openshift-mcp

# Set working directory
WORKDIR /home/openshift-mcp

# Copy binary from builder stage
COPY --from=builder --chown=1001:0 /opt/app-root/src/openshift-mcp ./

# Ensure binary is executable
RUN chmod +x openshift-mcp

# Copy web templates if they exist
COPY --from=builder --chown=1001:0 /opt/app-root/src/web ./web

# Create a non-root user script for better security
RUN echo '#!/bin/bash\nset -e\nexec "$@"' > /usr/local/bin/entrypoint.sh && \
    chmod +x /usr/local/bin/entrypoint.sh && \
    chown 1001:0 /usr/local/bin/entrypoint.sh

# Switch to non-root user (OpenShift security requirement)
USER 1001

# Expose application port
EXPOSE 8080

# Add health check using curl
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables for configuration
ENV OPENSHIFT_MCP_HOST=0.0.0.0 \
    OPENSHIFT_MCP_PORT=8080 \
    OPENSHIFT_MCP_DATABASE_PATH=/home/openshift-mcp/.config/openshift-mcp/memory.db \
    HOME=/home/openshift-mcp

# Use entrypoint script for better signal handling
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Default command
CMD ["./openshift-mcp", "--host", "0.0.0.0", "--port", "8080"]

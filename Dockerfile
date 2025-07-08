# Build stage
FROM registry.access.redhat.com/ubi9/go-toolset:1.21 AS builder

# Set working directory
USER 0
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o openshift-mcp ./cmd/openshift-mcp

# Final stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

# Install runtime dependencies
RUN microdnf update -y && \
    microdnf install -y ca-certificates tzdata curl && \
    microdnf clean all

# Create non-root user
RUN useradd -r -u 1001 -g 0 -d /home/openshift-mcp -s /bin/bash \
    -c "OpenShift MCP User" openshift-mcp && \
    mkdir -p /home/openshift-mcp && \
    chown -R 1001:0 /home/openshift-mcp && \
    chmod -R g=u /home/openshift-mcp

# Set working directory
WORKDIR /home/openshift-mcp

# Copy binary from builder stage
COPY --from=builder /app/openshift-mcp .

# Create config directory
RUN mkdir -p .config/openshift-mcp && \
    chown -R 1001:0 .config && \
    chmod -R g=u .config

# Switch to non-root user
USER 1001

# Expose port
EXPOSE 8080

# Health check using curl (UBI includes curl)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Add labels for better container metadata
LABEL name="openshift-mcp" \
      vendor="OpenShift Community" \
      version="${VERSION}" \
      release="${COMMIT}" \
      summary="AI-powered OpenShift SRE Assistant" \
      description="OpenShift MCP provides intelligent cluster management, diagnostics, and automation through a conversational AI interface." \
      io.k8s.description="OpenShift MCP - AI SRE Assistant" \
      io.k8s.display-name="OpenShift MCP" \
      io.openshift.tags="openshift,ai,sre,assistant,go"

# Set default command
CMD ["./openshift-mcp", "--host", "0.0.0.0", "--port", "8080"]

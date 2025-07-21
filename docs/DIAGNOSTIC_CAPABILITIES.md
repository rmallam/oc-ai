# Enhanced Diagnostic and Analysis Capabilities

## Overview

The OpenShift MCP server has been enhanced with comprehensive diagnostic collection and analysis capabilities that mirror the workflows of experienced support engineers. These tools enable automated collection of system diagnostics, network captures, logs, and intelligent analysis of the collected data.

## Diagnostic Collection Tools

### 1. SOS Report Collection (`collect_sosreport`)

Collects comprehensive system reports from OpenShift nodes using sosreport, similar to what support engineers do for system-level troubleshooting.

**Parameters:**
- `node_name` (required): Target node name
- `output_dir` (optional): Custom output directory

**Example:**
```json
{
  "tool": "collect_sosreport",
  "parameters": {
    "node_name": "worker-node-01",
    "output_dir": "/tmp/diagnostics/sosreports"
  }
}
```

**What it does:**
- Creates a privileged debug pod on the target node
- Runs sosreport within the node's filesystem
- Collects system configuration, logs, and diagnostic data
- Packages everything into a compressed archive
- Provides size, duration, and location information

### 2. Network Packet Capture (`collect_tcpdump`)

Performs network packet capture for troubleshooting connectivity issues, similar to network troubleshooting workflows.

**Parameters:**
- `pod_name` (optional): Target pod for pod-level capture
- `node_name` (optional): Target node for node-level capture
- `namespace` (optional): Pod namespace (required if pod_name specified)
- `duration` (optional): Capture duration (default: 60s)
- `filter` (optional): Tcpdump filter expression
- `output_dir` (optional): Custom output directory

**Examples:**

Pod-level capture:
```json
{
  "tool": "collect_tcpdump",
  "parameters": {
    "pod_name": "myapp-123",
    "namespace": "production",
    "duration": "2m",
    "filter": "port 80 or port 443"
  }
}
```

Node-level capture:
```json
{
  "tool": "collect_tcpdump",
  "parameters": {
    "node_name": "worker-node-01",
    "duration": "5m",
    "filter": "host 10.0.1.100"
  }
}
```

### 3. Log Collection (`collect_logs`)

Collects comprehensive logs from pods, containers, and system components.

**Parameters:**
- `pod_name` (optional): Specific pod to collect logs from
- `namespace` (optional): Namespace to collect logs from
- `include_previous` (optional): Include previous container logs
- `output_dir` (optional): Custom output directory

**Examples:**

Specific pod logs:
```json
{
  "tool": "collect_logs",
  "parameters": {
    "pod_name": "myapp-123",
    "namespace": "production",
    "include_previous": true
  }
}
```

Namespace-wide logs:
```json
{
  "tool": "collect_logs",
  "parameters": {
    "namespace": "openshift-monitoring"
  }
}
```

### 4. Must-Gather Collection (`openshift_must_gather`)

Collects OpenShift must-gather data using official Red Hat tools.

**Parameters:**
- `image` (optional): Must-gather image to use
- `dest_dir` (optional): Destination directory

## Analysis Tools

### 1. Must-Gather Analysis (`analyze_must_gather`)

Analyzes collected must-gather data to identify cluster issues, similar to how support engineers review must-gather archives.

**Parameters:**
- `must_gather_path` (required): Path to must-gather directory

**Analysis includes:**
- Cluster version and health status
- Node conditions and health
- Pod issues (CrashLoopBackOff, ImagePullBackOff, etc.)
- Cluster events analysis
- Operator logs examination
- Critical/warning/info issue categorization

### 2. Log Analysis (`analyze_logs`)

Analyzes log files using pattern matching to identify common issues and errors.

**Parameters:**
- `log_path` (required): Path to log file or directory

**Analysis includes:**
- Out of memory errors
- Network connectivity issues
- DNS resolution failures
- Disk space problems
- Permission/security issues
- Operator reconciliation errors

### 3. Network Capture Analysis (`analyze_tcpdump`)

Analyzes packet capture files to identify network issues.

**Parameters:**
- `pcap_path` (required): Path to pcap file

**Analysis includes:**
- Basic file validation
- Packet count and size metrics
- Network flow analysis (when tshark available)
- Connection pattern identification

## LLM Integration

The diagnostic tools are fully integrated with the LLM system, allowing natural language requests like:

- "Collect must-gather data and analyze it for issues"
- "Capture network traffic on pod myapp-123 for 2 minutes and analyze connectivity"
- "Collect sosreport from worker-node-01 and check for performance issues"
- "Analyze logs from the monitoring namespace for errors"

## Example Workflows

### 1. Pod Crash Investigation

```bash
curl -X POST http://localhost:8080/api/enhanced-chat \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Pod myapp-pod-123 in namespace production is crashing. Collect logs and analyze the root cause."
  }'
```

### 2. Network Connectivity Issues

```bash
curl -X POST http://localhost:8080/api/enhanced-chat \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Investigate network connectivity issues between pods in the frontend namespace. Capture packets and analyze traffic."
  }'
```

### 3. Node Performance Problems

```bash
curl -X POST http://localhost:8080/api/enhanced-chat \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Worker node compute-1 is showing high CPU usage. Collect sosreport and analyze system performance."
  }'
```

### 4. Comprehensive Cluster Health Check

```bash
curl -X POST http://localhost:8080/api/enhanced-chat \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Perform a comprehensive health check of the OpenShift cluster including must-gather collection and analysis."
  }'
```

## Output Format

All tools provide structured output including:

### Collection Results
- **Summary**: Brief description of what was collected
- **Location**: File path or directory of collected data
- **Duration**: Time taken for collection
- **Size**: Size of collected data in MB
- **Status**: Success/failure status

### Analysis Results
- **Summary**: Overall analysis summary
- **Issues Found**: Categorized by severity (critical, warning, info)
- **Recommendations**: Actionable recommendations
- **Metrics**: Quantitative analysis data

### Issue Format
Each identified issue includes:
- **Severity**: critical, warning, info
- **Category**: component category (pod, node, network, etc.)
- **Title**: Brief issue description
- **Description**: Detailed issue description
- **Location**: Where the issue was found
- **Evidence**: Supporting evidence (log lines, etc.)
- **Resolution**: Recommended resolution steps

## Security Considerations

- SOS report collection requires privileged access to nodes
- Network captures may contain sensitive traffic data
- All collected data should be handled according to security policies
- Temporary debug pods are automatically cleaned up after collection

## Prerequisites

- OpenShift cluster access with appropriate RBAC permissions
- For node-level operations: cluster-admin or equivalent permissions
- For packet captures: CAP_NET_RAW capability in debug pods
- For sosreports: Red Hat support-tools image availability

## Storage Requirements

- Must-gather: 100MB - 1GB depending on cluster size
- SOS reports: 50MB - 500MB per node
- Packet captures: Varies based on duration and traffic volume
- Logs: Varies based on verbosity and time range

The diagnostic storage location can be configured and should have adequate space for multiple concurrent collections.

# Network Troubleshooting Integration

## Overview

The OpenShift MCP Go project now includes comprehensive network troubleshooting capabilities using tcpdump, nsenter, and other network debugging tools. This integration allows users to perform advanced network troubleshooting using natural language prompts without direct SSH access to nodes.

## Features

### 🔍 Automatic Detection
The system automatically detects network troubleshooting queries using an extensive keyword analysis system:

- **tcpdump & Packet Capture**: `tcpdump`, `packet capture`, `wireshark`, `capture packets`
- **Network Connectivity**: `ping`, `connectivity test`, `network test`, `traceroute`
- **DNS Resolution**: `dns resolution`, `nslookup`, `dig`, `dns test`
- **HTTP/HTTPS Testing**: `curl`, `http test`, `https test`
- **Network Statistics**: `netstat`, `ss`, `network connections`, `network statistics`
- **Advanced Debugging**: `network namespace`, `nsenter`, `network debug`, `network troubleshoot`

### 🛠️ Supported Workflows

#### 1. Tcpdump Packet Capture
- **Query**: `"tcpdump on pod my-app-123 in namespace production"`
- **Process**:
  1. Find the node where the pod is running
  2. Launch `oc debug node/<nodename>` session
  3. Find pod ID and network namespace path
  4. Execute tcpdump using nsenter in the pod's network namespace
  5. Save to .pcap file and guide user to copy it

#### 2. Network Connectivity Testing
- **Query**: `"ping from pod nginx-456 to 8.8.8.8"`
- **Process**:
  1. Find the node where the pod is running
  2. Launch debug session
  3. Execute ping using nsenter in the pod's network namespace

#### 3. DNS Resolution Testing
- **Query**: `"test DNS resolution from pod backend-321"`
- **Process**:
  1. Find the node where the pod is running
  2. Launch debug session
  3. Execute nslookup/dig using nsenter in the pod's network namespace

#### 4. HTTP/HTTPS Testing
- **Query**: `"curl from pod web-app-666 to https://api.example.com"`
- **Process**:
  1. Find the node where the pod is running
  2. Launch debug session
  3. Execute curl using nsenter in the pod's network namespace

#### 5. Network Statistics
- **Query**: `"show network connections in pod api-server-777"`
- **Process**:
  1. Find the node where the pod is running
  2. Launch debug session
  3. Execute netstat/ss using nsenter in the pod's network namespace

#### 6. Advanced Network Debugging
- **Query**: `"debug network namespace for pod cache-999"`
- **Process**:
  1. Find the node where the pod is running
  2. Launch debug session
  3. Execute comprehensive network debugging commands

## Technical Implementation

### Code Structure
- **`pkg/decision/network_troubleshooter.go`**: Main network troubleshooting logic
- **`pkg/decision/engine.go`**: Integration with the decision engine
- **`prompts.md`**: Added comprehensive network troubleshooting prompts

### Key Functions

#### `IsNetworkTroubleshootingQuery(query string) bool`
Detects if a query is related to network troubleshooting based on keyword analysis.

#### `handleNetworkTroubleshooting(analysis *types.Analysis) (*types.Analysis, error)`
Main handler that processes network troubleshooting requests and generates appropriate workflows.

#### `extractPodInfo(query string) *PodInfo`
Extracts pod information (name, namespace, interface, etc.) from the query using regex patterns.

#### `generateTcpdumpWorkflow(podInfo *PodInfo) []string`
Generates step-by-step tcpdump workflow instructions.

#### `executeNetworkWorkflow(workflow string, steps []string, podInfo *PodInfo) (string, error)`
Executes the network troubleshooting workflow and generates commands.

## OpenShift/OCP Version Compatibility

The system supports both OpenShift 4.8 and newer versions with automatic detection:

### OpenShift 4.8 and lower:
```bash
pod_id=$(chroot /host crictl pods --namespace ${NAMESPACE} --name ${NAME} -q)
pid=$(chroot /host bash -c "runc state $pod_id | jq .pid")
nsenter_parameters="-n -t $pid"
```

### OpenShift 4.9 and higher:
```bash
pod_id=$(chroot /host crictl pods --namespace ${NAMESPACE} --name ${NAME} -q)
ns_path="/host$(chroot /host bash -c "crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\"network\").path' -r")"
nsenter_parameters="--net=${ns_path}"
```

## Usage Examples

### Basic Tcpdump
```
User: "tcpdump on pod my-app-123 in namespace production"
System: Generates complete tcpdump workflow with nsenter commands
```

### Advanced Packet Capture
```
User: "capture packets from pod nginx-456 interface eth0 port 8080"
System: Generates tcpdump command with specific interface and port filtering
```

### Connectivity Testing
```
User: "ping from pod frontend-789 to backend-service"
System: Generates ping commands executed from the pod's network namespace
```

### DNS Debugging
```
User: "test DNS resolution from pod backend-321"
System: Generates nslookup/dig commands from the pod's perspective
```

## Security Considerations

1. **RBAC**: The system respects existing OpenShift RBAC policies
2. **Node Access**: Uses `oc debug node` which requires appropriate permissions
3. **Network Isolation**: Commands are executed within the pod's network namespace
4. **Audit Trail**: All commands are logged and can be audited

## Generated Commands

The system generates production-ready commands that can be executed directly:

### Find Pod Node
```bash
kubectl get pod my-app-123 -n production -o jsonpath='{.spec.nodeName}'
```

### Debug Node Session
```bash
oc debug node/ip-10-0-1-100.us-west-2.compute.internal
```

### Tcpdump Execution
```bash
pod_id=$(chroot /host crictl pods --namespace production --name my-app-123 -q)
ns_path="/host$(chroot /host bash -c "crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\"network\").path' -r")"
nsenter --net="$ns_path" -- tcpdump -nn -i eth0 port 8080 -w /host/var/tmp/capture.pcap
```

### File Copy
```bash
oc cp <debug-pod>:/var/tmp/capture.pcap ./capture.pcap
```

## Integration with Decision Engine

The network troubleshooter is fully integrated with the existing decision engine:

1. **Query Detection**: Automatically detects network troubleshooting queries
2. **Workflow Generation**: Creates appropriate workflows based on query type
3. **Command Execution**: Generates safe, executable commands
4. **Prompt Categorization**: Automatically categorizes prompts for analytics
5. **Evidence Collection**: Collects evidence for troubleshooting analysis

## Testing

Comprehensive tests are available:
- **`test-network-detection.sh`**: Tests query detection logic
- **`test-network-integration-full.sh`**: Full integration testing
- **Unit tests**: Available in `pkg/decision/` test suite

## Future Enhancements

1. **Real-time Packet Analysis**: Integration with packet analysis tools
2. **Network Policy Debugging**: Automated network policy troubleshooting
3. **Service Mesh Integration**: Support for Istio/OpenShift Service Mesh debugging
4. **Performance Monitoring**: Network performance analysis capabilities
5. **Automated Remediation**: Automatic fixing of common network issues

## Troubleshooting

### Common Issues

1. **Pod Not Found**: Verify pod name and namespace
2. **Node Access Denied**: Check RBAC permissions for `oc debug node`
3. **Network Namespace Not Found**: Ensure pod is running and not in terminated state
4. **Tcpdump Not Available**: The debug image includes tcpdump by default

### Debug Commands

```bash
# Check pod status
kubectl get pod <pod-name> -n <namespace>

# Verify node access
oc debug node/<node-name> --dry-run

# Check network namespace
crictl pods --namespace <namespace> --name <pod-name> -q
```

---

## Conclusion

The network troubleshooting integration provides a comprehensive solution for OpenShift network debugging using natural language prompts. It combines the power of tcpdump, nsenter, and other network tools with the convenience of conversational AI, making network troubleshooting accessible to both technical and non-technical users.

The system is production-ready, secure, and fully integrated with the existing OpenShift MCP Go infrastructure.

package decision

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/types"
	"github.com/sirupsen/logrus"
)

// NetworkTroubleshooter handles advanced network troubleshooting workflows
type NetworkTroubleshooter struct {
	engine *Engine
}

// NewNetworkTroubleshooter creates a new network troubleshooter
func NewNetworkTroubleshooter(engine *Engine) *NetworkTroubleshooter {
	return &NetworkTroubleshooter{
		engine: engine,
	}
}

// IsNetworkTroubleshootingQuery checks if the query is for network troubleshooting
func (nt *NetworkTroubleshooter) IsNetworkTroubleshootingQuery(query string) bool {
	networkKeywords := []string{
		"tcpdump", "packet capture", "network capture", "wireshark",
		"nsenter", "network namespace", "netns", "ping from pod",
		"traceroute", "nslookup", "dig", "curl from pod",
		"network debug", "network troubleshoot", "capture packets",
		"network analysis", "packet analysis", "traffic capture",
		"network connectivity", "pod networking", "service networking",
		"dns resolution", "network policy", "firewall", "iptables",
		"network interface", "eth0", "lo", "veth", "bridge",
		"netstat", "ss", "lsof", "netcat", "nc", "telnet",
		"network connections", "show connections", "network statistics",
		"network routes", "ip route", "ip addr", "arp", "ping",
		"connectivity test", "network test", "http test", "https test",
		"dns test", "socket connections", "network config",
	}

	lowerQuery := strings.ToLower(query)
	for _, keyword := range networkKeywords {
		if strings.Contains(lowerQuery, keyword) {
			return true
		}
	}
	return false
}

// handleNetworkTroubleshooting handles network troubleshooting requests
func (nt *NetworkTroubleshooter) handleNetworkTroubleshooting(analysis *types.Analysis) (*types.Analysis, error) {
	query := strings.ToLower(analysis.Query)

	// Extract pod and namespace information
	podInfo := nt.extractPodInfo(analysis.Query)

	// Determine the type of network troubleshooting
	var workflow string
	var steps []string

	switch {
	case strings.Contains(query, "tcpdump") || strings.Contains(query, "packet capture") || strings.Contains(query, "capture packets"):
		workflow = "tcpdump"
		steps = nt.generateTcpdumpWorkflow(podInfo)
	case strings.Contains(query, "ping") || strings.Contains(query, "connectivity"):
		workflow = "ping"
		steps = nt.generatePingWorkflow(podInfo)
	case strings.Contains(query, "dns") || strings.Contains(query, "nslookup") || strings.Contains(query, "dig"):
		workflow = "dns"
		steps = nt.generateDNSWorkflow(podInfo)
	case strings.Contains(query, "curl") || strings.Contains(query, "http"):
		workflow = "http"
		steps = nt.generateHTTPWorkflow(podInfo)
	case strings.Contains(query, "netstat") || strings.Contains(query, "ss") || strings.Contains(query, "lsof"):
		workflow = "netstat"
		steps = nt.generateNetstatWorkflow(podInfo)
	default:
		workflow = "general"
		steps = nt.generateGeneralNetworkWorkflow(podInfo)
	}

	// Execute the workflow
	result, err := nt.executeNetworkWorkflow(workflow, steps, podInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to execute network workflow: %w", err)
	}

	// Update analysis with results
	analysis.Response = result
	analysis.Confidence = 0.9
	analysis.Metadata["workflow"] = workflow
	analysis.Metadata["pod_info"] = podInfo
	analysis.Metadata["steps"] = steps

	// Add evidence
	analysis.Evidence = append(analysis.Evidence, types.Evidence{
		Type:      "network_troubleshooting",
		Source:    fmt.Sprintf("pod:%s namespace:%s", podInfo.Name, podInfo.Namespace),
		Content:   fmt.Sprintf("Network troubleshooting workflow: %s", workflow),
		Timestamp: time.Now(),
	})

	return analysis, nil
}

// PodInfo contains information about the pod for network troubleshooting
type PodInfo struct {
	Name      string
	Namespace string
	Node      string
	Interface string
	Command   string
	Args      []string
}

// extractPodInfo extracts pod information from the query
func (nt *NetworkTroubleshooter) extractPodInfo(query string) *PodInfo {
	info := &PodInfo{
		Interface: "eth0", // default interface
	}

	// Extract pod name (look for patterns like "pod my-app-123", "my-app-123 pod", "on <pod>", "for <pod>")
	podPatterns := []string{
		`pod\s+([a-zA-Z0-9-]+)`,
		`([a-zA-Z0-9-]+)\s+pod`,
		`pod/([a-zA-Z0-9-]+)`,
		`on\s+([a-zA-Z0-9-]+)`,
		`for\s+([a-zA-Z0-9-]+)`,
	}

	stopwords := map[string]bool{"in": true, "on": true, "for": true, "with": true, "which": true, "and": true, "the": true}

	for _, pattern := range podPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(query); matches != nil {
			candidate := matches[1]
			if !stopwords[strings.ToLower(candidate)] {
				info.Name = candidate
				break
			}
		}
	}

	// Extract namespace
	nsPatterns := []string{
		`namespace\s+([a-zA-Z0-9-]+)`,
		`ns\s+([a-zA-Z0-9-]+)`,
		`-n\s+([a-zA-Z0-9-]+)`,
		`in\s+([a-zA-Z0-9-]+)\s+namespace`,
	}

	for _, pattern := range nsPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(query); matches != nil {
			info.Namespace = matches[1]
			break
		}
	}

	// Extract interface
	ifPatterns := []string{
		`interface\s+([a-zA-Z0-9-]+)`,
		`-i\s+([a-zA-Z0-9-]+)`,
		`eth\d+`,
		`lo`,
	}

	for _, pattern := range ifPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(query); matches != nil {
			info.Interface = matches[1]
			break
		}
	}

	// Extract command and args for tcpdump
	if strings.Contains(strings.ToLower(query), "tcpdump") {
		info.Command = "tcpdump"
		// Extract tcpdump arguments
		if strings.Contains(query, "port") {
			portPattern := `port\s+(\d+)`
			if matches := regexp.MustCompile(portPattern).FindStringSubmatch(query); matches != nil {
				info.Args = append(info.Args, "port", matches[1])
			}
		}
		if strings.Contains(query, "host") {
			hostPattern := `host\s+([a-zA-Z0-9.-]+)`
			if matches := regexp.MustCompile(hostPattern).FindStringSubmatch(query); matches != nil {
				info.Args = append(info.Args, "host", matches[1])
			}
		}
	}

	return info
}

// generateTcpdumpWorkflow generates steps for tcpdump packet capture
func (nt *NetworkTroubleshooter) generateTcpdumpWorkflow(podInfo *PodInfo) []string {
	steps := []string{
		"1. Find the node where the pod is running",
		"2. Launch an 'oc debug node' session",
		"3. Find the pod ID and network namespace path",
		"4. Execute tcpdump using nsenter in the pod's network namespace",
		"5. Capture packets and save to .pcap file",
		"6. Guide user to copy the .pcap file for analysis",
	}

	if podInfo.Name != "" {
		steps = append(steps, fmt.Sprintf("   - Target pod: %s", podInfo.Name))
	}
	if podInfo.Namespace != "" {
		steps = append(steps, fmt.Sprintf("   - Target namespace: %s", podInfo.Namespace))
	}
	if podInfo.Interface != "" {
		steps = append(steps, fmt.Sprintf("   - Target interface: %s", podInfo.Interface))
	}

	return steps
}

// generatePingWorkflow generates steps for ping connectivity test
func (nt *NetworkTroubleshooter) generatePingWorkflow(podInfo *PodInfo) []string {
	return []string{
		"1. Find the node where the pod is running",
		"2. Launch an 'oc debug node' session",
		"3. Find the pod ID and network namespace path",
		"4. Execute ping using nsenter in the pod's network namespace",
		"5. Test connectivity to specified target",
	}
}

// generateDNSWorkflow generates steps for DNS resolution test
func (nt *NetworkTroubleshooter) generateDNSWorkflow(podInfo *PodInfo) []string {
	return []string{
		"1. Find the node where the pod is running",
		"2. Launch an 'oc debug node' session",
		"3. Find the pod ID and network namespace path",
		"4. Execute nslookup/dig using nsenter in the pod's network namespace",
		"5. Test DNS resolution from pod's perspective",
	}
}

// generateHTTPWorkflow generates steps for HTTP connectivity test
func (nt *NetworkTroubleshooter) generateHTTPWorkflow(podInfo *PodInfo) []string {
	return []string{
		"1. Find the node where the pod is running",
		"2. Launch an 'oc debug node' session",
		"3. Find the pod ID and network namespace path",
		"4. Execute curl using nsenter in the pod's network namespace",
		"5. Test HTTP connectivity from pod's perspective",
	}
}

// generateNetstatWorkflow generates steps for network statistics
func (nt *NetworkTroubleshooter) generateNetstatWorkflow(podInfo *PodInfo) []string {
	return []string{
		"1. Find the node where the pod is running",
		"2. Launch an 'oc debug node' session",
		"3. Find the pod ID and network namespace path",
		"4. Execute netstat/ss using nsenter in the pod's network namespace",
		"5. Show network connections and statistics",
	}
}

// generateGeneralNetworkWorkflow generates steps for general network troubleshooting
func (nt *NetworkTroubleshooter) generateGeneralNetworkWorkflow(podInfo *PodInfo) []string {
	return []string{
		"1. Find the node where the pod is running",
		"2. Launch an 'oc debug node' session",
		"3. Find the pod ID and network namespace path",
		"4. Execute network troubleshooting commands using nsenter",
		"5. Analyze network configuration and connectivity",
	}
}

// executeNetworkWorkflow executes the network troubleshooting workflow
func (nt *NetworkTroubleshooter) executeNetworkWorkflow(workflow string, steps []string, podInfo *PodInfo) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("🔍 **Network Troubleshooting Workflow: %s**\n\n", strings.Title(workflow)))
	result.WriteString("**Steps to follow:**\n")
	for _, step := range steps {
		result.WriteString(fmt.Sprintf("%s\n", step))
	}
	result.WriteString("\n")

	// If we have pod information, try to execute some commands
	if podInfo.Name != "" && podInfo.Namespace != "" {
		result.WriteString("**Executing commands:**\n\n")

		// Step 1: Find the node where the pod is running
		nodeResult, err := nt.findPodNode(podInfo)
		if err != nil {
			result.WriteString(fmt.Sprintf("❌ Error finding pod node: %v\n", err))
			result.WriteString("\n**Troubleshooting tips:**\n")
			result.WriteString("- Double-check the pod name and namespace.\n")
			result.WriteString("- Make sure the pod is running and not terminated.\n")
			result.WriteString("- Ensure your kubeconfig context is correct and you have cluster access.\n")
			result.WriteString("- Try: kubectl get pod <pod-name> -n <namespace>\n")
			return result.String(), nil
		} else {
			result.WriteString(fmt.Sprintf("✅ Pod node: %s\n", nodeResult))
			podInfo.Node = nodeResult
		}

		// Step 2: Generate the debug commands
		commands := nt.generateDebugCommands(workflow, podInfo)
		result.WriteString("\n**Commands to execute:**\n")
		for i, cmd := range commands {
			result.WriteString(fmt.Sprintf("%d. ```bash\n%s\n```\n", i+1, cmd))
		}

		// Step 3: Execute some safe commands if possible
		if workflow == "tcpdump" {
			tcpdumpResult, err := nt.executeTcpdumpWorkflow(podInfo)
			if err != nil {
				result.WriteString(fmt.Sprintf("\n❌ Error executing tcpdump workflow: %v\n", err))
			} else {
				result.WriteString(fmt.Sprintf("\n✅ Tcpdump workflow result:\n%s\n", tcpdumpResult))
			}
		}
	} else {
		result.WriteString("**Note:** Please provide pod name and namespace for automated execution.\n")
		result.WriteString("Example: `tcpdump on pod my-app-123 in namespace production`\n")
	}

	return result.String(), nil
}

// findPodNode finds the node where the pod is running
func (nt *NetworkTroubleshooter) findPodNode(podInfo *PodInfo) (string, error) {
	cmdArgs := []string{"get", "pod", podInfo.Name, "-n", podInfo.Namespace, "-o", "jsonpath={.spec.nodeName}"}
	cmd := exec.Command("kubectl", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		// Try to get stderr for more diagnostics
		exitErr, ok := err.(*exec.ExitError)
		stderr := ""
		if ok {
			stderr = string(exitErr.Stderr)
		}
		return "", fmt.Errorf("Pod '%s' in namespace '%s' was not found or is not running.\n\nTROUBLESHOOTING DETAILS:\n- Command: kubectl %s\n- Stdout: %s\n- Stderr: %s\n\nPlease check the pod name and namespace, and ensure you have access to the cluster.",
			podInfo.Name, podInfo.Namespace, strings.Join(cmdArgs, " "), string(output), stderr)
	}

	return strings.TrimSpace(string(output)), nil
}

// generateDebugCommands generates the debug commands for the workflow
func (nt *NetworkTroubleshooter) generateDebugCommands(workflow string, podInfo *PodInfo) []string {
	var commands []string

	// Base debug node command
	debugNodeCmd := fmt.Sprintf("oc debug node/%s", podInfo.Node)
	commands = append(commands, debugNodeCmd)

	// Inside the debug node session
	switch workflow {
	case "tcpdump":
		commands = append(commands, nt.generateTcpdumpCommands(podInfo)...)
	case "ping":
		commands = append(commands, nt.generatePingCommands(podInfo)...)
	case "dns":
		commands = append(commands, nt.generateDNSCommands(podInfo)...)
	case "http":
		commands = append(commands, nt.generateHTTPCommands(podInfo)...)
	case "netstat":
		commands = append(commands, nt.generateNetstatCommands(podInfo)...)
	default:
		commands = append(commands, nt.generateGeneralNetworkCommands(podInfo)...)
	}

	return commands
}

// generateTcpdumpCommands generates tcpdump-specific commands
func (nt *NetworkTroubleshooter) generateTcpdumpCommands(podInfo *PodInfo) []string {
	commands := []string{
		"# Find the pod ID using crictl",
		fmt.Sprintf("pod_id=$(chroot /host crictl pods --namespace %s --name %s -q)", podInfo.Namespace, podInfo.Name),
		"echo \"Pod ID: $pod_id\"",
		"",
		"# Find the network namespace path",
		"if crictl inspectp \"$pod_id\" | grep -q 'runtimeSpec'; then",
		"  ns_path=\"/host$(chroot /host bash -c \"crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\\\"network\\\").path' -r\")\"",
		"else",
		"  pid=$(chroot /host bash -c \"runc state $pod_id | jq .pid\")",
		"  ns_path=\"/proc/$pid/ns/net\"",
		"fi",
		"echo \"Network namespace path: $ns_path\"",
		"",
		"# Execute tcpdump in the pod's network namespace",
	}

	// Build tcpdump command
	tcpdumpCmd := fmt.Sprintf("nsenter --net=\"$ns_path\" -- tcpdump -nn -i %s", podInfo.Interface)
	if len(podInfo.Args) > 0 {
		tcpdumpCmd += " " + strings.Join(podInfo.Args, " ")
	}

	// Add pcap file output
	pcapFile := fmt.Sprintf("capture-%s-%d.pcap", podInfo.Name, time.Now().Unix())
	tcpdumpCmd += fmt.Sprintf(" -w /host/var/tmp/%s", pcapFile)

	commands = append(commands, tcpdumpCmd)
	commands = append(commands, "")
	commands = append(commands, "# To stop tcpdump, press Ctrl+C")
	commands = append(commands, fmt.Sprintf("# The capture file will be saved as: /var/tmp/%s", pcapFile))
	commands = append(commands, "# Copy the file using: oc cp node-debug-pod:/var/tmp/"+pcapFile+" ./"+pcapFile)

	return commands
}

// generatePingCommands generates ping-specific commands
func (nt *NetworkTroubleshooter) generatePingCommands(podInfo *PodInfo) []string {
	return []string{
		"# Find the pod ID and network namespace",
		fmt.Sprintf("pod_id=$(chroot /host crictl pods --namespace %s --name %s -q)", podInfo.Namespace, podInfo.Name),
		"ns_path=\"/host$(chroot /host bash -c \"crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\\\"network\\\").path' -r\")\"",
		"",
		"# Execute ping in the pod's network namespace",
		"nsenter --net=\"$ns_path\" -- ping -c 4 <target_host>",
		"",
		"# Test connectivity to common services",
		"nsenter --net=\"$ns_path\" -- ping -c 4 8.8.8.8",
		"nsenter --net=\"$ns_path\" -- ping -c 4 kubernetes.default.svc.cluster.local",
	}
}

// generateDNSCommands generates DNS-specific commands
func (nt *NetworkTroubleshooter) generateDNSCommands(podInfo *PodInfo) []string {
	return []string{
		"# Find the pod ID and network namespace",
		fmt.Sprintf("pod_id=$(chroot /host crictl pods --namespace %s --name %s -q)", podInfo.Namespace, podInfo.Name),
		"ns_path=\"/host$(chroot /host bash -c \"crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\\\"network\\\").path' -r\")\"",
		"",
		"# Test DNS resolution",
		"nsenter --net=\"$ns_path\" -- nslookup kubernetes.default.svc.cluster.local",
		"nsenter --net=\"$ns_path\" -- nslookup google.com",
		"",
		"# Check DNS configuration",
		"nsenter --net=\"$ns_path\" -- cat /etc/resolv.conf",
	}
}

// generateHTTPCommands generates HTTP-specific commands
func (nt *NetworkTroubleshooter) generateHTTPCommands(podInfo *PodInfo) []string {
	return []string{
		"# Find the pod ID and network namespace",
		fmt.Sprintf("pod_id=$(chroot /host crictl pods --namespace %s --name %s -q)", podInfo.Namespace, podInfo.Name),
		"ns_path=\"/host$(chroot /host bash -c \"crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\\\"network\\\").path' -r\")\"",
		"",
		"# Test HTTP connectivity",
		"nsenter --net=\"$ns_path\" -- curl -v http://<target_service>",
		"nsenter --net=\"$ns_path\" -- curl -v https://kubernetes.default.svc.cluster.local",
	}
}

// generateNetstatCommands generates netstat-specific commands
func (nt *NetworkTroubleshooter) generateNetstatCommands(podInfo *PodInfo) []string {
	return []string{
		"# Find the pod ID and network namespace",
		fmt.Sprintf("pod_id=$(chroot /host crictl pods --namespace %s --name %s -q)", podInfo.Namespace, podInfo.Name),
		"ns_path=\"/host$(chroot /host bash -c \"crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\\\"network\\\").path' -r\")\"",
		"",
		"# Show network connections and statistics",
		"nsenter --net=\"$ns_path\" -- netstat -tulnp",
		"nsenter --net=\"$ns_path\" -- ss -tulnp",
		"nsenter --net=\"$ns_path\" -- ip addr show",
		"nsenter --net=\"$ns_path\" -- ip route show",
	}
}

// generateGeneralNetworkCommands generates general network commands
func (nt *NetworkTroubleshooter) generateGeneralNetworkCommands(podInfo *PodInfo) []string {
	return []string{
		"# Find the pod ID and network namespace",
		fmt.Sprintf("pod_id=$(chroot /host crictl pods --namespace %s --name %s -q)", podInfo.Namespace, podInfo.Name),
		"ns_path=\"/host$(chroot /host bash -c \"crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\\\"network\\\").path' -r\")\"",
		"",
		"# General network troubleshooting",
		"nsenter --net=\"$ns_path\" -- ip addr show",
		"nsenter --net=\"$ns_path\" -- ip route show",
		"nsenter --net=\"$ns_path\" -- netstat -i",
		"nsenter --net=\"$ns_path\" -- arp -a",
	}
}

// executeTcpdumpWorkflow executes the tcpdump workflow if safe to do so
func (nt *NetworkTroubleshooter) executeTcpdumpWorkflow(podInfo *PodInfo) (string, error) {
	// This is a safe dry-run that shows what would be executed
	// In a production environment, you might want to actually execute some commands

	var result strings.Builder
	result.WriteString("🔍 **Tcpdump Workflow Execution:**\n\n")

	// Check if the pod exists
	checkCmd := exec.Command("kubectl", "get", "pod", podInfo.Name, "-n", podInfo.Namespace)
	if err := checkCmd.Run(); err != nil {
		return "", fmt.Errorf("pod %s not found in namespace %s", podInfo.Name, podInfo.Namespace)
	}

	result.WriteString("✅ Pod exists and is accessible\n")

	// Generate the script that would be executed
	scriptPath := nt.generateTcpdumpScript(podInfo)
	result.WriteString(fmt.Sprintf("📝 Generated tcpdump script: %s\n", scriptPath))

	return result.String(), nil
}

// generateTcpdumpScript generates a script for tcpdump execution
func (nt *NetworkTroubleshooter) generateTcpdumpScript(podInfo *PodInfo) string {
	// Create a temporary script file
	tmpDir := "/tmp/openshift-mcp-tcpdump"
	os.MkdirAll(tmpDir, 0755)

	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("tcpdump-%s-%d.sh", podInfo.Name, time.Now().Unix()))

	scriptContent := fmt.Sprintf(`#!/bin/bash
# Generated tcpdump script for pod: %s
# Namespace: %s
# Node: %s
# Interface: %s

set -euo pipefail

POD_NAME="%s"
POD_NAMESPACE="%s"
INTERFACE="%s"
PCAP_FILE="capture-${POD_NAME}-$(date +%%s).pcap"

echo "Starting tcpdump workflow for pod: $POD_NAME"
echo "Namespace: $POD_NAMESPACE"
echo "Interface: $INTERFACE"
echo "Output file: $PCAP_FILE"

# Find pod ID
pod_id=$(chroot /host crictl pods --namespace "$POD_NAMESPACE" --name "$POD_NAME" -q)
if [ -z "$pod_id" ]; then
	echo "Error: Pod not found"
	exit 1
fi

echo "Found pod ID: $pod_id"

# Find network namespace path
if crictl inspectp "$pod_id" | grep -q 'runtimeSpec'; then
	ns_path="/host$(chroot /host bash -c "crictl inspectp $pod_id | jq '.info.runtimeSpec.linux.namespaces[]|select(.type==\"network\").path' -r")"
else
	pid=$(chroot /host bash -c "runc state $pod_id | jq .pid")
	ns_path="/proc/$pid/ns/net"
fi

echo "Network namespace path: $ns_path"

# Execute tcpdump
echo "Starting tcpdump..."
nsenter --net="$ns_path" -- tcpdump -nn -i "$INTERFACE" %s -w "/host/var/tmp/$PCAP_FILE"

echo "Tcpdump completed. File saved as: /var/tmp/$PCAP_FILE"
echo "To copy the file, run: oc cp <debug-pod>:/var/tmp/$PCAP_FILE ./$PCAP_FILE"
`, podInfo.Name, podInfo.Namespace, podInfo.Node, podInfo.Interface,
		podInfo.Name, podInfo.Namespace, podInfo.Interface, strings.Join(podInfo.Args, " "))

	// Write the script to file
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		logrus.WithError(err).Warn("Failed to write tcpdump script")
		return ""
	}

	return scriptPath
}

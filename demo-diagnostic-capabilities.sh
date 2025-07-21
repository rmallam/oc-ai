#!/bin/bash

# Enhanced Support Engineer Capabilities Demo
# This script demonstrates the new diagnostic and analysis capabilities

echo "🔧 Enhanced OpenShift Support Engineer Capabilities"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Available Diagnostic Collection Tools:${NC}"
echo "• collect_sosreport     - Collect sosreport from nodes"
echo "• collect_tcpdump      - Network packet capture"
echo "• collect_logs         - Comprehensive log collection"
echo "• openshift_must_gather - OpenShift must-gather data"
echo ""

echo -e "${BLUE}Available Analysis Tools:${NC}"
echo "• analyze_must_gather  - Analyze must-gather data"
echo "• analyze_logs         - Analyze log files for issues"
echo "• analyze_tcpdump      - Analyze network captures"
echo ""

echo -e "${YELLOW}Example Usage Scenarios:${NC}"
echo ""

echo -e "${GREEN}1. Collect and Analyze Must-Gather:${NC}"
echo "   curl -X POST http://localhost:8080/api/enhanced-chat \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"prompt\": \"collect must-gather data and analyze it for issues\"}'"
echo ""

echo -e "${GREEN}2. Network Troubleshooting:${NC}"
echo "   curl -X POST http://localhost:8080/api/enhanced-chat \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"prompt\": \"capture network traffic on pod myapp-123 in namespace production for 2 minutes and analyze for connectivity issues\"}'"
echo ""

echo -e "${GREEN}3. Pod Crash Investigation:${NC}"
echo "   curl -X POST http://localhost:8080/api/enhanced-chat \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"prompt\": \"collect logs from crashed pod and analyze for root cause\"}'"
echo ""

echo -e "${GREEN}4. Node Performance Issues:${NC}"
echo "   curl -X POST http://localhost:8080/api/enhanced-chat \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"prompt\": \"collect sosreport from worker-node-01 and analyze system performance\"}'"
echo ""

echo -e "${GREEN}5. Comprehensive Cluster Health Check:${NC}"
echo "   curl -X POST http://localhost:8080/api/enhanced-chat \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"prompt\": \"perform comprehensive cluster analysis including must-gather, node diagnostics, and log analysis\"}'"
echo ""

echo -e "${BLUE}Tool Parameters:${NC}"
echo ""

echo -e "${YELLOW}collect_sosreport:${NC}"
echo "  • node_name (required)  - Target node name"
echo "  • output_dir (optional) - Custom output directory"
echo ""

echo -e "${YELLOW}collect_tcpdump:${NC}"
echo "  • pod_name OR node_name - Target for capture"
echo "  • namespace            - Pod namespace (if pod_name used)"
echo "  • duration             - Capture duration (e.g., 60s, 5m)"
echo "  • filter               - Tcpdump filter expression"
echo "  • output_dir           - Custom output directory"
echo ""

echo -e "${YELLOW}collect_logs:${NC}"
echo "  • pod_name (optional)   - Specific pod to collect from"
echo "  • namespace (optional)  - Namespace to collect from"
echo "  • include_previous      - Include previous container logs"
echo "  • output_dir           - Custom output directory"
echo ""

echo -e "${YELLOW}analyze_* tools:${NC}"
echo "  • Path parameter pointing to collected data"
echo ""

echo -e "${RED}Starting Server...${NC}"
echo "Use the examples above to test the enhanced capabilities!"
echo ""

# Start the server
./bin/server --config config/llm_config.yaml

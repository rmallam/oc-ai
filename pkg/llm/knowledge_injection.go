package llm

import (
	"fmt"
	"strings"
)

// KnowledgeInjector provides advanced knowledge injection strategies for making Gemini an OpenShift expert
type KnowledgeInjector struct {
	coreKnowledge           string
	troubleshootingPatterns string
	commandReference        string
	securityPatterns        string
	performancePatterns     string
	incidentPatterns        string
}

// NewKnowledgeInjector creates a new knowledge injector with comprehensive OpenShift expertise
func NewKnowledgeInjector() *KnowledgeInjector {
	return &KnowledgeInjector{
		coreKnowledge:           getOpenShiftCoreKnowledge(),
		troubleshootingPatterns: getOpenShiftTroubleshootingPatterns(),
		commandReference:        getOpenShiftCommandReference(),
		securityPatterns:        getOpenShiftSecurityPatterns(),
		performancePatterns:     getOpenShiftPerformancePatterns(),
		incidentPatterns:        getOpenShiftIncidentPatterns(),
	}
}

// InjectOpenShiftKnowledge injects comprehensive OpenShift knowledge into a prompt
func (ki *KnowledgeInjector) InjectOpenShiftKnowledge(userQuery string) string {
	return fmt.Sprintf(`%s

%s

%s

Now, using the above OpenShift knowledge base, troubleshooting patterns, and command reference, please address this SRE request:

USER REQUEST: %s

Provide a comprehensive response using the knowledge and patterns above. Include specific commands, step-by-step procedures, and reference the relevant concepts from the knowledge base.`,
		ki.coreKnowledge,
		ki.troubleshootingPatterns,
		ki.commandReference,
		userQuery)
}

// InjectSpecializedKnowledge injects domain-specific knowledge for specialized requests
func (ki *KnowledgeInjector) InjectSpecializedKnowledge(userQuery, requestType string, context map[string]string) string {
	var specializedKnowledge string

	switch strings.ToLower(requestType) {
	case "security":
		specializedKnowledge = ki.securityPatterns
	case "performance":
		specializedKnowledge = ki.performancePatterns
	case "incident":
		specializedKnowledge = ki.incidentPatterns
	default:
		specializedKnowledge = ki.troubleshootingPatterns
	}

	contextStr := ""
	if len(context) > 0 {
		contextStr = "\nADDITIONAL CONTEXT:\n"
		for key, value := range context {
			if value != "" {
				contextStr += fmt.Sprintf("%s: %s\n", strings.ToUpper(key), value)
			}
		}
	}

	return fmt.Sprintf(`%s

%s

%s

%s
%s

USER REQUEST: %s

Provide a comprehensive, specialized response using the above knowledge and patterns. Include specific commands, step-by-step procedures, and detailed analysis relevant to this %s scenario.`,
		ki.coreKnowledge,
		ki.troubleshootingPatterns,
		ki.commandReference,
		specializedKnowledge,
		contextStr,
		userQuery,
		requestType)
}

// getOpenShiftCoreKnowledge returns comprehensive OpenShift core concepts
func getOpenShiftCoreKnowledge() string {
	return `
OPENSHIFT CORE CONCEPTS REFERENCE:

## Pods and Containers
- Pod: Smallest deployable unit, contains one or more containers
- Container lifecycle: Init containers run before app containers
- Pod phases: Pending, Running, Succeeded, Failed, Unknown
- Restart policies: Always, OnFailure, Never

Common Pod Issues:
- CrashLoopBackOff: Container fails and restarts repeatedly
  - Check: 'oc logs <pod> -p' (previous container logs)
  - Debug: 'oc describe pod <pod>' for events
  - Fix: Review application logs, resource limits, health checks

- ImagePullBackOff: Cannot pull container image
  - Check: 'oc describe pod <pod>' for image pull errors
  - Fix: Verify image name, registry credentials, network connectivity

- Pending: Pod cannot be scheduled
  - Check: 'oc describe pod <pod>' for scheduling failures
  - Fix: Add worker nodes, check resource requests, node selectors

## Networking
- Services: Stable IP/DNS for pods (ClusterIP, NodePort, LoadBalancer)
- Routes: OpenShift's ingress mechanism (HTTP/HTTPS traffic)
- Ingress: Kubernetes-native ingress (available in OpenShift 4.x)
- NetworkPolicies: Control traffic between pods

Network Troubleshooting:
- DNS issues: 'oc exec <pod> -- nslookup <service>'
- Connectivity: 'oc exec <pod> -- curl <service>:<port>'
- Route problems: 'oc get routes', check TLS certificates
- Service endpoints: 'oc get endpoints <service>'

## Storage
- PersistentVolumes (PV): Cluster storage resources
- PersistentVolumeClaims (PVC): Storage requests from pods
- StorageClasses: Dynamic provisioning templates
- Volume modes: Filesystem, Block

Storage Issues:
- Mount failures: Check PVC status, storage class, node access
- Performance: Monitor IOPS, bandwidth, latency
- Capacity: Set up monitoring for disk usage alerts

## Security
- RBAC: Role-Based Access Control (Users, Groups, Roles, RoleBindings)
- SCCs: Security Context Constraints (Pod security policies)
- ServiceAccounts: Pod identity for API access
- NetworkPolicies: Traffic control between pods

Security Best Practices:
- Use least-privilege RBAC
- Run containers as non-root
- Implement network segmentation
- Regular security scanning

## Cluster Operators
- Authentication: Manages user authentication
- DNS: Cluster DNS resolution
- Ingress: Cluster ingress controllers
- Monitoring: Prometheus and Grafana stack
- Network: Software-defined networking (SDN/OVN)
- Storage: Storage drivers and provisioners
`
}

// getOpenShiftTroubleshootingPatterns returns systematic troubleshooting methodologies
func getOpenShiftTroubleshootingPatterns() string {
	return `
OPENSHIFT TROUBLESHOOTING METHODOLOGY:

## 1. SYSTEMATIC APPROACH
1. **Gather Information**
   - What changed recently?
   - When did the issue start?
   - What is the impact scope?
   - Are there error messages?

2. **Quick Health Checks**
   - Cluster: 'oc get nodes', 'oc get co' (cluster operators)
   - Applications: 'oc get pods --all-namespaces'
   - Events: 'oc get events --sort-by=.metadata.creationTimestamp'

3. **Focused Investigation**
   - Drill down to affected components
   - Collect relevant logs and metrics
   - Test hypotheses systematically

## 2. COMMON ISSUE PATTERNS

### Application Won't Start (CrashLoopBackOff)
Investigation Path:
1. 'oc logs <pod> -p' - Check previous container logs
2. 'oc describe pod <pod>' - Look for events and resource issues
3. 'oc get events -n <namespace>' - Check namespace events
4. 'oc exec <pod> -- <command>' - Interactive debugging (if possible)

Common Causes:
- Application configuration errors
- Missing environment variables or secrets
- Insufficient resources (CPU/memory limits)
- Failed health checks (liveness/readiness probes)
- Image issues or missing dependencies

### Network Connectivity Issues
Investigation Path:
1. 'oc get svc,routes,ingress' - Check service definitions
2. 'oc get endpoints <service>' - Verify service has endpoints
3. 'oc exec <pod> -- nslookup <service>' - Test DNS resolution
4. 'oc exec <pod> -- curl <service>:<port>' - Test connectivity

Common Causes:
- Service selector doesn't match pod labels
- NetworkPolicy blocking traffic
- Route/Ingress misconfiguration
- DNS resolution failures
- Firewall/security group issues

### Storage Problems
Investigation Path:
1. 'oc get pv,pvc' - Check volume status
2. 'oc describe pvc <pvc-name>' - Look for binding issues
3. 'oc get sc' - Verify storage classes
4. 'oc get events | grep -i volume' - Check volume events

Common Causes:
- No available PVs matching PVC requirements
- StorageClass misconfiguration
- Node storage exhaustion
- Permission issues (fsGroup, runAsUser)
- CSI driver problems

### Performance Issues
Investigation Path:
1. 'oc adm top nodes' - Check node resource usage
2. 'oc adm top pods' - Check pod resource usage
3. 'oc describe node <node>' - Check node conditions and capacity
4. 'oc get hpa' - Check horizontal pod autoscalers

Common Causes:
- Resource contention (CPU, memory, storage)
- Inadequate resource limits/requests
- Node overcommitment
- Storage I/O bottlenecks
- Network latency issues
`
}

// getOpenShiftCommandReference returns essential commands for SRE operations
func getOpenShiftCommandReference() string {
	return `
ESSENTIAL OPENSHIFT COMMANDS FOR SRE:

## Cluster Health
oc get nodes                                    # Node status
oc get co                                       # Cluster operators
oc adm top nodes                               # Node resource usage
oc get machinesets -n openshift-machine-api   # Machine sets
oc get clusterversion                          # OpenShift version

## Pod Troubleshooting
oc get pods --all-namespaces                  # All pods
oc get pods -o wide                           # Pods with node info
oc describe pod <pod-name>                    # Detailed pod info
oc logs <pod-name> -f                         # Follow logs
oc logs <pod-name> -p                         # Previous container logs
oc exec -it <pod-name> -- /bin/bash           # Interactive shell
oc debug node/<node-name>                     # Debug node access

## Resource Investigation
oc get events --sort-by=.metadata.creationTimestamp  # Recent events
oc get all -n <namespace>                     # All resources in namespace
oc describe <resource-type> <resource-name>   # Resource details
oc get <resource> -o yaml                     # Resource YAML
oc get <resource> -o json | jq '.status'      # Resource status (with jq)

## Network Debugging
oc get svc,routes,ingress                     # Network resources
oc get endpoints <service-name>               # Service endpoints
oc get networkpolicies                        # Network policies
oc port-forward <pod-name> <local-port>:<pod-port>  # Port forwarding
oc rsh <pod-name>                             # Remote shell access

## Storage Operations
oc get pv,pvc                                 # Persistent volumes
oc get sc                                     # Storage classes
oc describe pvc <pvc-name>                    # PVC details
oc get volumeattachments                      # Volume attachment status

## RBAC and Security
oc whoami                                     # Current user
oc auth can-i <verb> <resource>               # Permission check
oc get rolebindings,clusterrolebindings       # Role bindings
oc get scc                                    # Security context constraints
oc adm policy who-can <verb> <resource>       # Who has permissions

## Performance and Monitoring
oc adm top pods                               # Pod resource usage
oc adm top nodes                              # Node resource usage
oc get hpa                                    # Horizontal Pod Autoscalers
oc get pdb                                    # Pod Disruption Budgets
oc get limits                                 # Resource limits

## Cluster Administration
oc get mc                                     # Machine configs
oc get nodes -l node-role.kubernetes.io/worker  # Worker nodes only
oc get operators                              # Installed operators
oc get csv -A                                # Cluster service versions
oc get installplan -A                        # Install plans

## Advanced Diagnostics
oc adm must-gather                           # Collect cluster data
oc adm inspect <resource>                    # Inspect specific resources
oc get etcdapiserver -o yaml                 # ETCD API server status
oc get kubeapiserver -o yaml                 # Kube API server status
`
}

// getOpenShiftSecurityPatterns returns security-specific knowledge
func getOpenShiftSecurityPatterns() string {
	return `
OPENSHIFT SECURITY PATTERNS AND BEST PRACTICES:

## RBAC (Role-Based Access Control)
### Investigation Commands:
- 'oc get rolebindings,clusterrolebindings' - List all role bindings
- 'oc auth can-i <verb> <resource>' - Check user permissions
- 'oc adm policy who-can <verb> <resource>' - See who has permissions
- 'oc describe clusterrole <role>' - Role details

### Common Security Issues:
1. **Overprivileged Service Accounts**
   - Check: 'oc get sa -A', 'oc describe sa <sa-name>'
   - Fix: Apply principle of least privilege

2. **Missing Network Policies**
   - Check: 'oc get networkpolicies -A'
   - Fix: Implement deny-all default, allow specific traffic

3. **Privileged Containers**
   - Check: 'oc get pods -o jsonpath="{..securityContext}"'
   - Fix: Use non-root users, drop capabilities

## Security Context Constraints (SCCs)
### Investigation:
- 'oc get scc' - List all SCCs
- 'oc describe scc <scc-name>' - SCC details
- 'oc adm policy scc-subject-review -f <pod-spec>' - Check SCC compliance

### Security Review Checklist:
1. Container runs as non-root user
2. Read-only root filesystem where possible
3. No privileged escalation
4. Minimal capabilities granted
5. Resource limits enforced
6. Network policies implemented
7. Service account follows least privilege

## Certificate and TLS Issues
### Investigation:
- 'oc get secrets -A | grep tls' - TLS secrets
- 'oc get routes -o custom-columns=NAME:.metadata.name,TLS:.spec.tls.termination'
- 'openssl x509 -in <cert-file> -text -noout' - Certificate details

### Common TLS Problems:
- Expired certificates
- Wrong certificate chain
- Mismatched certificate names
- Missing intermediate certificates
`
}

// getOpenShiftPerformancePatterns returns performance-specific knowledge
func getOpenShiftPerformancePatterns() string {
	return `
OPENSHIFT PERFORMANCE ANALYSIS PATTERNS:

## Resource Monitoring and Analysis
### Key Metrics to Monitor:
- CPU utilization and throttling
- Memory usage and OOM kills
- Storage IOPS and latency
- Network throughput and packet loss
- Pod startup and scheduling times

### Investigation Commands:
- 'oc adm top nodes --sort-by=cpu' - Node CPU usage
- 'oc adm top pods --sort-by=memory' - Pod memory usage
- 'oc describe node <node>' - Node capacity and allocations
- 'oc get events | grep -i oom' - Out of memory events

## Performance Bottleneck Patterns
### CPU Bottlenecks:
1. **Symptoms**: High CPU usage, slow response times
2. **Investigation**: 
   - 'oc adm top nodes'
   - 'oc get hpa' - Check autoscaling
   - 'oc describe node <node>' - Check CPU pressure
3. **Solutions**: Scale horizontally, optimize code, increase CPU limits

### Memory Bottlenecks:
1. **Symptoms**: OOMKilled events, memory pressure
2. **Investigation**:
   - 'oc get events | grep -i oom'
   - 'oc adm top pods --sort-by=memory'
   - 'oc describe pod <pod>' - Check memory limits
3. **Solutions**: Increase memory limits, optimize memory usage, scale out

### Storage Performance:
1. **Symptoms**: Slow I/O, high latency
2. **Investigation**:
   - 'oc get pv -o custom-columns=NAME:.metadata.name,STORAGECLASS:.spec.storageClassName'
   - 'oc describe pv <pv>' - Check storage backend
3. **Solutions**: Use faster storage class, optimize database queries, implement caching

## Scaling Patterns
### Horizontal Pod Autoscaling (HPA):
- 'oc get hpa' - Current autoscaler status
- 'oc describe hpa <hpa-name>' - Scaling events and metrics
- 'oc autoscale deployment <name> --min=1 --max=10 --cpu-percent=80'

### Vertical Pod Autoscaling (VPA):
- Monitor resource recommendations
- Implement resource right-sizing
- Balance performance vs. cost
`
}

// getOpenShiftIncidentPatterns returns incident response knowledge
func getOpenShiftIncidentPatterns() string {
	return `
OPENSHIFT INCIDENT RESPONSE PATTERNS:

## Incident Classification
### P1 (Critical): Complete service outage
- Cluster down, multiple services affected
- Data loss or corruption
- Security breach

### P2 (High): Major service degradation
- Single critical service down
- Performance severely impacted
- Significant user impact

### P3 (Medium): Minor service issues
- Non-critical service affected
- Partial functionality loss
- Limited user impact

### P4 (Low): Minor issues
- Cosmetic issues
- Documentation problems
- Enhancement requests

## Emergency Response Procedures
### Immediate Actions (First 15 minutes):
1. **Assess Impact**: 'oc get nodes', 'oc get co'
2. **Check Critical Services**: 'oc get pods -A | grep -E "(api|etcd|dns)"'
3. **Review Recent Changes**: Check deployment history, operator updates
4. **Establish Communication**: Update incident channel, stakeholders

### Investigation Phase:
1. **Collect Data**: 'oc adm must-gather'
2. **Analyze Logs**: Focus on error patterns, timing
3. **Check Dependencies**: External services, networking, storage
4. **Test Hypotheses**: Systematic troubleshooting approach

### Recovery Actions:
1. **Immediate Mitigation**: Rollback, traffic routing, scaling
2. **Root Cause Fix**: Address underlying issue
3. **Verification**: Confirm full service restoration
4. **Post-Incident**: Document learnings, improve monitoring

## Critical Service Recovery
### API Server Issues:
- Check etcd cluster health
- Verify master node status
- Review API server logs
- Consider emergency etcd recovery

### Network Outages:
- Check SDN/OVN operator status
- Verify node network connectivity
- Review network policies
- Test DNS resolution

### Storage Failures:
- Check storage operator status
- Verify CSI driver health
- Review PV/PVC status
- Consider emergency volume recovery

## Communication Templates
### Initial Update:
"INCIDENT: <Brief description>
STATUS: Investigating
IMPACT: <Affected services>
ETA: <Next update time>"

### Resolution Update:
"INCIDENT: <Brief description>
STATUS: Resolved
DURATION: <Total time>
CAUSE: <Root cause>
ACTIONS: <Preventive measures>"
`
}

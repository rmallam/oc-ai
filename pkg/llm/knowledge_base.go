package llm

import (
	"fmt"
	"strings"
)

// OpenShiftKnowledgeBase contains comprehensive OpenShift SRE knowledge
type OpenShiftKnowledgeBase struct {
	CoreConcepts            string
	TroubleshootingPatterns string
	CommandReference        string
	SecurityBestPractices   string
	PerformanceTuning       string
	IncidentResponse        string
}

// NewOpenShiftKnowledgeBase creates a new knowledge base instance
func NewOpenShiftKnowledgeBase() *OpenShiftKnowledgeBase {
	return &OpenShiftKnowledgeBase{
		CoreConcepts:            getCoreConcepts(),
		TroubleshootingPatterns: getTroubleshootingPatterns(),
		CommandReference:        getCommandReference(),
		SecurityBestPractices:   getSecurityBestPractices(),
		PerformanceTuning:       getPerformanceTuning(),
		IncidentResponse:        getIncidentResponse(),
	}
}

// getCoreConcepts returns the core OpenShift concepts reference
func getCoreConcepts() string {
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

## Cluster Operations
- Cluster Operators: Manage cluster components
- Machine APIs: Node lifecycle management
- etcd: Cluster state storage
- Control Plane: API server, scheduler, controller manager

Cluster Health Checks:
- 'oc get co' - Cluster operator status
- 'oc get nodes' - Node health
- 'oc get clusterversion' - Update status
- 'oc adm top nodes' - Resource utilization
`
}

// getTroubleshootingPatterns returns systematic troubleshooting methodologies
func getTroubleshootingPatterns() string {
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
1. 'oc adm top nodes' - Node resource usage
2. 'oc adm top pods' - Pod resource usage
3. 'oc describe node <node>' - Node conditions and capacity
4. Check monitoring dashboards (Prometheus/Grafana)

Common Causes:
- Resource exhaustion (CPU, memory, disk)
- Network bandwidth limitations
- Storage performance bottlenecks
- Inefficient application code
- Lack of resource limits causing resource contention
`
}

// getCommandReference returns essential OpenShift commands for SRE
func getCommandReference() string {
	return `
ESSENTIAL OPENSHIFT COMMANDS FOR SRE:

## Cluster Health
oc get nodes                                    # Node status
oc get co                                       # Cluster operators
oc adm top nodes                               # Node resource usage
oc get machinesets -n openshift-machine-api   # Machine sets
oc get clusterversion                          # OpenShift version
oc get clusteroperators                        # Detailed operator status

## Pod Troubleshooting
oc get pods --all-namespaces                  # All pods
oc get pods -o wide                           # Pods with node info
oc describe pod <pod-name>                    # Detailed pod info
oc logs <pod-name> -f                         # Follow logs
oc logs <pod-name> -p                         # Previous container logs
oc logs <pod-name> -c <container>             # Specific container logs
oc exec -it <pod-name> -- /bin/bash           # Interactive shell

## Resource Investigation
oc get events --sort-by=.metadata.creationTimestamp  # Recent events
oc get events --field-selector reason=Failed  # Filter specific events
oc get all -n <namespace>                     # All resources in namespace
oc describe <resource-type> <resource-name>   # Resource details
oc get <resource> -o yaml                     # Resource YAML
oc get <resource> -o jsonpath='{.status}'     # Specific fields

## Network Debugging
oc get svc,routes,ingress                     # Network resources
oc get endpoints <service-name>               # Service endpoints
oc get networkpolicies                        # Network policies
oc port-forward <pod-name> <local-port>:<pod-port>  # Port forwarding
oc get routes -o wide                         # Route details with hosts

## Storage Operations
oc get pv,pvc                                 # Persistent volumes
oc get sc                                     # Storage classes
oc describe pvc <pvc-name>                    # PVC details
oc get volumeattachments                      # Volume attachments

## RBAC and Security
oc whoami                                     # Current user
oc auth can-i <verb> <resource>               # Permission check
oc auth can-i --list                          # List permissions
oc get rolebindings,clusterrolebindings       # Role bindings
oc get scc                                    # Security context constraints
oc adm policy who-can <verb> <resource>       # Who has permissions

## Performance and Monitoring
oc adm top pods                               # Pod resource usage
oc adm top nodes                              # Node resource usage
oc get hpa                                    # Horizontal Pod Autoscalers
oc get vpa                                    # Vertical Pod Autoscalers
oc get metrics                                # Available metrics

## Cluster Administration
oc get mc                                     # Machine configs
oc get nodes -l node-role.kubernetes.io/worker  # Worker nodes only
oc get operators                              # Installed operators
oc get csv -A                                 # Cluster service versions
oc get subscription -A                        # Operator subscriptions

## Troubleshooting Specific Issues
# Certificate issues
oc get secrets --field-selector type=kubernetes.io/tls

# Image pull issues
oc get secret --field-selector type=kubernetes.io/dockerconfigjson

# DNS debugging
oc run test-pod --image=busybox --rm -it -- nslookup kubernetes.default

# Resource quotas
oc get resourcequota
oc describe resourcequota

# Limit ranges
oc get limitrange
oc describe limitrange
`
}

// getSecurityBestPractices returns security guidelines
func getSecurityBestPractices() string {
	return `
OPENSHIFT SECURITY BEST PRACTICES:

## Pod Security Standards
- Always run containers as non-root user
- Set explicit runAsUser and runAsGroup
- Use read-only root filesystem when possible
- Drop unnecessary capabilities
- Set resource limits and requests
- Use security contexts appropriately

## RBAC Best Practices
- Follow principle of least privilege
- Use specific resource names when possible
- Avoid cluster-admin unless absolutely necessary
- Regular RBAC audits and reviews
- Use service accounts for applications
- Implement namespace-based access controls

## Network Security
- Implement NetworkPolicies for pod-to-pod communication
- Use encrypted communication (TLS)
- Restrict ingress and egress traffic
- Monitor network traffic patterns
- Use Routes with proper TLS configuration

## Image Security
- Use trusted base images
- Scan images for vulnerabilities
- Implement image signing and verification
- Use private registries for sensitive workloads
- Regular image updates and patching

## Secrets Management
- Never hardcode secrets in images or configs
- Use OpenShift Secrets or external secret managers
- Rotate secrets regularly
- Limit secret access to necessary pods only
- Monitor secret usage and access patterns
`
}

// getPerformanceTuning returns performance optimization guidelines
func getPerformanceTuning() string {
	return `
OPENSHIFT PERFORMANCE TUNING:

## Resource Management
- Set appropriate CPU and memory limits
- Use requests to guarantee minimum resources
- Implement Horizontal Pod Autoscaler (HPA)
- Use Vertical Pod Autoscaler (VPA) for right-sizing
- Monitor resource utilization patterns

## Node Optimization
- Optimize node sizing for workload requirements
- Use dedicated nodes for specific workloads
- Implement proper node labeling and taints
- Monitor node resource usage and capacity
- Plan for node maintenance and updates

## Storage Performance
- Choose appropriate storage classes for workloads
- Use local storage for performance-critical applications
- Implement proper backup and disaster recovery
- Monitor storage performance metrics
- Optimize persistent volume configurations

## Network Performance
- Use appropriate service types for traffic patterns
- Implement proper load balancing strategies
- Monitor network latency and throughput
- Optimize DNS resolution
- Use appropriate CNI configurations

## Application Optimization
- Implement proper health checks
- Use multi-stage container builds
- Optimize container startup times
- Implement proper logging and monitoring
- Use connection pooling and caching strategies
`
}

// getIncidentResponse returns incident response procedures
func getIncidentResponse() string {
	return `
OPENSHIFT INCIDENT RESPONSE PROCEDURES:

## Immediate Response (0-5 minutes)
1. **Assess Impact**
   - Check cluster health: 'oc get nodes', 'oc get co'
   - Verify affected services and applications
   - Determine user impact and scope

2. **Initial Triage**
   - Gather basic information about the incident
   - Check recent changes or deployments
   - Review monitoring alerts and logs

3. **Communication**
   - Notify stakeholders about the incident
   - Update incident tracking system
   - Establish communication channels

## Investigation Phase (5-30 minutes)
1. **Data Collection**
   - Collect relevant logs and metrics
   - Gather configuration information
   - Document error messages and symptoms

2. **Root Cause Analysis**
   - Follow systematic troubleshooting approach
   - Test hypotheses methodically
   - Isolate affected components

3. **Containment**
   - Prevent further damage or escalation
   - Implement temporary workarounds if possible
   - Preserve evidence for analysis

## Resolution Phase (30+ minutes)
1. **Solution Implementation**
   - Apply fixes based on root cause analysis
   - Test solutions in staging if possible
   - Implement changes incrementally

2. **Verification**
   - Confirm that the issue is resolved
   - Monitor system stability
   - Validate that all services are functioning

3. **Recovery**
   - Restore normal operations
   - Remove temporary workarounds
   - Update documentation and procedures

## Post-Incident Activities
1. **Documentation**
   - Create detailed incident report
   - Document timeline and actions taken
   - Record lessons learned

2. **Follow-up**
   - Conduct post-mortem meeting
   - Implement preventive measures
   - Update monitoring and alerting
   - Review and update procedures
`
}

// InjectKnowledge combines all knowledge areas for comprehensive context
func (kb *OpenShiftKnowledgeBase) InjectKnowledge(userQuery string) string {
	return fmt.Sprintf(`%s

%s

%s

%s

%s

%s

Now, using the above comprehensive OpenShift knowledge base, troubleshooting patterns, command reference, security best practices, performance tuning guidelines, and incident response procedures, please address this SRE request:

USER REQUEST: %s

Provide a comprehensive response using the knowledge and patterns above. Include specific commands, step-by-step procedures, and reference the relevant concepts from the knowledge base.`,
		kb.CoreConcepts,
		kb.TroubleshootingPatterns,
		kb.CommandReference,
		kb.SecurityBestPractices,
		kb.PerformanceTuning,
		kb.IncidentResponse,
		userQuery)
}

// GetSpecializedPrompt creates a specialized prompt for specific SRE scenarios
func (kb *OpenShiftKnowledgeBase) GetSpecializedPrompt(scenario, userQuery string) string {
	systemPrompt := `You are a senior OpenShift Site Reliability Engineer (SRE) with 10+ years of experience managing production OpenShift clusters. You have deep expertise in:

CORE OPENSHIFT KNOWLEDGE:
- OpenShift 4.x architecture (control plane, compute nodes, etcd)
- Container orchestration and Pod lifecycle management
- Networking (SDN, OVN-Kubernetes, Routes, Services, Ingress)
- Storage (Persistent Volumes, Storage Classes, CSI drivers)
- Security (RBAC, SCCs, Network Policies, Pod Security Standards)
- Operators and Operator Lifecycle Manager (OLM)
- GitOps with ArgoCD/Flux on OpenShift
- Monitoring with Prometheus, Grafana, and AlertManager
- Logging with EFK/ELK stack
- Image management and registries

SRE PRACTICES:
- Incident response and troubleshooting methodologies
- Capacity planning and resource optimization
- Performance tuning and scaling strategies
- Backup/restore procedures and disaster recovery
- Change management and deployment strategies
- SLI/SLO definition and monitoring
- Root cause analysis and post-mortem processes

When responding to requests:
1. Provide step-by-step troubleshooting approaches
2. Include specific 'oc' and 'kubectl' commands
3. Reference relevant OpenShift documentation
4. Consider security and compliance implications
5. Suggest monitoring and alerting improvements
6. Provide preventive measures for future occurrences`

	switch strings.ToLower(scenario) {
	case "troubleshooting":
		return kb.buildTroubleshootingPrompt(systemPrompt, userQuery)
	case "security":
		return kb.buildSecurityPrompt(systemPrompt, userQuery)
	case "incident":
		return kb.buildIncidentPrompt(systemPrompt, userQuery)
	case "performance":
		return kb.buildPerformancePrompt(systemPrompt, userQuery)
	default:
		return kb.InjectKnowledge(userQuery)
	}
}

// buildTroubleshootingPrompt creates a troubleshooting-specific prompt
func (kb *OpenShiftKnowledgeBase) buildTroubleshootingPrompt(systemPrompt, userQuery string) string {
	return fmt.Sprintf(`%s

%s

%s

TROUBLESHOOTING REQUEST: %s

Please provide a comprehensive troubleshooting guide following this structure:

## 🔍 IMMEDIATE DIAGNOSTIC STEPS
1. **Initial Assessment**
   - Quick health checks to assess impact
   - Commands to gather critical information
   - Initial containment measures

2. **Targeted Investigation**
   - Specific commands for this issue type
   - Log collection and analysis steps
   - Resource usage verification

## 🛠️ ROOT CAUSE ANALYSIS
- Most likely causes based on symptoms
- How to verify each potential cause
- Decision tree for narrowing down the issue

## 💡 SOLUTION STEPS
- Step-by-step remediation
- Verification commands after each step
- Rollback procedures if needed

## 🚀 PREVENTION MEASURES
- Monitoring alerts to detect early
- Configuration changes to prevent recurrence
- Process improvements

## 📚 RELATED DOCUMENTATION
- Relevant OpenShift documentation links
- Best practices references

Provide specific, actionable commands and avoid generic advice.`,
		systemPrompt, kb.TroubleshootingPatterns, kb.CommandReference, userQuery)
}

// buildSecurityPrompt creates a security-specific prompt
func (kb *OpenShiftKnowledgeBase) buildSecurityPrompt(systemPrompt, userQuery string) string {
	return fmt.Sprintf(`%s

%s

SECURITY REVIEW REQUEST: %s

Perform a comprehensive security analysis following this structure:

## 🔒 SECURITY ASSESSMENT

### 1. **Pod Security Standards**
- SecurityContext configuration
- runAsUser, runAsGroup, fsGroup settings
- Capabilities and privilege escalation
- seccompProfile and SELinux contexts

### 2. **RBAC Analysis**
- Role and ClusterRole permissions
- ServiceAccount assignments
- Principle of least privilege compliance
- Excessive permissions detection

### 3. **Network Security**
- NetworkPolicy configurations
- Service exposure (NodePort, LoadBalancer)
- Ingress/Route security settings
- Inter-pod communication restrictions

### 4. **Resource Security**
- Resource limits and requests
- Storage security (fsGroup, access modes)
- ConfigMap/Secret handling
- Image pull policies and registries

## ⚠️ SECURITY FINDINGS
For each finding:
- **Severity**: Critical/High/Medium/Low
- **Issue**: Specific vulnerability or misconfiguration
- **Impact**: Potential security impact
- **Fix**: Exact configuration changes needed

## ✅ RECOMMENDATIONS
Provide specific security improvements and best practices.

Focus on production-ready, enterprise security standards.`,
		systemPrompt, kb.SecurityBestPractices, userQuery)
}

// buildIncidentPrompt creates an incident response prompt
func (kb *OpenShiftKnowledgeBase) buildIncidentPrompt(systemPrompt, userQuery string) string {
	return fmt.Sprintf(`%s

%s

🚨 INCIDENT RESPONSE REQUEST: %s

Provide immediate incident response guidance following this structure:

## 🚨 IMMEDIATE ACTIONS (First 5 minutes)
1. **Triage Steps**
   - Quick health checks to assess impact
   - Commands to gather critical information
   - Initial containment measures

2. **Escalation Criteria**
   - When to escalate to on-call engineer
   - Communication requirements
   - Stakeholder notification

## 🔍 INVESTIGATION PLAYBOOK (5-15 minutes)
1. **Data Collection**
   - Essential logs to collect
   - Metrics to check
   - Configuration verification

2. **Hypothesis Formation**
   - Most likely causes for this incident type
   - Diagnostic commands for each hypothesis
   - Decision matrix for next steps

## 🛠️ REMEDIATION STRATEGIES (15+ minutes)
1. **Quick Fixes**
   - Immediate workarounds
   - Service restoration steps
   - Impact mitigation

2. **Root Cause Resolution**
   - Systematic problem resolution
   - Verification procedures
   - Service validation

## 📊 MONITORING & COMMUNICATION
- Key metrics to monitor during resolution
- Communication templates for updates
- Success criteria and validation steps

## 📝 POST-INCIDENT CHECKLIST
- Data to preserve for post-mortem
- Initial timeline construction
- Follow-up actions required

Prioritize service restoration while preserving forensic data.`,
		systemPrompt, kb.IncidentResponse, userQuery)
}

// buildPerformancePrompt creates a performance tuning prompt
func (kb *OpenShiftKnowledgeBase) buildPerformancePrompt(systemPrompt, userQuery string) string {
	return fmt.Sprintf(`%s

%s

PERFORMANCE OPTIMIZATION REQUEST: %s

Provide comprehensive performance analysis following this structure:

## 📊 CURRENT STATE ANALYSIS
1. **Resource Utilization Assessment**
   - CPU, Memory, Storage consumption patterns
   - Network bandwidth utilization
   - Node resource allocation efficiency

2. **Performance Bottlenecks**
   - Identified constraints and limits
   - Resource contention points
   - Performance degradation indicators

## 🔧 OPTIMIZATION RECOMMENDATIONS
1. **Resource Optimization**
   - Right-sizing recommendations
   - Resource limit adjustments
   - Scaling strategy improvements

2. **Configuration Tuning**
   - Application-specific optimizations
   - Infrastructure configuration changes
   - Performance monitoring setup

## 🚀 IMPLEMENTATION PLAN
- Step-by-step optimization procedures
- Risk mitigation strategies
- Performance validation methods

## 📈 MONITORING & METRICS
- Key performance indicators to track
- Alerting thresholds and policies
- Long-term performance trending

Provide specific numbers and actionable recommendations.`,
		systemPrompt, kb.PerformanceTuning, userQuery)
}

# OpenShift/Kubernetes Cluster Management Prompts

This document contains comprehensive prompts for day-to-day activities in OpenShift/Kubernetes cluster management. These prompts are designed to work with the OpenShift MCP Go assistant.

## Table of Contents

1. [Cluster Administration](#cluster-administration)
2. [SRE Tasks](#sre-tasks)
3. [Application Deployment](#application-deployment)
4. [Networking](#networking)
5. [Storage Management](#storage-management)
6. [Security & RBAC](#security--rbac)
7. [Monitoring & Observability](#monitoring--observability)
8. [Troubleshooting](#troubleshooting)
9. [Node Management](#node-management)
10. [CI/CD & DevOps](#cicd--devops)

---

## Cluster Administration

### Basic Cluster Information
- "show cluster version"
- "get cluster info"
- "list all namespaces"
- "show cluster nodes"
- "get cluster operators"
- "show cluster capacity"
- "get cluster resource quotas"
- "show api resources"
- "get cluster events"
- "show cluster certificates"

### Namespace Management
- "create namespace production"
- "delete namespace staging"
- "list all namespaces with labels"
- "show namespace quotas"
- "get namespace limits"
- "describe namespace monitoring"
- "set namespace labels"
- "show namespace events"

### Resource Quotas & Limits
- "show resource quotas in all namespaces"
- "create resource quota for namespace production"
- "get limit ranges"
- "show pod resource usage"
- "get resource consumption by namespace"
- "set resource limits for deployment"

### Cluster Upgrades & Maintenance
- "check cluster upgrade status"
- "show available cluster versions"
- "get cluster update history"
- "check node maintenance status"
- "show cluster health"
- "get cluster backup status"

---

## SRE Tasks

### Health Checks
- "show cluster health overview"
- "get all unhealthy pods"
- "show failing deployments"
- "list pods with high restart count"
- "get critical alerts"
- "show node health status"
- "check etcd health"
- "get api server status"

### Performance Monitoring
- "show top pods by CPU usage"
- "get memory usage by namespace"
- "show disk usage on nodes"
- "get network traffic statistics"
- "show pod resource consumption"
- "get node performance metrics"
- "show slowest responding pods"

### Incident Response
- "show crashing pods in the cluster"
- "get pods with ImagePullBackOff"
- "list pending pods"
- "show failed jobs"
- "get evicted pods"
- "show pods with OOMKilled status"
- "get recent error events"
- "show pods stuck in terminating state"

### Capacity Planning
- "show cluster resource utilization"
- "get node capacity and allocatable resources"
- "show pod density per node"
- "get resource requests vs limits"
- "show namespace resource consumption"
- "get cluster growth trends"

---

## Application Deployment

### Deployment Management
- "deploy nginx application"
- "scale deployment nginx to 5 replicas"
- "update deployment image to nginx:1.21"
- "rollback deployment to previous version"
- "pause deployment rollout"
- "resume deployment rollout"
- "show deployment status"
- "get deployment history"

### Helm Chart Management
- "install helm chart nginx from bitnami"
- "install prometheus in monitoring namespace"
- "install redis in redis namespace"
- "upgrade helm chart redis"
- "list all helm releases"
- "get helm status of nginx"
- "uninstall helm chart redis"
- "show helm chart values"
- "rollback helm release to previous version"

### Pod Management
- "get all pods in production namespace"
- "describe pod nginx-123 in default namespace"
- "get pod logs for app-456"
- "exec into pod nginx-789"
- "copy file from pod to local"
- "port-forward pod service to local port"
- "show pod resource usage"
- "get pod environment variables"

### Service Management
- "create service for deployment nginx"
- "expose deployment on port 80"
- "get all services in namespace"
- "show service endpoints"
- "test service connectivity"
- "get service DNS resolution"

---

## Networking

### Service Discovery
- "show all services in cluster"
- "get service endpoints"
- "test service connectivity"
- "show service DNS records"
- "get service mesh configuration"
- "show ingress controllers"

### Ingress & Routes
- "create ingress for service nginx"
- "get all ingress resources"
- "show route configuration"
- "test ingress connectivity"
- "get SSL certificate status"
- "show ingress annotations"
- "create route with custom domain"

### Network Policies
- "show network policies"
- "create network policy for namespace"
- "test network policy rules"
- "get network policy violations"
- "show allowed traffic flows"

### Load Balancing
- "show load balancer services"
- "get load balancer IP addresses"
- "test load balancer health"
- "show load balancer configuration"
- "get load balancer metrics"

### DNS & Service Mesh
- "show DNS configuration"
- "test DNS resolution"
- "get service mesh status"
- "show service mesh metrics"
- "get service mesh configuration"

### Network Troubleshooting & Packet Capture
- "tcpdump on pod my-app-123 in namespace production"
- "capture packets from pod nginx-456 interface eth0"
- "run tcpdump on pod frontend-789 port 8080"
- "packet capture for pod backend-321 host 10.0.0.1"
- "tcpdump traffic on pod my-service-555 in namespace staging"
- "capture network traffic from pod web-app-666 to wireshark"
- "run packet capture on pod api-server-777 for 60 seconds"
- "tcpdump on pod database-888 interface lo"
- "capture packets from pod cache-999 port 6379"
- "network capture for pod queue-111 host redis.example.com"

### Network Connectivity Testing
- "ping from pod my-app-123 to google.com"
- "test connectivity from pod nginx-456 to kubernetes.default.svc.cluster.local"
- "ping from pod frontend-789 to 8.8.8.8"
- "test network connectivity from pod backend-321 to database service"
- "ping from pod web-app-666 to external service"
- "test reachability from pod api-server-777 to upstream service"
- "check connectivity from pod cache-999 to redis cluster"
- "ping from pod queue-111 to message broker"

### DNS Resolution Testing
- "test DNS resolution from pod my-app-123"
- "nslookup from pod nginx-456 for kubernetes.default.svc.cluster.local"
- "check DNS from pod frontend-789 for external domain"
- "test DNS resolution from pod backend-321 for service discovery"
- "dig from pod web-app-666 for internal service"
- "check DNS configuration in pod api-server-777"
- "test DNS from pod cache-999 for redis service"
- "nslookup from pod queue-111 for message broker service"

### HTTP/HTTPS Testing
- "curl from pod my-app-123 to https://api.example.com"
- "test HTTP from pod nginx-456 to internal service"
- "curl from pod frontend-789 to backend service"
- "test HTTPS from pod backend-321 to external API"
- "http test from pod web-app-666 to upstream service"
- "curl from pod api-server-777 to kubernetes API"
- "test HTTP connectivity from pod cache-999 to redis web UI"
- "curl from pod queue-111 to message broker web interface"

### Network Statistics & Analysis
- "show network connections in pod my-app-123"
- "netstat from pod nginx-456"
- "show network interfaces in pod frontend-789"
- "get network statistics from pod backend-321"
- "show socket connections in pod web-app-666"
- "network analysis for pod api-server-777"
- "show network routes in pod cache-999"
- "get network configuration from pod queue-111"

### Advanced Network Debugging
- "debug network namespace for pod my-app-123"
- "show network interfaces in pod nginx-456 namespace"
- "get network routes from pod frontend-789 perspective"
- "debug pod backend-321 network configuration"
- "show iptables rules affecting pod web-app-666"
- "network troubleshooting for pod api-server-777"
- "debug network policy affecting pod cache-999"
- "show network bridges for pod queue-111"

---

## Storage Management

### Persistent Volumes
- "show all persistent volumes"
- "get persistent volume claims"
- "create persistent volume"
- "show storage classes"
- "get volume snapshots"
- "show volume usage"
- "test volume mounting"

### Storage Configuration
- "show storage class configuration"
- "get default storage class"
- "create storage class"
- "show volume binding modes"
- "get storage provisioner status"

### Backup & Recovery
- "show backup status"
- "create volume snapshot"
- "restore from snapshot"
- "get backup schedules"
- "show backup retention policies"

---

## Security & RBAC

### Authentication & Authorization
- "show cluster roles"
- "get role bindings"
- "create service account"
- "show user permissions"
- "get cluster admin users"
- "show RBAC configuration"
- "test user access"

### Security Contexts
- "show pod security policies"
- "get security context constraints"
- "show privileged containers"
- "get pod security standards"
- "show security violations"

### Secrets Management
- "show all secrets"
- "create secret for database"
- "get secret values"
- "show secret usage"
- "rotate secret keys"
- "get certificate expiration"

### Network Security
- "show network policies"
- "get security groups"
- "show firewall rules"
- "test network security"
- "get security scan results"

---

## Monitoring & Observability

### Metrics Collection
- "show prometheus metrics"
- "get grafana dashboards"
- "show application metrics"
- "get custom metrics"
- "show resource metrics"

### Logging
- "get logs for all pods in namespace"
- "show log aggregation status"
- "get audit logs"
- "show log retention policies"
- "search logs for error patterns"

### Alerting
- "show active alerts"
- "get alert manager configuration"
- "create alert rule"
- "test alert notifications"
- "show alert history"

### Tracing
- "show distributed tracing"
- "get trace data"
- "show jaeger configuration"
- "get performance traces"

---

## Troubleshooting

### Pod Issues
- "why is pod nginx-123 in crashloop?"
- "check why pod is stuck in pending state"
- "troubleshoot pod startup failure"
- "diagnose pod networking issues"
- "check pod resource constraints"
- "fix pod image pull errors"

### Deployment Issues
- "troubleshoot deployment failure"
- "check why deployment is not rolling out"
- "diagnose replica set issues"
- "fix deployment stuck in progress"
- "check deployment readiness probes"

### Network Issues
- "troubleshoot service connectivity"
- "check ingress not working"
- "diagnose DNS resolution problems"
- "fix network policy blocking traffic"
- "check load balancer issues"

### Storage Issues
- "troubleshoot persistent volume mounting"
- "check storage class not working"
- "diagnose volume provisioning failure"
- "fix storage permission issues"
- "check disk space issues"

### Performance Issues
- "troubleshoot slow pod startup"
- "check high CPU usage"
- "diagnose memory leaks"
- "fix network latency issues"
- "check disk I/O problems"

---

## Node Management

### Node Operations
- "show node status"
- "get node capacity"
- "drain node for maintenance"
- "uncordon node"
- "add node to cluster"
- "remove node from cluster"
- "show node labels"
- "set node labels"

### Node Troubleshooting
- "troubleshoot node not ready"
- "check node resource pressure"
- "diagnose node disk issues"
- "fix node network problems"
- "check node kubelet status"

### Node Scheduling
- "show node scheduling status"
- "get pod scheduling constraints"
- "show node affinity rules"
- "get node taints and tolerations"
- "test pod scheduling"

---

## CI/CD & DevOps

### Pipeline Management
- "show build configurations"
- "get build status"
- "start build pipeline"
- "show pipeline history"
- "get deployment pipeline status"

### Image Management
- "show image streams"
- "get image registry status"
- "push image to registry"
- "scan image for vulnerabilities"
- "get image metadata"

### GitOps
- "show gitops configuration"
- "get argocd applications"
- "sync application deployment"
- "show git webhook status"
- "get deployment diff"

### Automation
- "show scheduled jobs"
- "create cronjob"
- "get job execution history"
- "show automation workflows"
- "get operator status"

---

## Advanced Operations

### Disaster Recovery
- "show backup status"
- "create cluster backup"
- "restore from backup"
- "test disaster recovery"
- "get backup verification"

### Multi-Cluster Management
- "show cluster federation"
- "get multi-cluster status"
- "sync across clusters"
- "show cluster replication"
- "get cross-cluster networking"

### Compliance & Auditing
- "show compliance reports"
- "get audit trail"
- "check security compliance"
- "show policy violations"
- "get governance reports"

---

## Usage Examples

### Common Workflow Examples

**Morning Health Check:**
```
"show cluster health overview"
"get all unhealthy pods"
"show critical alerts"
"get node status"
"show recent error events"
```

**Application Deployment:**
```
"install helm chart nginx in production namespace"
"scale deployment nginx to 3 replicas"
"create ingress for service nginx"
"test service connectivity"
"get deployment status"
```

**Incident Response:**
```
"show crashing pods in the cluster"
"get pods with high restart count"
"troubleshoot pod startup failure"
"get logs for failed pod"
"check service connectivity"
```

**Capacity Planning:**
```
"show cluster resource utilization"
"get node capacity and allocatable resources"
"show top pods by CPU usage"
"get memory usage by namespace"
"show disk usage on nodes"
```

---

## Best Practices

1. **Be Specific**: Include namespace, resource names, and specific conditions when possible
2. **Use Natural Language**: The AI assistant understands conversational prompts
3. **Combine Operations**: You can ask for multiple related operations in one prompt
4. **Include Context**: Mention the environment (production, staging, etc.) when relevant
5. **Safety First**: The system will ask for confirmation on destructive operations

---

## Notes

- All prompts are designed to work with the OpenShift MCP Go assistant
- The system will automatically determine the appropriate kubectl, oc, or helm commands
- Commands are executed safely with built-in security validations
- Namespace creation is handled automatically when needed
- The assistant provides both command execution and diagnostic analysis

For more information on specific commands or troubleshooting, refer to the main documentation or ask the assistant directly.

---

## User-Generated Prompts

*These prompts were automatically collected from user interactions and usage patterns.*

### Application Deployment

#### Helm Charts
- "is gatekeeper operator installed" *(used 13 times, 95% confidence)*
- "is gatekeeper operator c installed" *(used 8 times, 30% confidence)*
- "install helm chart nginx in test namespace"
- "install skupper v2 in test namespace"

#### Pod Management
- "tcpdump on  httpd-676f79d94c-mzwdv pod in app1 namespace" *(used 4 times, 90% confidence)*
- "show me all pods in production namespace" *(used 2 times, 90% confidence)*
- "get all pods" *(used 2 times, 90% confidence)*
- "list all pods"
- "List all pods in the production namespace"

#### Service Management
- "create a namespace called new and create a service account called newsa" *(used 2 times, 30% confidence)*
- "create a service account called test-sa in test namespace"

### Cluster Administration

#### Basic Information
- "is gatekeeper operator installed on this cluster?" *(used 3 times, 95% confidence)*
- "apply this on the cluster https://skupper.io/v2/install.yaml"

#### Namespace Management
- "Troubleshoot the httpd pod in app1 namespace" *(used 9 times, 95% confidence)*
- "tcpdump on one of the httpd pod in app1 namespace" *(used 6 times, 90% confidence)*
- "tcpdump on httpd-676f79d94c-mzwdv pod in app1 namespace" *(used 4 times, 90% confidence)*
- "is pod httpd-676f79d94c-mzwdv running in app1 namespace"
- "create a namespace called test"
- "delete the namespace test"
- "make test-sa in test project an admin for that namespace"
- "can you make test-sa in test project an admin for that namespace"
- "status of the pod in httpd namespace"
- "status of the pod in rakesh namespace"
- "what is the state of the  pod in rakesh namespace"

### Node Management

#### Node Operations
- "Show cluster node status"

### Security

#### RBAC
- "create a test namespace and a service account called test-sa, this service accout should have admin access to test namespace" *(used 5 times, 90% confidence)*
- "create a namespace testing and create a service account test-sa in testing namespace" *(used 5 times, 30% confidence)*
- "create a namespace called test and create a service account test-sa in that namespace and that SA should have admin access to only that namespace" *(used 2 times, 90% confidence)*
- "create a servicce account called test-sa in test namespace and it should have admin access only to test namespace" *(used 2 times, 90% confidence)*
- "create a service account called test-sa in test namespace and it should have admin access only to test namespace" *(used 2 times, 90% confidence)*

### Troubleshooting

#### General
- "what can you do"
- "vscode terminal doesnt work"
- "how can i learn openshift"

#### Pod Issues
- "why is my nginx pod crashing in default namespace?"
- "why is httpd. pod stuck in container creating"
- "Troubleshoot ImagePullBackOff error"


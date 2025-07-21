package diagnostics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DiagnosticCollector handles collection of various diagnostic data
type DiagnosticCollector struct {
	logger     *logrus.Logger
	workingDir string
	timeout    time.Duration
}

// CollectionOptions defines options for diagnostic collection
type CollectionOptions struct {
	Namespace      string            `json:"namespace,omitempty"`
	PodName        string            `json:"pod_name,omitempty"`
	NodeName       string            `json:"node_name,omitempty"`
	OutputDir      string            `json:"output_dir,omitempty"`
	Duration       string            `json:"duration,omitempty"`
	Filters        map[string]string `json:"filters,omitempty"`
	IncludeLogs    bool              `json:"include_logs,omitempty"`
	IncludeMetrics bool              `json:"include_metrics,omitempty"`
	Compressed     bool              `json:"compressed,omitempty"`
}

// CollectionResult represents the result of a diagnostic collection
type CollectionResult struct {
	Type     string            `json:"type"`
	Status   string            `json:"status"`
	FilePath string            `json:"file_path,omitempty"`
	Size     int64             `json:"size,omitempty"`
	Duration time.Duration     `json:"duration"`
	Metadata map[string]string `json:"metadata,omitempty"`
	ErrorMsg string            `json:"error_msg,omitempty"`
	Summary  string            `json:"summary,omitempty"`
}

// NewDiagnosticCollector creates a new diagnostic collector
func NewDiagnosticCollector(logger *logrus.Logger, workingDir string) *DiagnosticCollector {
	if workingDir == "" {
		workingDir = "/tmp/diagnostics"
	}

	// Ensure working directory exists
	os.MkdirAll(workingDir, 0755)

	return &DiagnosticCollector{
		logger:     logger,
		workingDir: workingDir,
		timeout:    30 * time.Minute,
	}
}

// CollectMustGather collects OpenShift must-gather data
func (dc *DiagnosticCollector) CollectMustGather(ctx context.Context, opts *CollectionOptions) (*CollectionResult, error) {
	start := time.Now()
	result := &CollectionResult{
		Type:     "must-gather",
		Metadata: make(map[string]string),
	}

	// Set default image if not specified
	image := "registry.redhat.io/openshift4/ose-must-gather:latest"
	if opts.Filters != nil && opts.Filters["image"] != "" {
		image = opts.Filters["image"]
	}

	// Create output directory
	outputDir := filepath.Join(dc.workingDir, fmt.Sprintf("must-gather-%d", time.Now().Unix()))
	if opts.OutputDir != "" {
		outputDir = opts.OutputDir
	}
	os.MkdirAll(outputDir, 0755)

	// Build must-gather command
	args := []string{
		"adm", "must-gather",
		"--image=" + image,
		"--dest-dir=" + outputDir,
	}

	if opts.Namespace != "" {
		args = append(args, "--source-dir=/must-gather/"+opts.Namespace)
	}

	dc.logger.Infof("Starting must-gather collection with image: %s", image)

	cmd := exec.CommandContext(ctx, "oc", args...)
	output, err := cmd.CombinedOutput()

	result.Duration = time.Since(start)
	result.FilePath = outputDir
	result.Metadata["image"] = image
	result.Metadata["command"] = strings.Join(args, " ")

	if err != nil {
		result.Status = "failed"
		result.ErrorMsg = fmt.Sprintf("Must-gather failed: %v, output: %s", err, string(output))
		return result, err
	}

	// Get directory size
	if size, err := dc.getDirSize(outputDir); err == nil {
		result.Size = size
	}

	result.Status = "completed"
	result.Summary = fmt.Sprintf("Must-gather collected successfully in %s (%.2f MB)",
		outputDir, float64(result.Size)/(1024*1024))

	dc.logger.Infof("Must-gather collection completed: %s", result.Summary)
	return result, nil
}

// CollectSosReport collects sosreport from a specific node
func (dc *DiagnosticCollector) CollectSosReport(ctx context.Context, opts *CollectionOptions) (*CollectionResult, error) {
	start := time.Now()
	result := &CollectionResult{
		Type:     "sosreport",
		Metadata: make(map[string]string),
	}

	if opts.NodeName == "" {
		return nil, fmt.Errorf("node name is required for sosreport collection")
	}

	// Create output directory
	outputDir := filepath.Join(dc.workingDir, fmt.Sprintf("sosreport-%s-%d", opts.NodeName, time.Now().Unix()))
	os.MkdirAll(outputDir, 0755)

	// Create debug pod for sosreport collection
	debugPodYAML := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: sosreport-collector-%d
  namespace: default
spec:
  nodeName: %s
  hostNetwork: true
  hostPID: true
  hostIPC: true
  tolerations:
  - operator: Exists
  securityContext:
    privileged: true
  containers:
  - name: collector
    image: registry.redhat.io/rhel8/support-tools:latest
    command: ["/bin/bash", "-c"]
    args:
    - |
      chroot /host sosreport --batch --tmp-dir=/host/tmp/sosreport-%s --name=%s
      tar -czf /host/tmp/sosreport-%s.tar.gz -C /host/tmp sosreport-%s-*
      sleep infinity
    volumeMounts:
    - name: host
      mountPath: /host
  volumes:
  - name: host
    hostPath:
      path: /
  restartPolicy: Never
`, time.Now().Unix(), opts.NodeName, opts.NodeName, opts.NodeName, opts.NodeName, opts.NodeName)

	// Apply debug pod
	podFile := filepath.Join(outputDir, "sosreport-pod.yaml")
	if err := os.WriteFile(podFile, []byte(debugPodYAML), 0644); err != nil {
		result.Status = "failed"
		result.ErrorMsg = fmt.Sprintf("Failed to write pod YAML: %v", err)
		return result, err
	}

	dc.logger.Infof("Creating sosreport collection pod on node: %s", opts.NodeName)

	// Apply the pod
	cmd := exec.CommandContext(ctx, "oc", "apply", "-f", podFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		result.Status = "failed"
		result.ErrorMsg = fmt.Sprintf("Failed to create sosreport pod: %v, output: %s", err, string(output))
		return result, err
	}

	podName := fmt.Sprintf("sosreport-collector-%d", time.Now().Unix())

	// Wait for pod to be ready and collect sosreport
	// This is a simplified version - in production, you'd want better error handling and monitoring
	time.Sleep(60 * time.Second) // Wait for sosreport to complete

	// Copy sosreport from node
	copyCmd := exec.CommandContext(ctx, "oc", "cp",
		fmt.Sprintf("default/%s:/host/tmp/sosreport-%s.tar.gz", podName, opts.NodeName),
		filepath.Join(outputDir, fmt.Sprintf("sosreport-%s.tar.gz", opts.NodeName)))

	if output, err := copyCmd.CombinedOutput(); err != nil {
		dc.logger.Warnf("Failed to copy sosreport: %v, output: %s", err, string(output))
	}

	// Cleanup debug pod
	cleanupCmd := exec.CommandContext(ctx, "oc", "delete", "pod", podName)
	cleanupCmd.Run()

	result.Duration = time.Since(start)
	result.FilePath = outputDir
	result.Metadata["node"] = opts.NodeName
	result.Metadata["pod_name"] = podName

	if size, err := dc.getDirSize(outputDir); err == nil {
		result.Size = size
	}

	result.Status = "completed"
	result.Summary = fmt.Sprintf("Sosreport collected from node %s in %s", opts.NodeName, outputDir)

	dc.logger.Infof("Sosreport collection completed: %s", result.Summary)
	return result, nil
}

// CollectTcpdump performs network packet capture
func (dc *DiagnosticCollector) CollectTcpdump(ctx context.Context, opts *CollectionOptions) (*CollectionResult, error) {
	start := time.Now()
	result := &CollectionResult{
		Type:     "tcpdump",
		Metadata: make(map[string]string),
	}

	if opts.PodName == "" && opts.NodeName == "" {
		return nil, fmt.Errorf("either pod name or node name is required for tcpdump")
	}

	// Create output directory
	outputDir := filepath.Join(dc.workingDir, fmt.Sprintf("tcpdump-%d", time.Now().Unix()))
	os.MkdirAll(outputDir, 0755)

	duration := "60s"
	if opts.Duration != "" {
		duration = opts.Duration
	}

	var cmd *exec.Cmd
	var outputFile string

	if opts.PodName != "" {
		// Tcpdump on specific pod
		outputFile = filepath.Join(outputDir, fmt.Sprintf("tcpdump-%s.pcap", opts.PodName))

		// Create tcpdump command for pod
		tcpdumpArgs := []string{
			"exec", opts.PodName,
		}

		if opts.Namespace != "" {
			tcpdumpArgs = append(tcpdumpArgs, "-n", opts.Namespace)
		}

		tcpdumpArgs = append(tcpdumpArgs, "--", "timeout", duration, "tcpdump", "-i", "any", "-w", "/tmp/capture.pcap")

		// Add filters if specified
		if opts.Filters != nil {
			for key, value := range opts.Filters {
				if key == "filter" {
					tcpdumpArgs = append(tcpdumpArgs, strings.Split(value, " ")...)
				}
			}
		}

		cmd = exec.CommandContext(ctx, "oc", tcpdumpArgs...)

	} else {
		// Tcpdump on node using debug pod
		outputFile = filepath.Join(outputDir, fmt.Sprintf("tcpdump-%s.pcap", opts.NodeName))

		debugPodYAML := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: tcpdump-collector-%d
  namespace: default
spec:
  nodeName: %s
  hostNetwork: true
  tolerations:
  - operator: Exists
  securityContext:
    privileged: true
  containers:
  - name: tcpdump
    image: registry.redhat.io/rhel8/support-tools:latest
    command: ["/bin/bash", "-c"]
    args:
    - |
      timeout %s tcpdump -i any -w /tmp/capture.pcap
      sleep infinity
    securityContext:
      privileged: true
  restartPolicy: Never
`, time.Now().Unix(), opts.NodeName, duration)

		// Apply debug pod and collect
		podFile := filepath.Join(outputDir, "tcpdump-pod.yaml")
		if err := os.WriteFile(podFile, []byte(debugPodYAML), 0644); err != nil {
			result.Status = "failed"
			result.ErrorMsg = fmt.Sprintf("Failed to write tcpdump pod YAML: %v", err)
			return result, err
		}

		cmd = exec.CommandContext(ctx, "oc", "apply", "-f", podFile)
	}

	dc.logger.Infof("Starting tcpdump collection for %s", duration)

	output, err := cmd.CombinedOutput()

	result.Duration = time.Since(start)
	result.FilePath = outputFile
	result.Metadata["duration"] = duration
	result.Metadata["command"] = cmd.String()

	if err != nil {
		result.Status = "failed"
		result.ErrorMsg = fmt.Sprintf("Tcpdump failed: %v, output: %s", err, string(output))
		return result, err
	}

	if size, err := dc.getFileSize(outputFile); err == nil {
		result.Size = size
	}

	result.Status = "completed"
	result.Summary = fmt.Sprintf("Tcpdump completed, capture saved to %s (%.2f MB)",
		outputFile, float64(result.Size)/(1024*1024))

	dc.logger.Infof("Tcpdump collection completed: %s", result.Summary)
	return result, nil
}

// CollectLogs collects container and system logs
func (dc *DiagnosticCollector) CollectLogs(ctx context.Context, opts *CollectionOptions) (*CollectionResult, error) {
	start := time.Now()
	result := &CollectionResult{
		Type:     "logs",
		Metadata: make(map[string]string),
	}

	// Create output directory
	outputDir := filepath.Join(dc.workingDir, fmt.Sprintf("logs-%d", time.Now().Unix()))
	os.MkdirAll(outputDir, 0755)

	var files []string

	// Collect pod logs if pod specified
	if opts.PodName != "" {
		logFile := filepath.Join(outputDir, fmt.Sprintf("%s.log", opts.PodName))

		args := []string{"logs", opts.PodName}
		if opts.Namespace != "" {
			args = append(args, "-n", opts.Namespace)
		}
		args = append(args, "--previous=false", "--timestamps=true")

		cmd := exec.CommandContext(ctx, "oc", args...)
		output, err := cmd.Output()
		if err == nil {
			os.WriteFile(logFile, output, 0644)
			files = append(files, logFile)
		}

		// Also collect previous logs if available
		prevLogFile := filepath.Join(outputDir, fmt.Sprintf("%s-previous.log", opts.PodName))
		args[len(args)-2] = "--previous=true"
		cmd = exec.CommandContext(ctx, "oc", args...)
		if output, err := cmd.Output(); err == nil {
			os.WriteFile(prevLogFile, output, 0644)
			files = append(files, prevLogFile)
		}
	}

	// Collect events
	eventsFile := filepath.Join(outputDir, "events.yaml")
	args := []string{"get", "events", "-o", "yaml"}
	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	} else {
		args = append(args, "--all-namespaces")
	}

	cmd := exec.CommandContext(ctx, "oc", args...)
	if output, err := cmd.Output(); err == nil {
		os.WriteFile(eventsFile, output, 0644)
		files = append(files, eventsFile)
	}

	result.Duration = time.Since(start)
	result.FilePath = outputDir
	result.Metadata["files_collected"] = fmt.Sprintf("%d", len(files))

	if size, err := dc.getDirSize(outputDir); err == nil {
		result.Size = size
	}

	result.Status = "completed"
	result.Summary = fmt.Sprintf("Collected %d log files in %s", len(files), outputDir)

	dc.logger.Infof("Log collection completed: %s", result.Summary)
	return result, nil
}

// Helper functions
func (dc *DiagnosticCollector) getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func (dc *DiagnosticCollector) getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

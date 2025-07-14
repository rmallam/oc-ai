#!/bin/bash

# Quick test to verify the correct command is generated for crashing pods

echo "Testing Command Generation for Crashing Pods"
echo "============================================="

echo ""
echo "Expected command for 'show me crashing pods':"
echo "kubectl get pods --all-namespaces | grep -E \"(CrashLoopBackOff|ImagePullBackOff|Error|Evicted|OOMKilled)\""
echo ""

echo "Let's test this command directly on your cluster:"
echo ""

# Test the command directly
echo "Running: kubectl get pods --all-namespaces | grep -E \"(CrashLoopBackOff|ImagePullBackOff|Error|Evicted|OOMKilled)\""
kubectl get pods --all-namespaces | grep -E "(CrashLoopBackOff|ImagePullBackOff|Error|Evicted|OOMKilled)"

exit_code=$?

echo ""
if [ $exit_code -eq 0 ]; then
    echo "✓ Found some pods with issues above"
elif [ $exit_code -eq 1 ]; then
    echo "✓ No crashing pods found (this is good!)"
    echo "This means the command works, but there are no problematic pods in your cluster"
else
    echo "✗ Command failed with exit code $exit_code"
    echo "This might indicate an issue with kubectl access or cluster connectivity"
fi

echo ""
echo "If the OpenShift MCP Go application generates this same command and executes it,"
echo "you should get the same result as above - either the problematic pods or an empty result."

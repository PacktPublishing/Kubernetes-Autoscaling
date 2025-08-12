#!/bin/bash
set -e

NODE_NAME=$(kubectl get nodes -l intent=apps -o jsonpath='{.items[0].metadata.name}')

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: debug-kill-containerd
spec:
  hostPID: true
  nodeSelector:
    intent: apps
  tolerations:
  - operator: Exists
  containers:
  - name: killer
    image: ubuntu:20.04
    securityContext:
      privileged: true
      runAsUser: 0
    command:
    - sh
    - -c
    - |
      echo "Killing containerd process on host..."
      chroot /host pkill -9 -f containerd || true
      echo "Containerd killed. Node should become unhealthy."
      sleep 3600
    volumeMounts:
    - name: host-root
      mountPath: /host
  volumes:
  - name: host-root
    hostPath:
      path: /
  restartPolicy: Never
  nodeSelector:
    kubernetes.io/hostname: "$NODE_NAME"
EOF
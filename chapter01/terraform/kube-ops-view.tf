#---------------------------------------------------------------
# Kubernetes Manifests
#---------------------------------------------------------------

resource "kubectl_manifest" "kube_ops_view_deployment" {
  yaml_body = <<-YAML
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      labels:
        application: kube-ops-view
        component: frontend
      name: kube-ops-view
    spec:
      replicas: 1
      selector:
        matchLabels:
          application: kube-ops-view
          component: frontend
      template:
        metadata:
          labels:
            application: kube-ops-view
            component: frontend
        spec:
          nodeSelector:
            intent: control-apps
          serviceAccountName: kube-ops-view
          containers:
          - name: service
            image: hjacobs/kube-ops-view:20.4.0
            ports:
            - containerPort: 8080
              protocol: TCP
            readinessProbe:
              httpGet:
                path: /health
                port: 8080
              initialDelaySeconds: 5
              timeoutSeconds: 1
            livenessProbe:
              httpGet:
                path: /health
                port: 8080
              initialDelaySeconds: 30
              periodSeconds: 30
              timeoutSeconds: 10
              failureThreshold: 5
            resources:
              limits:
                cpu: 400m
                memory: 400Mi
              requests:
                cpu: 400m
                memory: 400Mi
            securityContext:
              readOnlyRootFilesystem: true
              runAsNonRoot: true
              runAsUser: 1000
  YAML

  depends_on = [
    module.eks
  ]
}

resource "kubectl_manifest" "kube_ops_view_sa" {
  yaml_body = <<-YAML
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: kube-ops-view
  YAML

  depends_on = [
    module.eks
  ]
}

resource "kubectl_manifest" "kube_ops_view_clusterrole" {
  yaml_body = <<-YAML
    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: kube-ops-view
    rules:
    - apiGroups: [""]
      resources: ["nodes", "pods"]
      verbs:
        - list
    - apiGroups: ["metrics.k8s.io"]
      resources: ["nodes", "pods"]
      verbs:
        - get
        - list
  YAML

  depends_on = [
    module.eks
  ]
}

resource "kubectl_manifest" "kube_ops_view_clusterrole_binding" {
  yaml_body = <<-YAML
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: kube-ops-view
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: kube-ops-view
    subjects:
    - kind: ServiceAccount
      name: kube-ops-view
      namespace: default
  YAML

  depends_on = [
    module.eks
  ]
}

resource "kubectl_manifest" "kube_ops_view_service" {
  yaml_body = <<-YAML
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        application: kube-ops-view
        component: frontend
      name: kube-ops-view
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: external
        service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
        service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing
    spec:
      selector:
        application: kube-ops-view
        component: frontend
      type: LoadBalancer
      ports:
      - port: 80
        protocol: TCP
        targetPort: 8080
  YAML

  depends_on = [
    module.eks
  ]
}
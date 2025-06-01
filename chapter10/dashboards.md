# Deploy the Karpenter Grafana dashboards

```
curl -s https://karpenter.sh/preview/getting-started/getting-started-with-karpenter/karpenter-capacity-dashboard.json -o karpenter-capacity-dashboard.json

kubectl create configmap karpenter-capacity-dashboard \
  --from-file=karpenter-capacity-dashboard.json \
  --namespace monitoring

kubectl label configmap karpenter-capacity-dashboard \
  grafana_dashboard=1 \
  -n monitoring
```

```
curl -s https://karpenter.sh/preview/getting-started/getting-started-with-karpenter/karpenter-performance-dashboard.json -o karpenter-performance-dashboard.json

kubectl create configmap karpenter-performance-dashboard \
  --from-file=karpenter-performance-dashboard.json \
  --namespace monitoring\

kubectl label configmap karpenter-performance-dashboard \
  grafana_dashboard=1 \
  -n monitoring
```
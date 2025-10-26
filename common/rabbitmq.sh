#!/bin/bash

ACTION=$1
NODE_SELECTOR=$2

if [ "$ACTION" == "uninstall" ]; then
  echo "Uninstalling RabbitMQ..."
  kubectl delete rabbitmqcluster rabbitmq
  echo "RabbitMQ uninstalled!"
  exit 0
fi

if [ "$ACTION" != "install" ]; then
  echo "Usage: $0 [install|uninstall] [nodeSelector]"
  echo "Example: $0 install intent=apps"
  exit 1
fi

# Install RabbitMQ Cluster Operator
kubectl apply -f "https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml"

# Wait for operator to be ready
sleep 30

# Create credentials secret
kubectl create secret generic rabbitmq-default-user \
  --from-literal=username=user \
  --from-literal=password=autoscaling \
  --from-literal=default_user.conf="default_user = user
default_pass = autoscaling" \
  --dry-run=client -o yaml | kubectl apply -f -

# Deploy RabbitMQ cluster
if [ -z "$NODE_SELECTOR" ]; then
  # No nodeSelector
  cat <<EOF | kubectl apply -f -
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: rabbitmq
spec:
  replicas: 1
EOF
else
  # With nodeSelector
  KEY="${NODE_SELECTOR%%=*}"
  VALUE="${NODE_SELECTOR##*=}"
  cat <<EOF | kubectl apply -f -
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: rabbitmq
spec:
  replicas: 1
  override:
    statefulSet:
      spec:
        template:
          spec:
            nodeSelector:
              ${KEY}: "${VALUE}"
EOF
fi

# Wait for RabbitMQ to be ready
echo "Waiting for RabbitMQ to be ready..."
kubectl wait --for=condition=ReconcileSuccess rabbitmqcluster/rabbitmq --timeout=300s

echo "RabbitMQ is ready!"
echo "Username: user"
echo "Password: autoscaling"
echo ""
echo "Access management UI:"
echo "kubectl port-forward service/rabbitmq 15672:15672"

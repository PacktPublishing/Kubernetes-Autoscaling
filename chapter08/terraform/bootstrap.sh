#!/bin/bash

echo "Creating the Amazon EKS cluster ... "

helm registry logout public.ecr.aws
export TF_VAR_region=eu-west-1 # Change it to $AWS_REGION if you want to use the one provided from the environment

# Update Bottlerocket version to the latest
echo "Fetching the latest Bottlerocket version..."
LATEST_BOTTLEROCKET=$(aws ssm get-parameter \
  --name /aws/service/bottlerocket/aws-k8s-1.34/x86_64/latest/image_version \
  --region ${TF_VAR_region} \
  --query 'Parameter.Value' \
  --output text 2>/dev/null)

if [ -z "$LATEST_BOTTLEROCKET" ]; then
  echo "Warning: Could not fetch latest Bottlerocket version. Using existing version in files."
else
  echo "Latest Bottlerocket version: $LATEST_BOTTLEROCKET"
  echo "Updating nodepool.yaml files with latest Bottlerocket version..."
  
  # Update chapter08 nodepool.yaml
  if [ -f "../nodepool.yaml" ]; then
    sed -i "s/bottlerocket@[0-9.]\+/bottlerocket@${LATEST_BOTTLEROCKET}/g" ../nodepool.yaml
    echo "  ✓ Updated ../nodepool.yaml"
  fi
  
  # Update chapter09 files
  if [ -f "../../chapter09/nodepool.yaml" ]; then
    sed -i "s/bottlerocket@[0-9.]\+/bottlerocket@${LATEST_BOTTLEROCKET}/g" ../../chapter09/nodepool.yaml
    echo "  ✓ Updated ../../chapter09/nodepool.yaml"
  fi
  
  if [ -f "../../chapter09/nodepool.disruption.yaml" ]; then
    sed -i "s/bottlerocket@[0-9.]\+/bottlerocket@${LATEST_BOTTLEROCKET}/g" ../../chapter09/nodepool.disruption.yaml
    echo "  ✓ Updated ../../chapter09/nodepool.disruption.yaml"
  fi
  
  # Update chapter11 GPU nodeclass
  if [ -f "../../chapter11/gpu/nodeclass.yaml" ]; then
    sed -i "s/bottlerocket@v\?[0-9.]\+/bottlerocket@${LATEST_BOTTLEROCKET}/g" ../../chapter11/gpu/nodeclass.yaml
    echo "  ✓ Updated ../../chapter11/gpu/nodeclass.yaml"
  fi
  
  echo "Bottlerocket version update complete!"
fi

terraform init

# Retry logic for terraform apply
MAX_RETRIES=3
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  echo "Attempt $((RETRY_COUNT + 1)) of $MAX_RETRIES..."
  terraform apply --auto-approve
  
  if [ $? -eq 0 ]; then
    echo "Terraform apply succeeded!"
    break
  else
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
      echo "Terraform apply failed. Retrying..."
      sleep 5
    else
      echo "Terraform apply failed after $MAX_RETRIES attempts."
      exit 1
    fi
  fi
done

echo "Done! You should be able to connect to the cluster now."

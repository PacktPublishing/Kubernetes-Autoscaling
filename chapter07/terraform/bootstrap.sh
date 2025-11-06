#!/bin/bash

echo "Creating the Amazon EKS cluster ... "

helm registry logout public.ecr.aws
export TF_VAR_region=eu-west-1 # Change it to $AWS_REGION if you want to use the one provided from the environment
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
#!/bin/bash

echo "Destroying the Amazon EKS cluster ... "

kubectl delete --all svc
kubectl delete --all nodeclaim
kubectl delete --all nodepool
kubectl delete --all ec2nodeclass
export TF_VAR_region=eu-west-1 # Change it to $AWS_REGION if you want to use the one provided from the environment

# Retry logic for terraform destroy
MAX_RETRIES=3
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  echo "Attempt $((RETRY_COUNT + 1)) of $MAX_RETRIES..."
  terraform destroy --auto-approve
  
  if [ $? -eq 0 ]; then
    echo "Terraform destroy succeeded!"
    break
  else
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
      echo "Terraform destroy failed. Retrying..."
      sleep 5
    else
      echo "Terraform destroy failed after $MAX_RETRIES attempts."
      exit 1
    fi
  fi
done

echo "Done!"
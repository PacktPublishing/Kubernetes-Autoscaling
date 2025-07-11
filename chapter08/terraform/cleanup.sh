#!/bin/bash

echo "Destroying the Amazon EKS cluster ... "

kubectl delete --all svc
kubectl delete --all nodeclaim
kubectl delete --all nodepool
kubectl delete --all ec2nodeclass
export TF_VAR_region=eu-west-1 # Change it to $AWS_REGION if you want to use the one provided from the environment
terraform destroy -target="module.eks_blueprints_addons" --auto-approve
terraform destroy -target="module.eks" --auto-approve
terraform destroy --auto-approve

echo "Done!"
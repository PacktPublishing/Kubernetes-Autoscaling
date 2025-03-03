#!/bin/bash

echo "Destroying the Amazon EKS cluster ... "

kubectl delete svc --all

export TF_VAR_region=$AWS_REGION
terraform destroy -target="module.eks_blueprints_addons" --auto-approve
terraform destroy -target="module.eks" --auto-approve
terraform destroy --auto-approve

echo "Done!"
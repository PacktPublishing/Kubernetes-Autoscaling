#!/bin/bash

echo "Creating the Amazon EKS cluster ... "

helm registry logout public.ecr.aws
export TF_VAR_region=$AWS_REGION
terraform init
terraform apply -target="module.vpc" -auto-approve
terraform apply -target="module.eks" -auto-approve
terraform apply --auto-approve

echo "Done! You should be able to connect to the cluster now."
#!/bin/bash

echo "Creating the Amazon EKS cluster ... "

helm registry logout public.ecr.aws
export TF_VAR_region=eu-west-1 # Change it to $AWS_REGION if you want to use the one provided from the environment
terraform init
terraform apply -target="module.vpc" -auto-approve
terraform apply -target="module.eks" -auto-approve
terraform apply --auto-approve

echo "Done! You should be able to connect to the cluster now."

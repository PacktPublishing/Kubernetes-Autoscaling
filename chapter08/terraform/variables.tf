## NOTE: It's going to use your AWS_REGION or AWS_DEFAULT_REGION environment variable,
## but you can define which on to use in terraform.tfvars file as well, or pass it as an argument
## in the CLI like this "terraform apply -var 'region=eu-west-1'"
variable "region" {
  description = "Region to deploy the resources"
  type        = string
}

variable "eks_version" {
  description = "Amazon EKS version to use"
  type        = string
  default     = "1.34"
}

variable "karpenter_version" {
  description = "Karpenter version to install"
  type        = string
  default     = "1.8.2"
}

variable "region" {
    description = "The AWS region to create the EKS cluster in."
    default     = "us-west-2"
}
  
variable "repository_name" {
  description = "The name of the ECR repository."
  default     = "my-docker-registry"
}
    
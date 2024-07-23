output "cluster_endpoint" {
    description = "The endpoint for your EKS cluster."
    value       = aws_eks_cluster.eks.endpoint
  }
  
  output "cluster_name" {
    description = "The name of the EKS cluster."
    value       = aws_eks_cluster.eks.name
  }
  
  output "kubeconfig" {
    description = "Kubeconfig file to access the EKS cluster."
    value       = <<EOT
  apiVersion: v1
  clusters:
  - cluster:
      server: ${aws_eks_cluster.eks.endpoint}
      certificate-authority-data: ${base64decode(aws_eks_cluster.eks.certificate_authority[0].data)}
    name: eks-cluster
  contexts:
  - context:
      cluster: eks-cluster
      user: aws
    name: eks-context
  current-context: eks-context
  kind: Config
  users:
  - name: aws
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1alpha1
        command: aws
        args:
          - eks
          - get-token
          - --cluster-name
          - ${aws_eks_cluster.eks.name}
  EOT
  }
  
  output "repository_url" {
    description = "The URL of the created ECR repository."
    value       = aws_ecr_repository.main.repository_url
  }
  
  
  
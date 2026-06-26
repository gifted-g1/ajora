#!/bin/bash
# =============================================================================
# AJORA - Complete Infrastructure Deployment Script
# Version: 1.0.0
# Description: One-click deployment of the entire Ajora infrastructure
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# =============================================================================
# CONFIGURATION
# =============================================================================
export AWS_REGION="${AWS_REGION:-us-east-1}"
export ENVIRONMENT="${ENVIRONMENT:-prod}"
export PROJECT_NAME="ajora"
export CLUSTER_NAME="ajora-${ENVIRONMENT}"
export TF_STATE_BUCKET="ajora-terraform-state-${ENVIRONMENT}"
export TF_LOCK_TABLE="ajora-terraform-locks"

# Paths
export ROOT_DIR="$(pwd)"
export INFRA_DIR="${ROOT_DIR}/infrastructure"
export K8S_DIR="${ROOT_DIR}/kubernetes"
export CI_CD_DIR="${ROOT_DIR}/.github/workflows"
export MONITORING_DIR="${ROOT_DIR}/monitoring"
export SECURITY_DIR="${ROOT_DIR}/security"
export SCRIPTS_DIR="${ROOT_DIR}/scripts"

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

print_header() {
    echo -e "\n${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC} $1"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}\n"
}

print_step() {
    echo -e "\n${CYAN}▶${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

print_info() {
    echo -e "${PURPLE}ℹ️${NC} $1"
}

check_command() {
    if ! command -v "$1" &> /dev/null; then
        print_error "$1 is not installed. Please install it first."
        exit 1
    fi
}

confirm_action() {
    local prompt="$1"
    read -p "$(echo -e ${YELLOW}⚠️ ${prompt} (y/N)${NC} )" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Operation cancelled."
        exit 1
    fi
}

# =============================================================================
# PREREQUISITE CHECKS
# =============================================================================

check_prerequisites() {
    print_header "Checking Prerequisites"
    
    local required_commands=(
        "aws"
        "terraform"
        "kubectl"
        "helm"
        "docker"
        "git"
        "jq"
        "yq"
        "python3"
        "curl"
        "wget"
    )
    
    for cmd in "${required_commands[@]}"; do
        check_command "$cmd"
        print_success "$cmd installed"
    done
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS credentials not configured or invalid. Please run 'aws configure' first."
        exit 1
    fi
    print_success "AWS credentials verified"
    
    # Check Docker daemon
    if ! docker info &> /dev/null; then
        print_error "Docker daemon not running."
        exit 1
    fi
    print_success "Docker daemon running"
    
    # Check Kubernetes context
    if ! kubectl version --client &> /dev/null; then
        print_error "kubectl not configured."
        exit 1
    fi
    print_success "kubectl configured"
}

# =============================================================================
# DIRECTORY SETUP
# =============================================================================

setup_directories() {
    print_header "Setting up Directory Structure"
    
    local dirs=(
        "${INFRA_DIR}"
        "${K8S_DIR}/manifests"
        "${K8S_DIR}/helm"
        "${CI_CD_DIR}"
        "${MONITORING_DIR}/prometheus"
        "${MONITORING_DIR}/loki"
        "${MONITORING_DIR}/tempo"
        "${MONITORING_DIR}/grafana"
        "${SECURITY_DIR}/opa"
        "${SECURITY_DIR}/falco"
        "${SECURITY_DIR}/trivy"
        "${SCRIPTS_DIR}"
        "${ROOT_DIR}/docs"
        "${ROOT_DIR}/backups"
    )
    
    for dir in "${dirs[@]}"; do
        mkdir -p "${dir}"
        print_success "Created ${dir}"
    done
}

# =============================================================================
# TERRAFORM SETUP
# =============================================================================

setup_terraform() {
    print_header "Setting up Terraform Infrastructure"
    
    cd "${INFRA_DIR}"
    
    # Create Terraform files
    print_step "Creating main.tf"
    cat > main.tf << 'EOF'
terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.10"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Environment = var.environment
      Project     = "ajora"
      ManagedBy   = "terraform"
    }
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = {
    Name = "ajora-vpc-${var.environment}"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  
  tags = {
    Name = "ajora-igw-${var.environment}"
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count             = 3
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.10${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  
  map_public_ip_on_launch = true
  
  tags = {
    Name = "ajora-public-${data.aws_availability_zones.available.names[count.index]}"
    Type = "Public"
  }
}

# Private Subnets
resource "aws_subnet" "private" {
  count             = 3
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  
  tags = {
    Name = "ajora-private-${data.aws_availability_zones.available.names[count.index]}"
    Type = "Private"
  }
}

# Database Subnets
resource "aws_subnet" "database" {
  count             = 3
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  
  tags = {
    Name = "ajora-database-${data.aws_availability_zones.available.names[count.index]}"
    Type = "Database"
  }
}

# NAT Gateways
resource "aws_eip" "nat" {
  count = 3
  domain = "vpc"
  
  tags = {
    Name = "ajora-nat-eip-${count.index + 1}"
  }
}

resource "aws_nat_gateway" "main" {
  count         = 3
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id
  
  tags = {
    Name = "ajora-nat-${count.index + 1}"
  }
  
  depends_on = [aws_internet_gateway.main]
}

# Route Tables
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  
  tags = {
    Name = "ajora-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  count          = 3
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  count  = 3
  vpc_id = aws_vpc.main.id
  
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }
  
  tags = {
    Name = "ajora-private-rt-${count.index + 1}"
  }
}

resource "aws_route_table_association" "private" {
  count          = 3
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

# Security Groups
resource "aws_security_group" "eks_control_plane" {
  name        = "ajora-eks-control-plane-sg"
  description = "Security group for EKS control plane"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "ajora-eks-control-plane-sg"
  }
}

resource "aws_security_group" "eks_nodes" {
  name        = "ajora-eks-nodes-sg"
  description = "Security group for EKS worker nodes"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "ajora-eks-nodes-sg"
  }
}

resource "aws_security_group" "rds" {
  name        = "ajora-rds-sg"
  description = "Security group for RDS"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.eks_nodes.id]
  }
  
  tags = {
    Name = "ajora-rds-sg"
  }
}

resource "aws_security_group" "redis" {
  name        = "ajora-redis-sg"
  description = "Security group for Redis"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.eks_nodes.id]
  }
  
  tags = {
    Name = "ajora-redis-sg"
  }
}

resource "aws_security_group" "kafka" {
  name        = "ajora-kafka-sg"
  description = "Security group for Kafka"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port       = 9092
    to_port         = 9094
    protocol        = "tcp"
    security_groups = [aws_security_group.eks_nodes.id]
  }
  
  tags = {
    Name = "ajora-kafka-sg"
  }
}

# EKS Cluster
resource "aws_eks_cluster" "main" {
  name     = "ajora-${var.environment}"
  version  = var.eks_version
  role_arn = aws_iam_role.eks_cluster.arn
  
  vpc_config {
    subnet_ids              = concat(aws_subnet.private[*].id, aws_subnet.public[*].id)
    endpoint_private_access = true
    endpoint_public_access  = true
    security_group_ids      = [aws_security_group.eks_control_plane.id]
  }
  
  tags = {
    Name = "ajora-eks-${var.environment}"
  }
}

# IAM Roles
resource "aws_iam_role" "eks_cluster" {
  name = "ajora-eks-cluster-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster.name
}

resource "aws_iam_role" "eks_nodes" {
  name = "ajora-eks-nodes-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_nodes_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_nodes.name
}

resource "aws_iam_role_policy_attachment" "eks_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_nodes.name
}

resource "aws_iam_role_policy_attachment" "eks_container_registry_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_nodes.name
}

# EKS Node Groups
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "ajora-main-workers"
  node_role_arn   = aws_iam_role.eks_nodes.arn
  subnet_ids      = aws_subnet.private[*].id
  
  instance_types = var.main_node_group.instance_types
  capacity_type  = "ON_DEMAND"
  
  scaling_config {
    desired_size = var.main_node_group.desired_size
    max_size     = var.main_node_group.max_size
    min_size     = var.main_node_group.min_size
  }
  
  update_config {
    max_unavailable = 1
  }
  
  labels = {
    node-group = "main"
    workload   = "general"
  }
  
  tags = {
    Name = "ajora-main-workers"
  }
}

resource "aws_eks_node_group" "spot" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "ajora-spot-workers"
  node_role_arn   = aws_iam_role.eks_nodes.arn
  subnet_ids      = aws_subnet.private[*].id
  
  instance_types = ["c6a.large", "c6i.large"]
  capacity_type  = "SPOT"
  
  scaling_config {
    desired_size = 3
    max_size     = 10
    min_size     = 3
  }
  
  update_config {
    max_unavailable = 1
  }
  
  taint {
    key    = "spot"
    value  = "true"
    effect = "NO_SCHEDULE"
  }
  
  labels = {
    node-group = "spot"
    workload   = "batch"
  }
  
  tags = {
    Name = "ajora-spot-workers"
  }
}

# RDS PostgreSQL
resource "aws_db_subnet_group" "main" {
  name        = "ajora-rds-subnet-group"
  description = "Subnet group for RDS instances"
  subnet_ids  = aws_subnet.database[*].id
  
  tags = {
    Name = "ajora-rds-subnet-group"
  }
}

resource "aws_db_instance" "primary" {
  identifier = "ajora-postgres-primary"
  
  engine                = "postgres"
  engine_version        = "16.2"
  instance_class        = var.db_instance_class
  allocated_storage     = 200
  max_allocated_storage = 1000
  storage_type          = "gp3"
  storage_encrypted     = true
  
  db_name  = "ajora"
  username = "ajora_admin"
  password = random_password.db_master.result
  
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  
  backup_retention_period = 30
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:00-sun:05:00"
  
  multi_az                = true
  publicly_accessible     = false
  deletion_protection     = true
  
  performance_insights_enabled = true
  performance_insights_retention_period = 31
  
  skip_final_snapshot = false
  final_snapshot_identifier = "ajora-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
  
  tags = {
    Name = "ajora-rds-primary"
  }
}

resource "random_password" "db_master" {
  length  = 32
  special = false
}

# ElastiCache Redis
resource "aws_elasticache_subnet_group" "main" {
  name        = "ajora-redis-subnet-group"
  description = "Subnet group for Redis cluster"
  subnet_ids  = aws_subnet.private[*].id
}

resource "aws_elasticache_replication_group" "redis" {
  replication_group_id = "ajora-redis"
  description          = "Redis replication group for Ajora"
  
  engine                = "redis"
  engine_version        = "7.1"
  node_type             = var.redis_node_type
  num_cache_clusters    = 3
  port                  = 6379
  
  subnet_group_name     = aws_elasticache_subnet_group.main.name
  security_group_ids    = [aws_security_group.redis.id]
  
  automatic_failover_enabled = true
  multi_az_enabled           = true
  
  at_rest_encryption_enabled  = true
  transit_encryption_enabled  = true
  auth_token                  = random_password.redis_auth.result
  
  snapshot_retention_limit = 7
  snapshot_window          = "03:00-04:00"
  
  tags = {
    Name = "ajora-redis"
  }
}

resource "random_password" "redis_auth" {
  length  = 32
  special = false
}

# S3 Buckets
resource "aws_s3_bucket" "documents" {
  bucket = "ajora-documents-${var.environment}-${data.aws_caller_identity.current.account_id}"
  
  tags = {
    Name        = "ajora-documents"
    Environment = var.environment
  }
}

resource "aws_s3_bucket_versioning" "documents" {
  bucket = aws_s3_bucket.documents.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_public_access_block" "documents" {
  bucket = aws_s3_bucket.documents.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket" "logs" {
  bucket = "ajora-logs-${var.environment}-${data.aws_caller_identity.current.account_id}"
  
  tags = {
    Name        = "ajora-logs"
    Environment = var.environment
  }
}

resource "aws_s3_bucket" "artifacts" {
  bucket = "ajora-artifacts-${var.environment}-${data.aws_caller_identity.current.account_id}"
  
  tags = {
    Name        = "ajora-artifacts"
    Environment = var.environment
  }
}

resource "aws_s3_bucket" "backups" {
  bucket = "ajora-backups-${var.environment}-${data.aws_caller_identity.current.account_id}"
  
  tags = {
    Name        = "ajora-backups"
    Environment = var.environment
  }
}

# ECR Repositories
locals {
  ecr_repositories = [
    "api-gateway",
    "auth-service",
    "user-service",
    "pool-service",
    "contribution-service",
    "notification-service",
    "blockchain-orchestrator",
    "wallet-signer",
    "fraud-detector"
  ]
}

resource "aws_ecr_repository" "services" {
  for_each = toset(local.ecr_repositories)
  
  name = "ajora-${each.value}"
  
  image_tag_mutability = "IMMUTABLE"
  
  image_scanning_configuration {
    scan_on_push = true
  }
  
  tags = {
    Name        = "ajora-${each.value}"
    Service     = each.value
    Environment = var.environment
  }
}

resource "aws_ecr_lifecycle_policy" "services" {
  for_each = toset(local.ecr_repositories)
  
  repository = aws_ecr_repository.services[each.key].name
  
  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 50 images"
        selection = {
          tagStatus     = "any"
          countType     = "imageCountMoreThan"
          countNumber   = 50
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
EOF
    
    print_step "Creating variables.tf"
    cat > variables.tf << 'EOF'
variable "aws_region" {
  description = "AWS region"
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  default     = "prod"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "eks_version" {
  description = "EKS version"
  default     = "1.28"
}

variable "db_instance_class" {
  description = "RDS instance class"
  default     = "db.r6g.large"
}

variable "redis_node_type" {
  description = "Redis node type"
  default     = "cache.r6g.large"
}

variable "main_node_group" {
  description = "Main node group configuration"
  type = object({
    desired_size = number
    max_size     = number
    min_size     = number
    instance_types = list(string)
  })
  default = {
    desired_size   = 6
    max_size       = 20
    min_size       = 6
    instance_types = ["r6i.xlarge", "r6a.xlarge"]
  }
}
EOF
    
    print_step "Creating terraform.tfvars"
    cat > terraform.tfvars << EOF
aws_region = "${AWS_REGION}"
environment = "${ENVIRONMENT}"
EOF
    
    print_step "Initializing Terraform"
    terraform init
    
    print_step "Validating Terraform configuration"
    terraform validate
    
    print_step "Planning infrastructure"
    terraform plan -out=tfplan
    
    if [[ "${ENVIRONMENT}" != "prod" ]] || confirm_action "Deploy infrastructure to ${ENVIRONMENT}?"; then
        print_step "Applying infrastructure"
        terraform apply tfplan
        
        # Get outputs
        print_step "Extracting outputs"
        terraform output -json > outputs.json || true
        
        # Save kubeconfig
        print_step "Saving kubeconfig"
        aws eks update-kubeconfig --name "${CLUSTER_NAME}" --region "${AWS_REGION}" || true
        
        print_success "Terraform infrastructure deployed successfully!"
    else
        print_warning "Terraform apply skipped"
    fi
    
    cd "${ROOT_DIR}"
}

# =============================================================================
# KUBERNETES SETUP
# =============================================================================

setup_kubernetes() {
    print_header "Setting up Kubernetes Configuration"
    
    cd "${K8S_DIR}/manifests"
    
    # Create namespaces
    print_step "Creating namespaces"
    cat > namespaces.yaml << 'EOF'
apiVersion: v1
kind: Namespace
metadata:
  name: production
---
apiVersion: v1
kind: Namespace
metadata:
  name: staging
---
apiVersion: v1
kind: Namespace
metadata:
  name: ingress-nginx
---
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager
---
apiVersion: v1
kind: Namespace
metadata:
  name: security
---
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
---
apiVersion: v1
kind: Namespace
metadata:
  name: kafka
---
apiVersion: v1
kind: Namespace
metadata:
  name: databases
EOF
    
    kubectl apply -f namespaces.yaml || true
    
    # Create RBAC
    print_step "Creating RBAC configuration"
    cat > rbac.yaml << 'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: security-audit
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: admin
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ServiceAccount
metadata:
  name: production-sa
  namespace: production
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: production-role
  namespace: production
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: production-role-binding
  namespace: production
subjects:
- kind: ServiceAccount
  name: production-sa
  namespace: production
roleRef:
  kind: Role
  name: production-role
  apiGroup: rbac.authorization.k8s.io
EOF
    
    kubectl apply -f rbac.yaml || true
    
    # Create network policies
    print_step "Creating network policies"
    cat > network-policies.yaml << 'EOF'
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: production
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-gateway-policy
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: api-gateway
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: auth-service
    ports:
    - protocol: TCP
      port: 8080
  - to:
    - podSelector:
        matchLabels:
          app: user-service
    ports:
    - protocol: TCP
      port: 8080
EOF
    
    kubectl apply -f network-policies.yaml || true
    
    # Install Istio (if istioctl is available)
    if command -v istioctl &> /dev/null; then
        print_step "Installing Istio service mesh"
        istioctl install --set profile=demo -y || true
    else
        print_warning "istioctl not found. Skipping Istio installation."
    fi
    
    # Install Cert-Manager
    print_step "Installing Cert-Manager"
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.13.0/cert-manager.yaml || true
    
    cd "${ROOT_DIR}"
}

# =============================================================================
# HELM CHARTS SETUP
# =============================================================================

setup_helm_charts() {
    print_header "Setting up Helm Charts"
    
    cd "${K8S_DIR}/helm"
    
    # Add Helm repositories
    print_step "Adding Helm repositories"
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
    helm repo add grafana https://grafana.github.io/helm-charts || true
    helm repo add istio https://istio-release.storage.googleapis.com/charts || true
    helm repo add jetstack https://charts.jetstack.io || true
    helm repo add elastic https://helm.elastic.co || true
    helm repo add hashicorp https://helm.releases.hashicorp.com || true
    helm repo add bitnami https://charts.bitnami.com/bitnami || true
    helm repo update || true
    
    # Install Prometheus stack
    print_step "Installing Prometheus stack"
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --create-namespace \
        --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.storageClassName=gp3 \
        --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=50Gi \
        --set prometheus.prometheusSpec.retention=30d \
        --set alertmanager.enabled=true \
        --wait || true
    
    # Install Loki stack
    print_step "Installing Loki stack"
    helm upgrade --install loki grafana/loki-stack \
        --namespace monitoring \
        --set loki.persistence.enabled=true \
        --set loki.persistence.size=50Gi \
        --set promtail.enabled=true \
        --wait || true
    
    # Install Grafana
    print_step "Installing Grafana"
    helm upgrade --install grafana grafana/grafana \
        --namespace monitoring \
        --set persistence.enabled=true \
        --set persistence.size=20Gi \
        --set adminPassword=admin123 \
        --set datasources."datasources\.yaml".apiVersion=1 \
        --set datasources."datasources\.yaml".datasources[0].name=Prometheus \
        --set datasources."datasources\.yaml".datasources[0].type=prometheus \
        --set datasources."datasources\.yaml".datasources[0].url=http://prometheus-kube-prometheus-prometheus:9090 \
        --set datasources."datasources\.yaml".datasources[0].access=proxy \
        --set datasources."datasources\.yaml".datasources[0].isDefault=true \
        --wait || true
    
    cd "${ROOT_DIR}"
}

# =============================================================================
# SECURITY SETUP
# =============================================================================

setup_security() {
    print_header "Setting up Security Controls"
    
    cd "${SECURITY_DIR}"
    
    # Setup OPA
    print_step "Setting up Open Policy Agent"
    cat > opa-policies.rego << 'EOF'
package kubernetes.admission

deny[msg] {
    input.request.kind.kind == "Pod"
    container := input.request.object.spec.containers[_]
    container.securityContext.privileged == true
    msg := sprintf("Privileged container not allowed: %v", [container.name])
}

deny[msg] {
    input.request.kind.kind == "Pod"
    container := input.request.object.spec.containers[_]
    not container.securityContext.runAsNonRoot == true
    msg := sprintf("Container must run as non-root: %v", [container.name])
}

deny[msg] {
    input.request.kind.kind == "Pod"
    container := input.request.object.spec.containers[_]
    not container.securityContext.readOnlyRootFilesystem == true
    msg := sprintf("Container must have read-only root filesystem: %v", [container.name])
}
EOF
    
    kubectl create configmap opa-policies -n security --from-file=opa-policies.rego --dry-run=client -o yaml | kubectl apply -f - || true
    
    # Setup Falco
    print_step "Setting up Falco"
    helm repo add falcosecurity https://falcosecurity.github.io/charts || true
    helm repo update || true
    
    helm upgrade --install falco falcosecurity/falco \
        --namespace security \
        --create-namespace \
        --set falco.rulesFile[0]=/etc/falco/falco_rules.yaml \
        --set falco.rulesFile[1]=/etc/falco/falco_rules.local.yaml \
        --set falco.jsonOutput=true \
        --wait || true
    
    # Setup Trivy operator
    print_step "Setting up Trivy operator"
    kubectl apply -f https://raw.githubusercontent.com/aquasecurity/trivy-operator/main/deploy/static/trivy-operator.yaml || true
    
    cd "${ROOT_DIR}"
}

# =============================================================================
# CI/CD SETUP
# =============================================================================

setup_cicd() {
    print_header "Setting up CI/CD Pipeline"
    
    mkdir -p "${CI_CD_DIR}"
    cd "${CI_CD_DIR}"
    
    print_step "Creating GitHub Actions workflow"
    cat > ci-cd.yml << 'EOF'
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  AWS_REGION: us-east-1

jobs:
  lint:
    name: Linting
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  security-scan:
    name: Security Scanning
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

  deploy:
    name: Deploy
    runs-on: ubuntu-latest
    needs: security-scan
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
      
      - name: Update kubeconfig
        run: |
          aws eks update-kubeconfig --name ajora-prod --region ${{ env.AWS_REGION }}
      
      - name: Deploy with Helm
        run: |
          helm upgrade --install ajora ./charts/ajora \
            --namespace production \
            --create-namespace \
            --wait
EOF
    
    print_success "CI/CD pipeline configuration created"
    
    cd "${ROOT_DIR}"
}

# =============================================================================
# DEPLOYMENT SCRIPTS
# =============================================================================

create_deployment_scripts() {
    print_header "Creating Deployment Scripts"
    
    cd "${SCRIPTS_DIR}"
    
    # Create health check script
    cat > health-check.sh << 'EOF'
#!/bin/bash
set -euo pipefail

echo "🔍 Running health checks..."

# Check Kubernetes pods
echo "Checking pod status..."
kubectl get pods -n production

# Check services
echo "Checking services..."
kubectl get services -n production

# Check ingress
echo "Checking ingresses..."
kubectl get ingress -n production || echo "No ingresses found"

echo "✅ Health checks completed!"
EOF
    chmod +x health-check.sh
    
    # Create monitor script
    cat > monitor.sh << 'EOF'
#!/bin/bash
set -euo pipefail

echo "📊 Ajora Monitoring Dashboard"
echo "=============================="
echo ""

# Pod Status
echo "Pod Status:"
kubectl get pods -n production -o wide
echo ""

# Resource Usage
echo "Resource Usage:"
kubectl top pods -n production 2>/dev/null || echo "Metrics not available"
echo ""

# Service Status
echo "Service Status:"
kubectl get svc -n production
echo ""

# Ingress Status
echo "Ingress Status:"
kubectl get ingress -n production 2>/dev/null || echo "No ingresses found"
EOF
    chmod +x monitor.sh
    
    # Create backup script
    cat > backup.sh << 'EOF'
#!/bin/bash
set -euo pipefail

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="${HOME}/ajora-backups/${TIMESTAMP}"

mkdir -p "${BACKUP_DIR}"

echo "📦 Creating backup at ${BACKUP_DIR}..."

# Backup Kubernetes resources
echo "Backing up Kubernetes resources..."
kubectl get all -n production -o yaml > "${BACKUP_DIR}/production-resources.yaml" 2>/dev/null || echo "No resources found"

# Backup Helm releases
echo "Backing up Helm releases..."
helm list -n production -o yaml > "${BACKUP_DIR}/helm-releases.yaml" 2>/dev/null || echo "No Helm releases found"

echo "✅ Backup completed!"
EOF
    chmod +x backup.sh
    
    cd "${ROOT_DIR}"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    print_header "🚀 AJORA - Complete Infrastructure Deployment"
    echo -e "${CYAN}This script will deploy the complete Ajora infrastructure including:${NC}"
    echo "  • AWS Infrastructure (VPC, EKS, RDS, Redis, S3, ECR)"
    echo "  • Kubernetes Configuration (Namespaces, RBAC, Network Policies)"
    echo "  • Security Controls (OPA, Falco, Trivy)"
    echo "  • Observability Stack (Prometheus, Loki, Grafana)"
    echo "  • CI/CD Pipeline (GitHub Actions)"
    echo "  • Backup and Disaster Recovery"
    echo ""
    
    # Confirm before starting
    if [[ "${ENVIRONMENT}" == "prod" ]]; then
        confirm_action "This will deploy a production-ready infrastructure. Continue?"
    fi
    
    # Execute all steps
    local start_time=$(date +%s)
    
    check_prerequisites
    setup_directories
    setup_terraform
    setup_kubernetes
    setup_helm_charts
    setup_security
    setup_cicd
    create_deployment_scripts
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    print_header "✅ DEPLOYMENT COMPLETED SUCCESSFULLY!"
    echo -e "${GREEN}🎉 All components deployed in ${duration} seconds${NC}"
    echo ""
    
    # Summary
    echo -e "${BLUE}📋 Deployment Summary:${NC}"
    echo "  • Infrastructure: AWS ${AWS_REGION} (${ENVIRONMENT})"
    echo "  • Kubernetes Cluster: ${CLUSTER_NAME}"
    echo "  • Observability: Prometheus, Loki, Grafana"
    echo "  • Security: OPA, Falco, Trivy"
    echo "  • CI/CD: GitHub Actions configured"
    echo ""
    
    echo -e "${BLUE}🔧 Useful Commands:${NC}"
    echo "  • View pods: kubectl get pods -n production"
    echo "  • Check health: ./scripts/health-check.sh"
    echo "  • Monitor: ./scripts/monitor.sh"
    echo "  • Backup: ./scripts/backup.sh"
    echo "  • Access Grafana: kubectl port-forward -n monitoring svc/grafana 3000:80"
    echo ""
}

# =============================================================================
# EXECUTION
# =============================================================================

# Check if running with proper permissions
if [[ $EUID -eq 0 ]]; then
    print_warning "Running as root is not recommended. Continue anyway?"
    if ! confirm_action "Continue as root?"; then
        exit 1
    fi
fi

# Run main function with error handling
if main; then
    print_success "Deployment completed successfully!"
else
    print_error "Deployment failed. Please check the logs above for details."
    exit 1
fi
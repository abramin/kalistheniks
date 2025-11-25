# Kalistheniks Infrastructure - Terraform

This directory contains Terraform configuration for provisioning the complete AWS infrastructure for the Kalistheniks application.

## Architecture Overview

The infrastructure includes:
- **VPC** with public and private subnets across 2 availability zones
- **Application Load Balancer (ALB)** in public subnet
- **ECS Fargate** cluster for containerized application
- **RDS PostgreSQL** database in private subnet
- **ECR** repository for Docker images
- **SSM Parameter Store** for secrets management
- **Security Groups** with least-privilege access

## File Structure

```
terraform/
├── providers.tf         # AWS provider and version constraints
├── variables.tf         # Input variables
├── network.tf          # VPC, subnets, IGW, NAT gateway, route tables
├── security_groups.tf  # Security groups for ALB, ECS, and RDS
├── alb.tf             # Application Load Balancer, target group, listener
├── ecr.tf             # Elastic Container Registry
├── ecs.tf             # ECS cluster, task definition, service
├── rds.tf             # RDS PostgreSQL instance
├── ssm.tf             # SSM parameters for secrets
└── outputs.tf         # Output values
```

## Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Terraform** >= 1.0 installed
3. AWS account with permissions to create:
   - VPC, Subnets, Internet Gateway, NAT Gateway
   - Security Groups
   - Application Load Balancer
   - ECS Cluster and Services
   - RDS instances
   - ECR repositories
   - SSM parameters
   - IAM roles and policies

## Getting Started

### 1. Configure Variables

Copy the example tfvars file and customize it:

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` and set:
- `db_password` - Strong password for RDS (use a password generator)
- `jwt_secret` - Random secret for JWT signing (min 32 chars)
- `aws_region` - Your preferred AWS region (default: us-east-1)

### 2. Initialize Terraform

```bash
cd terraform
terraform init
```

This downloads the required providers and initializes the backend.

### 3. Review the Plan

```bash
terraform plan
```

Review the resources that will be created. Ensure everything looks correct.

### 4. Apply the Configuration

```bash
terraform apply
```

Type `yes` when prompted. This will:
1. Create the VPC and networking infrastructure (~2-3 min)
2. Provision RDS database (~5-10 min)
3. Create ECR repository
4. Set up ECS cluster and task definition
5. Create ALB and configure routing

**Total time: ~10-15 minutes**

### 5. Note the Outputs

After successful apply, note these important outputs:
- `alb_url` - Your application URL
- `ecr_repository_url` - Where to push your Docker images
- `rds_endpoint` - Database endpoint (also stored in SSM)

## Next Steps

### Push Your Docker Image to ECR

```bash
# Get ECR repository URL from terraform output
ECR_REPO=$(terraform output -raw ecr_repository_url)

# Authenticate Docker to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REPO

# Build and tag your image
docker build -t kalistheniks ../
docker tag kalistheniks:latest $ECR_REPO:latest

# Push to ECR
docker push $ECR_REPO:latest
```

### Run Database Migrations

You can run migrations from a local machine or from an ECS task:

```bash
# Get the DB endpoint
DB_ENDPOINT=$(terraform output -raw rds_address)

# Run migrations (adjust connection string as needed)
# Example: use the go-migrate tool or your application's migration command
```

### Access Your Application

```bash
# Get ALB URL
ALB_URL=$(terraform output -raw alb_url)

# Test health endpoint
curl $ALB_URL/health

# Test your API
curl $ALB_URL/api/v1/exercises
```

## Managing Secrets

All secrets are stored in AWS Systems Manager Parameter Store:
- `/kalistheniks/dev/db/host` - Database host
- `/kalistheniks/dev/db/name` - Database name
- `/kalistheniks/dev/db/user` - Database username
- `/kalistheniks/dev/db/password` - Database password (SecureString)
- `/kalistheniks/dev/jwt/secret` - JWT secret (SecureString)

The ECS task definition automatically loads these as environment variables.

## Cost Estimate (us-east-1)

Approximate monthly costs for dev environment:
- **NAT Gateway**: ~$32/month (+ data transfer)
- **ALB**: ~$16/month (+ LCU charges)
- **ECS Fargate** (1 task, 0.25 vCPU, 0.5 GB): ~$10/month
- **RDS db.t3.micro**: ~$13/month
- **Data Transfer**: Variable
- **ECR Storage**: Minimal for < 10 images

**Total: ~$70-80/month** for a basic dev environment

To reduce costs:
- Use db.t3.micro or db.t4g.micro for RDS
- Consider removing NAT Gateway if outbound internet not needed
- Use smaller ECS task sizes

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

**Warning**: This will delete:
- The RDS database (backups retained per backup_retention_period)
- All container images in ECR
- The entire VPC and networking infrastructure

## Security Notes

- RDS is in a private subnet with no public access
- Security groups follow principle of least privilege
- Secrets are stored in SSM Parameter Store (SecureString)
- Database password is marked sensitive in Terraform
- ECS tasks pull secrets at runtime via IAM roles
- All resources are tagged for easy identification

## Troubleshooting

### ECS Task Won't Start

Check CloudWatch Logs:
```bash
aws logs tail /ecs/kalistheniks --follow
```

Common issues:
- Docker image not found in ECR
- Task execution role missing permissions
- Invalid environment variables

### Can't Connect to RDS

- Verify security group allows ECS -> RDS traffic
- Check RDS is in private subnet
- Verify SSM parameters are correct

### ALB Health Checks Failing

- Ensure your app responds to `/health` endpoint
- Check ECS task logs for application errors
- Verify container port matches task definition

## Additional Resources

- [AWS ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)
- [Terraform AWS Provider Docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

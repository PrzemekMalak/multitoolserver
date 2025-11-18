# AWS ECS Deployment with CDK

This directory contains AWS CDK (Cloud Development Kit) code to deploy the Multitool Server to AWS ECS using Fargate.

## Architecture

The CDK stack creates:
- **VPC**: A new VPC with 2 availability zones and 1 NAT gateway
- **ECS Cluster**: Fargate cluster for running containers
- **Task Definition**: Defines the container image and resources (512 MB memory, 256 CPU units)
- **Fargate Service**: Runs 2 tasks for high availability
- **Application Load Balancer**: Public-facing load balancer for external access
- **CloudWatch Logs**: Automatic logging for container logs

## Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Python 3.8+** installed
3. **AWS CDK CLI** installed:
   ```bash
   npm install -g aws-cdk
   ```
4. **Docker image** available: `przemekmalak/multitoolserver:latest`

## Setup

1. **Navigate to the CDK directory:**
   ```bash
   cd cdk
   ```

2. **Create a Python virtual environment:**
   ```bash
   python3 -m venv .venv
   source .venv/bin/activate  # On Windows: .venv\Scripts\activate
   ```

3. **Install dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

4. **Bootstrap CDK (first time only):**
   ```bash
   cdk bootstrap
   ```

## Deployment

1. **Synthesize the CloudFormation template:**
   ```bash
   cdk synth
   ```

2. **Deploy the stack:**
   ```bash
   cdk deploy
   ```

3. **After deployment, the stack will output:**
   - `LoadBalancerDNS`: The DNS name of the load balancer
   - `ServiceURL`: The full URL to access your service

## Configuration

You can customize the deployment by editing `multitool_stack.py`:

- **Desired count**: Change `desired_count=2` to run more or fewer tasks
- **CPU/Memory**: Adjust `memory_limit_mib` and `cpu` in the task definition
- **Docker image**: Change the image name in `ContainerImage.from_registry()`
- **Environment variables**: Add or modify environment variables in the `environment` dict
- **Health check path**: Modify the health check path in `configure_health_check()`

## Useful Commands

- `cdk ls` - List all stacks
- `cdk synth` - Synthesize CloudFormation template
- `cdk deploy` - Deploy the stack
- `cdk diff` - Compare deployed stack with current state
- `cdk destroy` - Destroy the stack and all resources

## Cost Considerations

This deployment uses:
- **Fargate**: Pay per vCPU and memory used
- **Application Load Balancer**: ~$16/month + data transfer
- **NAT Gateway**: ~$32/month + data transfer
- **CloudWatch Logs**: Pay per GB ingested

Estimated monthly cost: ~$50-100 depending on usage.

## Troubleshooting

1. **Check CloudWatch Logs:**
   ```bash
   aws logs tail /aws/ecs/multitool --follow
   ```

2. **View service status:**
   ```bash
   aws ecs describe-services --cluster multitool-cluster --services multitool-service
   ```

3. **Check task status:**
   ```bash
   aws ecs list-tasks --cluster multitool-cluster
   ```

## Cleanup

To remove all resources:
```bash
cdk destroy
```


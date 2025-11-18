from aws_cdk import (
    Stack,
    Duration,
    aws_ec2 as ec2,
    aws_ecs as ecs,
    aws_ecs_patterns as ecs_patterns,
    aws_iam as iam,
    CfnOutput,
)
from constructs import Construct


class MultitoolStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        # Create VPC (or use existing)
        vpc = ec2.Vpc(
            self,
            "MultitoolVpc",
            max_azs=2,
            nat_gateways=1,
        )

        # Create ECS Cluster
        cluster = ecs.Cluster(
            self,
            "MultitoolCluster",
            vpc=vpc,
            cluster_name="multitool-cluster",
        )

        # Create Fargate Task Definition
        task_definition = ecs.FargateTaskDefinition(
            self,
            "MultitoolTaskDef",
            memory_limit_mib=512,
            cpu=256,
        )

        # Add container to task definition
        container = task_definition.add_container(
            "multitool-container",
            image=ecs.ContainerImage.from_registry("przemekmalak/multitoolserver:latest"),
            logging=ecs.LogDrivers.aws_logs(
                stream_prefix="multitool",
            ),
            environment={
                "RETURN_TEXT": "from ECS",
            },
        )

        # Add port mapping
        container.add_port_mappings(
            ecs.PortMapping(
                container_port=8080,
                protocol=ecs.Protocol.TCP,
            )
        )

        # Create Fargate Service with Application Load Balancer
        fargate_service = ecs_patterns.ApplicationLoadBalancedFargateService(
            self,
            "MultitoolService",
            cluster=cluster,
            task_definition=task_definition,
            desired_count=2,  # Run 2 tasks for high availability
            public_load_balancer=True,
            listener_port=80,
            service_name="multitool-service",
        )

        # Configure health check
        fargate_service.target_group.configure_health_check(
            path="/hello",
            interval=Duration.seconds(30),
            timeout=Duration.seconds(5),
            healthy_threshold_count=2,
            unhealthy_threshold_count=3,
        )

        # Allow container to make outbound HTTP requests (for /req endpoint)
        task_definition.task_role.add_to_policy(
            iam.PolicyStatement(
                effect=iam.Effect.ALLOW,
                actions=["ec2:DescribeNetworkInterfaces"],
                resources=["*"],
            )
        )

        # Output the load balancer URL
        CfnOutput(
            self,
            "LoadBalancerDNS",
            value=fargate_service.load_balancer.load_balancer_dns_name,
            description="DNS name of the load balancer",
        )

        CfnOutput(
            self,
            "ServiceURL",
            value=f"http://{fargate_service.load_balancer.load_balancer_dns_name}",
            description="URL to access the Multitool Server",
        )


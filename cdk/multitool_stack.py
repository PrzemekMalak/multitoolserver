from aws_cdk import (
    Stack,
    Duration,
    RemovalPolicy,
    aws_ec2 as ec2,
    aws_ecs as ecs,
    aws_ecs_patterns as ecs_patterns,
    aws_iam as iam,
    aws_logs as logs,
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

        # Create CloudWatch Log Group for Container Insights
        log_group = logs.LogGroup(
            self,
            "MultitoolLogGroup",
            log_group_name="/aws/ecs/multitool",
            retention=logs.RetentionDays.ONE_WEEK,
            removal_policy=RemovalPolicy.DESTROY,
        )

        # Create ECS Cluster with Container Insights enabled
        cluster = ecs.Cluster(
            self,
            "MultitoolCluster",
            vpc=vpc,
            cluster_name="multitool-cluster",
            container_insights=True,  # Enable Container Insights
        )

        # Create Fargate Task Definition
        task_definition = ecs.FargateTaskDefinition(
            self,
            "MultitoolTaskDef",
            memory_limit_mib=512,
            cpu=256,
        )

        # Add container to task definition with enhanced logging
        container = task_definition.add_container(
            "multitool-container",
            image=ecs.ContainerImage.from_registry("przemekmalak/multitoolserver:latest-amd64"),
            logging=ecs.LogDrivers.aws_logs(
                stream_prefix="multitool",
                log_group=log_group,
            ),
            environment={
                "RETURN_TEXT": "from ECS",
                # Process monitoring configuration
                "MONITOR_INTERVAL": "30",  # Collect metrics every 30 seconds
                "MONITOR_FILTER": "",  # Empty = monitor all processes, or set to "serv" to filter
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

        # Grant CloudWatch Logs permissions for Container Insights
        log_group.grant_write(task_definition.task_role)

        # Create a metric filter to extract process metrics from monitoring logs
        # This allows querying process/thread metrics in CloudWatch
        logs.MetricFilter(
            self,
            "ProcessMetricsFilter",
            log_group=log_group,
            metric_namespace="ECS/ProcessMonitoring",
            metric_name="ProcessCount",
            filter_pattern=logs.FilterPattern.exists("$.type"),
            metric_value="1",
            default_value=0,
        )

        # Add CloudWatch Container Insights output
        CfnOutput(
            self,
            "ContainerInsightsDashboard",
            value=f"https://console.aws.amazon.com/cloudwatch/home?region={self.region}#containerInsights:clusters/multitool-cluster",
            description="Link to Container Insights dashboard in CloudWatch",
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

        # Output CloudWatch Logs information
        CfnOutput(
            self,
            "CloudWatchLogGroup",
            value=log_group.log_group_name,
            description="CloudWatch Log Group for container and process monitoring logs",
        )

        CfnOutput(
            self,
            "CloudWatchLogsInsights",
            value=f"https://console.aws.amazon.com/cloudwatch/home?region={self.region}#logsV2:log-groups/log-group/{log_group.log_group_name.replace('/', '$252F')}",
            description="Link to CloudWatch Logs Insights for querying process monitoring logs",
        )


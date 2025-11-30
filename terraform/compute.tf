resource "aws_ecs_cluster" "main" {
  name = "Ecs-${var.environment}-UsEast1-PayBank"

  tags = {
    Name = "Ecs-${var.environment}-UsEast1-PayBank"
  }
}

resource "aws_ecs_task_definition" "api" {
  family                   = "Task-${var.environment}-UsEast1-PayBank-Api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  task_role_arn            = aws_iam_role.main.arn
  execution_role_arn       = aws_iam_role.main.arn
  cpu                      = 256
  memory                   = 512

  container_definitions = jsonencode([
    {
      name      = "api"
      image     = "${aws_ecr_repository.api.repository_url}:latest"
      essential = true
      portMappings = [
        {
          appProtocol   = "http"
          hostPort      = 3333
          containerPort = 3333
        }
      ]
    }
  ])

  tags = {
    Name = "Task-${var.environment}-UsEast1-PayBank-Api"
  }
}

resource "aws_ecr_repository" "api" {
  name                 = "paybank-api"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  tags = {
    Name = "Ecr-${var.environment}-UsEast1-PayBank-Api"
  }
}

resource "aws_lb" "main" {
  name               = "Alb-${var.environment}-UsEast1-PayBank"
  load_balancer_type = "application"
  subnets            = [aws_subnet.public_1a.id, aws_subnet.public_1b.id]
  security_groups    = [aws_security_group.alb.id]
  internal           = false

  tags = {
    Name = "Alb-${var.environment}-UsEast1-PayBank"
  }
}

resource "aws_lb_target_group" "tg_main" {
  name        = "Tg-${var.environment}-UsEast1-PayBank"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"
  protocol    = "HTTP"
  port        = 3333

  health_check {
    path                = "/health"
    protocol            = "HTTP"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 5
    unhealthy_threshold = 2
  }

  tags = {
    Name = "Tg-${var.environment}-UsEast1-PayBank"
  }
}

resource "aws_lb_listener" "listener_main" {
  load_balancer_arn = aws_lb.main.arn
  protocol          = "HTTP"
  port              = 80

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.tg_main.arn
  }

  tags = {
    Name = "Ltr-${var.environment}-UsEast1-PayBank"
  }
}

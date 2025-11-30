resource "aws_iam_role" "main" {
  name = "Iam-${var.environment}-UsEast1-PayBank"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "Iam-${var.environment}-UsEast1-PayBank"
  }
}

resource "aws_iam_role_policy_attachment" "attachment_1" {
  role       = aws_iam_role.main.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "attachment_2" {
  role       = aws_iam_role.main.name
  policy_arn = "arn:aws:iam::aws:policy/AWSSecretsManagerClientReadOnlyAccess"
}

resource "aws_security_group" "api" {
  name   = "Sgp-${var.environment}-UsEast1-PayBank-Api"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "Sgp-${var.environment}-UsEast1-PayBank-Api"
  }
}

resource "aws_vpc_security_group_ingress_rule" "api_allow_3333" {
  security_group_id = aws_security_group.api.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 3333
  to_port           = 3333
}

resource "aws_vpc_security_group_egress_rule" "api_allow_all_outbound" {
  security_group_id = aws_security_group.api.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = -1
}

resource "aws_security_group" "alb" {
  name   = "Sgp-${var.environment}-UsEast1-PayBank-Alb"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "Sgp-${var.environment}-UsEast1-PayBank-Alb"
  }
}

resource "aws_vpc_security_group_ingress_rule" "alb_allow_http" {
  security_group_id = aws_security_group.alb.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 80
  to_port           = 80
}

resource "aws_vpc_security_group_egress_rule" "alb_allow_all_outbound" {
  security_group_id = aws_security_group.alb.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = -1
}

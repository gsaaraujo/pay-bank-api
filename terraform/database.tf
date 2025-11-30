resource "aws_db_subnet_group" "db" {
  subnet_ids = [aws_subnet.private_1a.id, aws_subnet.private_1b.id]

  tags = {
    Name = "SubnetGroup-${var.environment}-UsEast1-PayBank"
  }
}

resource "aws_db_instance" "rds" {
  vpc_security_group_ids = [aws_security_group.api.id]
  db_subnet_group_name   = aws_db_subnet_group.db.name
  engine                 = "postgres"
  engine_version         = "17.6"
  instance_class         = "db.t3.micro"
  allocated_storage      = 20
  db_name                = "postgres"
  username               = "postgres"
  password               = var.db_password
  apply_immediately      = true

  tags = {
    Name = "Rds-${var.environment}-UsEast1-PayBank"
  }
}

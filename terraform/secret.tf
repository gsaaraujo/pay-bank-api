resource "aws_secretsmanager_secret" "api" {
  name                    = "Secret-${var.environment}-UsEast1-PayBank-Api"
  recovery_window_in_days = 0

  tags = {
    Name = "Secret-${var.environment}-UsEast1-PayBank-Api"
  }
}

resource "aws_secretsmanager_secret_version" "api" {
  secret_id = aws_secretsmanager_secret.api.id

  secret_string = jsonencode({
    POSTGRES_URL             = "postgres://${aws_db_instance.rds.username}:${var.db_password}@${aws_db_instance.rds.endpoint}:${aws_db_instance.rds.port}/${aws_db_instance.rds.db_name}"
    ACCESS_TOKEN_SIGNING_KEY = var.access_token_signing_key
  })
}

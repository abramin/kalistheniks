# SSM Parameter - Database Host
resource "aws_ssm_parameter" "db_host" {
  name        = "/${var.project_name}/${var.environment}/db/host"
  description = "Database host endpoint"
  type        = "String"
  value       = aws_db_instance.main.address

  tags = {
    Name = "${var.project_name}-db-host"
  }
}

# SSM Parameter - Database Name
resource "aws_ssm_parameter" "db_name" {
  name        = "/${var.project_name}/${var.environment}/db/name"
  description = "Database name"
  type        = "String"
  value       = var.db_name

  tags = {
    Name = "${var.project_name}-db-name"
  }
}

# SSM Parameter - Database User
resource "aws_ssm_parameter" "db_user" {
  name        = "/${var.project_name}/${var.environment}/db/user"
  description = "Database username"
  type        = "String"
  value       = var.db_username

  tags = {
    Name = "${var.project_name}-db-user"
  }
}

# SSM Parameter - Database Password
resource "aws_ssm_parameter" "db_password" {
  name        = "/${var.project_name}/${var.environment}/db/password"
  description = "Database password"
  type        = "SecureString"
  value       = var.db_password

  tags = {
    Name = "${var.project_name}-db-password"
  }
}

# SSM Parameter - JWT Secret
resource "aws_ssm_parameter" "jwt_secret" {
  name        = "/${var.project_name}/${var.environment}/jwt/secret"
  description = "JWT secret for application"
  type        = "SecureString"
  value       = var.jwt_secret

  tags = {
    Name = "${var.project_name}-jwt-secret"
  }
}

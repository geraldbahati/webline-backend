#!/bin/bash
# CI/CD setup script for webline backend

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Webline Backend CI/CD Setup ===${NC}"
echo -e "${YELLOW}This script will help you configure your CI/CD pipeline.${NC}"
echo ""

# Check for required tools
echo -e "${BLUE}Checking for required tools...${NC}"
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Docker is required but not installed. Aborting.${NC}" >&2; exit 1; }
command -v doctl >/dev/null 2>&1 || { echo -e "${YELLOW}Warning: doctl (Digital Ocean CLI) is not installed. You may need to install it for full functionality.${NC}" >&2; }
command -v gh >/dev/null 2>&1 || { echo -e "${YELLOW}Warning: GitHub CLI is not installed. You'll need to set up secrets manually.${NC}" >&2; }

echo -e "${GREEN}All required tools are available.${NC}"
echo ""

# Docker Hub Configuration
echo -e "${BLUE}Setting up Docker Hub configuration...${NC}"
read -p "Docker Hub Username: " docker_username
read -s -p "Docker Hub Password/Token: " docker_password
echo ""
echo -e "${GREEN}Docker Hub configuration saved.${NC}"
echo ""

# Digital Ocean Configuration
echo -e "${BLUE}Setting up Digital Ocean configuration...${NC}"
read -p "Digital Ocean API Token: " do_token
read -p "Path to SSH Private Key for Digital Ocean (default: ~/.ssh/id_rsa): " ssh_key_path
ssh_key_path=${ssh_key_path:-~/.ssh/id_rsa}

if [ ! -f "$ssh_key_path" ]; then
    echo -e "${RED}SSH key file not found at $ssh_key_path${NC}"
    exit 1
fi

ssh_key_fingerprint=$(ssh-keygen -l -E md5 -f "$ssh_key_path" | awk '{print $2}' | cut -d':' -f2-)
ssh_private_key=$(cat "$ssh_key_path")

echo -e "${GREEN}Digital Ocean configuration saved.${NC}"
echo ""

# Load values from .env file if it exists
ENV_FILE=".env"
if [ -f "$ENV_FILE" ]; then
    echo -e "${BLUE}Loading values from .env file...${NC}"
    # Export all variables from .env file
    export $(grep -v '^#' "$ENV_FILE" | xargs)
    echo -e "${GREEN}Values loaded from .env file.${NC}"
else
    echo -e "${YELLOW}No .env file found, you will need to enter values manually.${NC}"
fi

# Database Configuration
echo -e "${BLUE}Setting up Database configuration...${NC}"
read -p "Database User [${DB_USER}]: " db_user
db_user=${db_user:-${DB_USER}}
read -s -p "Database Password [keep existing]: " db_password
echo ""
db_password=${db_password:-${DB_PASSWORD}}
read -p "Database Name [${DB_NAME}]: " db_name
db_name=${db_name:-${DB_NAME}}
echo -e "${GREEN}Database configuration saved.${NC}"
echo ""

# AWS Configuration
echo -e "${BLUE}Setting up AWS configuration...${NC}"
read -p "AWS Access Key ID [${AWS_ACCESS_KEY_ID}]: " aws_access_key_id
aws_access_key_id=${aws_access_key_id:-${AWS_ACCESS_KEY_ID}}
read -s -p "AWS Secret Access Key [keep existing]: " aws_secret_access_key
echo ""
aws_secret_access_key=${aws_secret_access_key:-${AWS_SECRET_ACCESS_KEY}}
read -p "AWS Region [${AWS_REGION}]: " aws_region
aws_region=${aws_region:-${AWS_REGION}}
read -p "AWS Bucket Name [${AWS_BUCKET_NAME}]: " aws_bucket_name
aws_bucket_name=${aws_bucket_name:-${AWS_BUCKET_NAME}}
echo -e "${GREEN}AWS configuration saved.${NC}"
echo ""

# MPESA Configuration
echo -e "${BLUE}Setting up MPESA configuration...${NC}"
read -p "Business Shortcode [${BUSINESS_SHORTCODE}]: " business_shortcode
business_shortcode=${business_shortcode:-${BUSINESS_SHORTCODE}}
read -s -p "Passkey [keep existing]: " passkey
echo ""
passkey=${passkey:-${PASSKEY}}
read -p "Callback URL [${CALLBACK_URL}]: " callback_url
callback_url=${callback_url:-${CALLBACK_URL}}
read -p "Consumer Key [${CONSUMER_KEY}]: " consumer_key
consumer_key=${consumer_key:-${CONSUMER_KEY}}
read -s -p "Consumer Secret [keep existing]: " consumer_secret
echo ""
consumer_secret=${consumer_secret:-${CONSUMER_SECRET}}
read -p "Account Reference [${ACCOUNT_REFERENCE}]: " account_reference
account_reference=${account_reference:-${ACCOUNT_REFERENCE}}
echo -e "${GREEN}MPESA configuration saved.${NC}"
echo ""

# SMTP Configuration
echo -e "${BLUE}Setting up SMTP configuration...${NC}"
read -p "SMTP Host [${SMTP_HOST}]: " smtp_host
smtp_host=${smtp_host:-${SMTP_HOST}}
read -p "SMTP Port [${SMTP_PORT}]: " smtp_port
smtp_port=${smtp_port:-${SMTP_PORT}}
read -p "SMTP Username [${SMTP_USERNAME}]: " smtp_username
smtp_username=${smtp_username:-${SMTP_USERNAME}}
read -s -p "SMTP Password [keep existing]: " smtp_password
echo ""
smtp_password=${smtp_password:-${SMTP_PASSWORD}}
read -p "From Email [${FROM_EMAIL}]: " from_email
from_email=${from_email:-${FROM_EMAIL}}
read -p "From Name [${FROM_NAME}]: " from_name
from_name=${from_name:-${FROM_NAME}}
read -p "To Email [${TO_EMAIL}]: " to_email
to_email=${to_email:-${TO_EMAIL}}
echo -e "${GREEN}SMTP configuration saved.${NC}"
echo ""

# URL Configuration
echo -e "${BLUE}Setting up URL configuration...${NC}"
read -p "Frontend URL [${FRONTEND_URL}]: " frontend_url
frontend_url=${frontend_url:-${FRONTEND_URL}}
read -p "Backend URL [${BACKEND_URL}]: " backend_url
backend_url=${backend_url:-${BACKEND_URL}}
echo -e "${GREEN}URL configuration saved.${NC}"
echo ""

# Redis Configuration
echo -e "${BLUE}Setting up Redis configuration...${NC}"
read -s -p "Redis Password [keep existing]: " redis_password
echo ""
redis_password=${redis_password:-${REDIS_PASSWORD}}
echo -e "${GREEN}Redis configuration saved.${NC}"
echo ""

# JWT Configuration
echo -e "${BLUE}Setting up JWT configuration...${NC}"
read -s -p "JWT Access Secret [keep existing]: " jwt_access_secret
echo ""
jwt_access_secret=${jwt_access_secret:-${JWT_ACCESS_SECRET}}
read -s -p "JWT Refresh Secret [keep existing]: " jwt_refresh_secret
echo ""
jwt_refresh_secret=${jwt_refresh_secret:-${JWT_REFRESH_SECRET}}
read -s -p "JWT Verify Secret [keep existing]: " jwt_verify_secret
echo ""
jwt_verify_secret=${jwt_verify_secret:-${JWT_VERIFY_SECRET}}
read -s -p "JWT Guest Secret [keep existing]: " jwt_guest_secret
echo ""
jwt_guest_secret=${jwt_guest_secret:-${JWT_GUEST_SECRET}}
echo -e "${GREEN}JWT configuration saved.${NC}"
echo ""

# CORS Configuration
echo -e "${BLUE}Setting up CORS configuration...${NC}"
read -p "Allowed Origins (comma-separated) [${ALLOWED_ORIGINS}]: " allowed_origins
allowed_origins=${allowed_origins:-${ALLOWED_ORIGINS}}
echo -e "${GREEN}CORS configuration saved.${NC}"
echo ""

# Setting up GitHub Secrets
if command -v gh >/dev/null 2>&1; then
    echo -e "${BLUE}Setting up GitHub Secrets...${NC}"
    echo -e "${YELLOW}You'll need to be logged in to GitHub CLI.${NC}"

    # Check if logged in
    gh auth status || gh auth login

    # Set Docker Hub secrets
    echo -e "${BLUE}Setting DOCKER_HUB_USERNAME secret...${NC}"
    gh secret set DOCKER_HUB_USERNAME -b"$docker_username"

    echo -e "${BLUE}Setting DOCKER_HUB_PASSWORD secret...${NC}"
    gh secret set DOCKER_HUB_PASSWORD -b"$docker_password"

    # Set Digital Ocean secrets
    echo -e "${BLUE}Setting DO_TOKEN secret...${NC}"
    gh secret set DO_TOKEN -b"$do_token"

    echo -e "${BLUE}Setting DO_SSH_FINGERPRINT secret...${NC}"
    gh secret set DO_SSH_FINGERPRINT -b"$ssh_key_fingerprint"

    echo -e "${BLUE}Setting DO_SSH_PRIVATE_KEY secret...${NC}"
    gh secret set DO_SSH_PRIVATE_KEY -b"$ssh_private_key"

    # Set Database secrets
    echo -e "${BLUE}Setting DB_USER secret...${NC}"
    gh secret set DB_USER -b"$db_user"

    echo -e "${BLUE}Setting DB_PASSWORD secret...${NC}"
    gh secret set DB_PASSWORD -b"$db_password"

    echo -e "${BLUE}Setting DB_NAME secret...${NC}"
    gh secret set DB_NAME -b"$db_name"

    # Set AWS secrets
    echo -e "${BLUE}Setting AWS_ACCESS_KEY_ID secret...${NC}"
    gh secret set AWS_ACCESS_KEY_ID -b"$aws_access_key_id"

    echo -e "${BLUE}Setting AWS_SECRET_ACCESS_KEY secret...${NC}"
    gh secret set AWS_SECRET_ACCESS_KEY -b"$aws_secret_access_key"

    echo -e "${BLUE}Setting AWS_REGION secret...${NC}"
    gh secret set AWS_REGION -b"$aws_region"

    echo -e "${BLUE}Setting AWS_BUCKET_NAME secret...${NC}"
    gh secret set AWS_BUCKET_NAME -b"$aws_bucket_name"

    # Set MPESA secrets
    echo -e "${BLUE}Setting BUSINESS_SHORTCODE secret...${NC}"
    gh secret set BUSINESS_SHORTCODE -b"$business_shortcode"

    echo -e "${BLUE}Setting PASSKEY secret...${NC}"
    gh secret set PASSKEY -b"$passkey"

    echo -e "${BLUE}Setting CALLBACK_URL secret...${NC}"
    gh secret set CALLBACK_URL -b"$callback_url"

    echo -e "${BLUE}Setting CONSUMER_KEY secret...${NC}"
    gh secret set CONSUMER_KEY -b"$consumer_key"

    echo -e "${BLUE}Setting CONSUMER_SECRET secret...${NC}"
    gh secret set CONSUMER_SECRET -b"$consumer_secret"

    echo -e "${BLUE}Setting ACCOUNT_REFERENCE secret...${NC}"
    gh secret set ACCOUNT_REFERENCE -b"$account_reference"

    # Set SMTP secrets
    echo -e "${BLUE}Setting SMTP_HOST secret...${NC}"
    gh secret set SMTP_HOST -b"$smtp_host"

    echo -e "${BLUE}Setting SMTP_PORT secret...${NC}"
    gh secret set SMTP_PORT -b"$smtp_port"

    echo -e "${BLUE}Setting SMTP_USERNAME secret...${NC}"
    gh secret set SMTP_USERNAME -b"$smtp_username"

    echo -e "${BLUE}Setting SMTP_PASSWORD secret...${NC}"
    gh secret set SMTP_PASSWORD -b"$smtp_password"

    echo -e "${BLUE}Setting FROM_EMAIL secret...${NC}"
    gh secret set FROM_EMAIL -b"$from_email"

    echo -e "${BLUE}Setting FROM_NAME secret...${NC}"
    gh secret set FROM_NAME -b"$from_name"

    echo -e "${BLUE}Setting TO_EMAIL secret...${NC}"
    gh secret set TO_EMAIL -b"$to_email"

    # Set URL secrets
    echo -e "${BLUE}Setting FRONTEND_URL secret...${NC}"
    gh secret set FRONTEND_URL -b"$frontend_url"

    echo -e "${BLUE}Setting BACKEND_URL secret...${NC}"
    gh secret set BACKEND_URL -b"$backend_url"

    # Set Redis secrets
    echo -e "${BLUE}Setting REDIS_PASSWORD secret...${NC}"
    gh secret set REDIS_PASSWORD -b"$redis_password"

    # Set JWT secrets
    echo -e "${BLUE}Setting JWT_ACCESS_SECRET secret...${NC}"
    gh secret set JWT_ACCESS_SECRET -b"$jwt_access_secret"

    echo -e "${BLUE}Setting JWT_REFRESH_SECRET secret...${NC}"
    gh secret set JWT_REFRESH_SECRET -b"$jwt_refresh_secret"

    echo -e "${BLUE}Setting JWT_VERIFY_SECRET secret...${NC}"
    gh secret set JWT_VERIFY_SECRET -b"$jwt_verify_secret"

    echo -e "${BLUE}Setting JWT_GUEST_SECRET secret...${NC}"
    gh secret set JWT_GUEST_SECRET -b"$jwt_guest_secret"

    # Set CORS secrets
    echo -e "${BLUE}Setting ALLOWED_ORIGINS secret...${NC}"
    gh secret set ALLOWED_ORIGINS -b"$allowed_origins"

    echo -e "${GREEN}GitHub secrets have been set successfully.${NC}"
else
    echo -e "${YELLOW}GitHub CLI not found. You'll need to manually set all these secrets in your GitHub repository.${NC}"
    echo -e "Here's a list of all the secrets you need to set:"
    echo -e "${BLUE}DOCKER_HUB_USERNAME${NC}"
    echo -e "${BLUE}DOCKER_HUB_PASSWORD${NC}"
    echo -e "${BLUE}DO_TOKEN${NC}"
    echo -e "${BLUE}DO_SSH_FINGERPRINT${NC}"
    echo -e "${BLUE}DO_SSH_PRIVATE_KEY${NC}"
    echo -e "${BLUE}DB_USER${NC}"
    echo -e "${BLUE}DB_PASSWORD${NC}"
    echo -e "${BLUE}DB_NAME${NC}"
    echo -e "${BLUE}AWS_ACCESS_KEY_ID${NC}"
    echo -e "${BLUE}AWS_SECRET_ACCESS_KEY${NC}"
    echo -e "${BLUE}AWS_REGION${NC}"
    echo -e "${BLUE}AWS_BUCKET_NAME${NC}"
    echo -e "${BLUE}BUSINESS_SHORTCODE${NC}"
    echo -e "${BLUE}PASSKEY${NC}"
    echo -e "${BLUE}CALLBACK_URL${NC}"
    echo -e "${BLUE}CONSUMER_KEY${NC}"
    echo -e "${BLUE}CONSUMER_SECRET${NC}"
    echo -e "${BLUE}ACCOUNT_REFERENCE${NC}"
    echo -e "${BLUE}SMTP_HOST${NC}"
    echo -e "${BLUE}SMTP_PORT${NC}"
    echo -e "${BLUE}SMTP_USERNAME${NC}"
    echo -e "${BLUE}SMTP_PASSWORD${NC}"
    echo -e "${BLUE}FROM_EMAIL${NC}"
    echo -e "${BLUE}FROM_NAME${NC}"
    echo -e "${BLUE}TO_EMAIL${NC}"
    echo -e "${BLUE}FRONTEND_URL${NC}"
    echo -e "${BLUE}BACKEND_URL${NC}"
    echo -e "${BLUE}REDIS_PASSWORD${NC}"
    echo -e "${BLUE}JWT_ACCESS_SECRET${NC}"
    echo -e "${BLUE}JWT_REFRESH_SECRET${NC}"
    echo -e "${BLUE}JWT_VERIFY_SECRET${NC}"
    echo -e "${BLUE}JWT_GUEST_SECRET${NC}"
    echo -e "${BLUE}ALLOWED_ORIGINS${NC}"
fi

echo ""
echo -e "${GREEN}CI/CD setup completed successfully!${NC}"
echo -e "${BLUE}Next steps:${NC}"
echo -e "1. Push your code to GitHub to trigger the CI/CD pipeline."
echo -e "2. Monitor the GitHub Actions workflow to ensure successful deployment."
echo -e "3. Check your Digital Ocean dashboard for new droplets created by the pipeline."

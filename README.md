## Webline Backend Documentation

This documentation outlines the functionality of the Webline Backend system, including its various components and how they interact.

### System Overview

The Webline Backend is a robust system designed to handle the following key functionalities:

- **User Management:**  Manages user registration, authentication, profile updates, and role assignments.
- **Product Management:** Enables the administration of products, including creating, updating, deleting, and managing images and specifications.
- **Category Management:** Provides tools for organizing products into a hierarchical structure of categories.
- **Order Management:** Handles order creation, processing, payment, and shipment management.
- **Discount Management:** Allows for the definition and application of discounts on products.
- **Promotion Management:** Enables the creation of promotional campaigns featuring product selections and associated details.
- **Admin Requests:** Facilitates user requests to be elevated to an admin role with approval processes.

### Deployment

The backend application is designed to be containerized using Docker and Docker Compose. This ensures easy deployment and portability. The following files are critical to the deployment process:

- `.gitignore` : Specifies files to be excluded from version control.
- `Dockerfile` : Contains instructions for building the Docker image.
- `docker-compose.yml` : Defines the services and dependencies required for running the application with Docker Compose.
- `entrypoint.sh` : Provides a script to execute upon container startup, handling migrations and starting the application.
- `crontab-appuser` : Defines scheduled tasks, such as backups, for the application.
- `backup.sh` :  A shell script to perform database backups and store them on AWS S3.

#### Local Development

For local development, you can use the provided script to build and run the application:

```bash
./scripts/local-build.sh
```

This script will:
1. Check for Docker and Docker Compose
2. Create a basic `.env` file if one doesn't exist
3. Modify the `docker-compose.yml` to use local builds instead of Docker Hub images
4. Build and start the containers
5. Display the status of the running containers

The application will be available at http://localhost:8080.

### Database Schema

The backend utilizes a PostgreSQL database to store data. The `sql/schema` directory contains SQL scripts for creating and modifying the database schema:

- `0001_initial.sql` :  Creates initial database tables for users, products, categories, orders, and related entities.
- `0002_tables_indexes.sql` : Adds indexes for improved performance on key tables.
- Subsequent migrations (e.g., `20240521120328_add_is_active_to_categories.sql`, `20240610062206_add_product_sizes_table.sql`, etc.): Introduce schema changes and optimizations over time.

### Code Structure

- **cmd:**  Contains the main entry point for the application.
- **internal:** Houses the core application logic:
    - **app_errors:** Defines application-specific error types and functions.
    - **appconfig:**  Manages configuration loading and environment variables.
    - **logger:** Provides logging functionality using Zap library.
    - **database:** Handles interactions with the PostgreSQL database using SQLC.
    - **handlers:** Implements HTTP request handlers for various endpoints.
    - **middleware:** Provides middleware to handle CORS, error handling, authentication, and metrics collection.
    - **model:** Defines the data structures used for data exchange between components.
    - **repository:**  Defines interfaces and implementations for database interactions.
    - **routes:**  Handles route registration and configuration.
    - **server:** Creates the server instance and initializes dependencies.
    - **services:** Implements the core business logic for the application.
- **pkg:** Contains utility packages:
    - **background:**  Handles background tasks (not implemented yet).
    - **mpesa:** Provides integration with M-Pesa (Safaricom's mobile money service).
    - **utils:**  Includes various utility functions for JWT generation, email sending, file management, and price formatting.

###  Code Walkthrough

The codebase is structured using a layered approach:

1. **Server:** The `internal/server/server.go` file initializes the server instance. It loads configuration, establishes database connections, creates S3 and Redis clients, instantiates repositories and services, and finally sets up the HTTP router.
2. **Services:** The `internal/services` directory contains services that handle the business logic for various parts of the application. For instance, the `UserService` manages user registration, authentication, and profile updates, while the `ProductService` handles product creation, retrieval, and updates.
3. **Handlers:** The `internal/handlers` directory holds HTTP request handlers that translate incoming requests into actions performed by the corresponding services. These handlers provide the API interface to the external world.
4. **Repositories:** The `internal/repository` directory defines interfaces and implementations for interacting with the database. It provides a layer of abstraction for database access.
5. **Model:** The `internal/model` directory defines data structures (structs) that are used for representing and exchanging data within the application.

###  API Endpoints

The Webline Backend exposes a range of API endpoints for various operations. For detailed information on each endpoint, please refer to the `internal/routes` directory, specifically `routes.go`.

### Contributing

Contributions to the Webline Backend are welcome!  Here are some ways you can contribute:

- **Bug Reports:** If you encounter any issues, please file a bug report with clear steps to reproduce the problem.
- **Feature Requests:** Suggest new features or improvements to enhance the functionality of the system.
- **Code Contributions:** Submit pull requests for bug fixes or new feature implementations.

###  License

The Webline Backend project is licensed under the MIT License. See the LICENSE file for more details.

##  Appendix

- **Database Queries:**  The `sql/queries` directory contains SQL queries generated using SQLC.
- **Database Migrations:** The `sql/schema` directory contains SQL scripts for database migrations.

## CI/CD Setup

This project uses GitHub Actions for Continuous Integration and Continuous Deployment (CI/CD) to automatically build, test, and deploy the application to Digital Ocean.

### Prerequisites

Before you can use the CI/CD pipeline, you need to set up the following:

1. **Docker Hub Account**: Create an account on [Docker Hub](https://hub.docker.com/) to store your Docker images.

2. **Digital Ocean Account**: Create an account on [Digital Ocean](https://www.digitalocean.com/) to host your application.

3. **GitHub Repository Secrets**: Add the following secrets to your GitHub repository:

   - **Docker Hub Credentials:**
     - `DOCKER_HUB_USERNAME`: Your Docker Hub username
     - `DOCKER_HUB_PASSWORD`: Your Docker Hub password or access token

   - **Digital Ocean Access:**
     - `DO_TOKEN`: Digital Ocean API token with write access
     - `DO_SSH_KEY_ID`: The numeric ID of your SSH key in Digital Ocean (not the fingerprint)
     - `DO_SSH_PRIVATE_KEY`: Private SSH key for accessing Digital Ocean droplets

   - **Database Configuration:**
     - `DB_USER`: Database username
     - `DB_PASSWORD`: Database password
     - `DB_NAME`: Database name

   - **AWS Configuration:**
     - `AWS_ACCESS_KEY_ID`: AWS access key ID for S3 access
     - `AWS_SECRET_ACCESS_KEY`: AWS secret access key
     - `AWS_REGION`: AWS region (e.g., us-east-1)
     - `AWS_BUCKET_NAME`: S3 bucket name for storage

   - **MPESA Configuration:**
     - `BUSINESS_SHORTCODE`: MPESA business shortcode
     - `PASSKEY`: MPESA passkey
     - `CALLBACK_URL`: Callback URL for MPESA transactions
     - `CONSUMER_KEY`: MPESA consumer key
     - `CONSUMER_SECRET`: MPESA consumer secret
     - `ACCOUNT_REFERENCE`: MPESA account reference

   - **SMTP Configuration:**
     - `SMTP_HOST`: SMTP server host
     - `SMTP_PORT`: SMTP server port
     - `SMTP_USERNAME`: SMTP username
     - `SMTP_PASSWORD`: SMTP password
     - `FROM_EMAIL`: Email sender address
     - `FROM_NAME`: Email sender name
     - `TO_EMAIL`: Default recipient email

   - **URL Configuration:**
     - `FRONTEND_URL`: URL for the frontend application
     - `BACKEND_URL`: URL for the backend application

   - **Redis Configuration:**
     - `REDIS_PASSWORD`: Redis password

   - **JWT Configuration:**
     - `JWT_ACCESS_SECRET`: Secret for access tokens
     - `JWT_REFRESH_SECRET`: Secret for refresh tokens
     - `JWT_VERIFY_SECRET`: Secret for verification tokens
     - `JWT_GUEST_SECRET`: Secret for guest tokens

   - **CORS Configuration:**
     - `ALLOWED_ORIGINS`: Comma-separated list of allowed origins for CORS

4. **Setup Script**: For convenience, you can use the included setup script to configure all these secrets:

   ```bash
   ./scripts/setup-cicd.sh
   ```

   This script will:
   - Load values from your local .env file if available
   - Prompt for any missing values
   - Set up all required GitHub secrets if GitHub CLI is installed
   - Provide instructions for manual setup if GitHub CLI is not available

### SSH Authentication Setup

For the CI/CD pipeline to work correctly, you need to set up SSH authentication:

#### Option 1: Key-based Authentication (Recommended)

To set up SSH key-based authentication:

1. Make sure your SSH private key is added as a GitHub secret named `DO_SSH_PRIVATE_KEY`
2. Run one of the helper scripts to add your public key to the server:

   ```bash
   # If you have the private key file locally:
   ./scripts/add-ssh-key-to-server.sh ~/.ssh/id_rsa gerald-bahati 209.97.128.72

   # Or if you want to use the key from GitHub secrets (requires GitHub CLI):
   ./scripts/add-github-ssh-key-to-server.sh gerald-bahati 209.97.128.72
   ```

#### Option 2: Password Authentication (Fallback)

If key-based authentication fails, the CI/CD pipeline will fall back to password authentication:

1. Add your server password as a GitHub secret named `DO_SSH_PASSWORD`
2. The password will be used with sudo commands during deployment

### Troubleshooting

If you encounter SSH authentication issues:

1. Check that your SSH key is correctly formatted
2. Verify the public key is in `~/.ssh/authorized_keys` on the server
3. Ensure the permissions are correct: `chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys`
4. Test the connection manually: `ssh -i /path/to/private_key gerald-bahati@209.97.128.72`

### How It Works

Our CI/CD pipeline follows these steps:

1. **Build Stage**:
   - Builds the Docker image for the application
   - Tags the image with both `latest` and the short Git commit SHA
   - Pushes the image to Docker Hub

2. **Deployment Stage**:
   - Creates or updates infrastructure on Digital Ocean using Terraform
   - Provisions and configures the `webline-backend` droplet
   - Copies the Docker Compose file and environment variables to the droplet
   - Pulls the latest image and starts the containers
   - Performs health checks to ensure the deployment is working correctly

### Finding Your SSH Key ID in Digital Ocean

To get your SSH key ID from Digital Ocean:

1. Log in to your Digital Ocean dashboard
2. Go to Settings → Security → SSH Keys
3. Find your SSH key in the list
4. The SSH key ID is the numeric identifier associated with your key

Alternatively, you can use the Digital Ocean CLI:
```bash
doctl compute ssh-key list
```

### Customizing the Deployment

You can customize the deployment by modifying the following files:

- `.github/workflows/ci-cd.yml`: Main workflow file for GitHub Actions
- `terraform/main.tf`: Terraform configuration for infrastructure (created by CI/CD)
- `terraform/variables.tf`: Variables for Terraform configuration (created by CI/CD)

### Manual Deployment

To trigger a deployment manually:

1. Go to your GitHub repository
2. Click on "Actions"
3. Select "CI/CD Pipeline"
4. Click "Run workflow"
5. Select the branch to deploy from
6. Click "Run workflow"

### Troubleshooting

If you encounter issues with the deployment:

1. Check the GitHub Actions logs for detailed error messages
2. Verify that all required secrets are correctly set
3. Ensure the health check endpoint (`/health`) is properly implemented in your application
4. Check the Digital Ocean droplet logs for application-specific issues

#### Common Issues

- **Docker Image Not Found**: If you're getting an error like `manifest for username/webline-backend:latest not found` when running locally, use the local development script: `./scripts/local-build.sh`. This script will build the image locally instead of pulling from Docker Hub.

- **SSH Key Issues**: Make sure you're using the numeric SSH key ID from Digital Ocean, not the fingerprint. You can find this ID in the Digital Ocean dashboard or by running `doctl compute ssh-key list`.

- **Environment Variables**: Make sure all required environment variables are set in your GitHub secrets for production deployment, and in your local `.env` file for development.

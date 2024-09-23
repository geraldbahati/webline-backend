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


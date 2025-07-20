# SSO Service - Authentication API Guide

A lightweight microservice designed for centralized authentication. As a core component of a modern, distributed system architecture, the SSO Service aims to streamline user authentication by providing a single point of entry for verifying user identities across multiple applications. Its microservice nature ensures high scalability, maintainability, and loose coupling, allowing for independent deployment and evolution.

## Table of Contents

1. Prerequisites
2. Setup and Run
3. API Endpoints
   * 3.1. Authentication
     * 3.1.1. Login (`POST /api/{version}/login`)
     * 3.1.2. Refresh Token (`POST /api/{version}/refresh`)
     * 3.1.3. Get My Profile (`GET /api/{version}/me`)
     * 3.1.4. Logout (`POST /api/{version}/logout`)
   * 3.2. User Management
     * 3.2.1. Create User (`POST /api/{version}/users`)
     * 3.2.2. Get All Users (`GET /api/{version}/users`)
     * 3.2.3. Get User by ID (`GET /api/{version}/users/:id`)
     * 3.2.4. Update User (`PUT /api/{version}/users/:id`)
     * 3.2.5. Delete User (`DELETE /api/{version}/users/:id`)
4. Roles and Permissions
5. Changelog

---

## 1. Prerequisites

Before you begin, ensure you have the following installed:

* **Go (Golang)**: Version 1.18 or higher.
* **PostgreSQL**: Database server.
* **`git`**: For cloning the repository (if applicable).
* **`curl`** or a tool like Postman/Insomnia for testing API endpoints.
---

## 2. Setup and Run

Follow these steps to get the SSO Service up and running:

1.  **Clone the repository** (if applicable, otherwise navigate to your project directory):

    ```bash
    git clone <your-repo-url>
    cd sso-service
    ```

2.  **Install Go Modules:**

    ```bash
    go mod tidy
    ```

3.  **Configure Environment Variables:**
    Create a `.env` file in the root directory of your project (`sso-service/`) with the following content. **Remember to replace placeholder values with your actual database credentials and a strong JWT secret.**

    ```dotenv
    API_VERSION=v1
    PORT=8080
    JWT_SECRET=your_super_secret_jwt_key_here_change_this_in_production

    # PostgreSQL Database Configuration
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=sso_user
    DB_PASSWORD=sso_password
    DB_NAME=sso_db
    DB_SSLMODE=disable # Use 'require' or 'verify-full' for production with SSL
    ```

4.  **Prepare PostgreSQL Database:**

    * Ensure your PostgreSQL server is running.
    * Create the database and user as specified in your `.env` file (e.g., `sso_db` and `sso_user`).
    * Connect to your PostgreSQL database and run the SQL schema commands to create all necessary tables. These commands were provided by the AI assistant in previous responses (look for the 'SSO Database SQL Schema' code block).
        * **Important:** You might need to run `CREATE EXTENSION IF NOT EXISTS "pgcrypto";` first.
        * **Important:** If you're updating an existing database, consider using `DROP TABLE IF EXISTS ... CASCADE;` commands before `CREATE TABLE` to ensure a clean slate, but **be cautious as this will delete all existing data.**

5.  **Insert Dummy Data:**

    * Generate bcrypt hashes for your dummy user passwords. You can use a simple Go script for this (e.g., `generate_hashes.go` from previous instructions).

        ```go
        // generate_hashes.go
        package main
        import (
        	"fmt"
        	"log"
        	"golang.org/x/crypto/bcrypt"
        )
        func main() {
        	passwordAdmin := "adminpassword"
        	passwordUser := "userpassword"
        	hashedAdminPassword, err := bcrypt.GenerateFromPassword([]byte(passwordAdmin), bcrypt.DefaultCost)
        	if err != nil { log.Fatalf("Error hashing admin password: %v", err) }
        	fmt.Printf("Hashed Admin Password for '%s': %s\n", passwordAdmin, string(hashedAdminPassword))
        	hashedUserPassword, err := bcrypt.GenerateFromPassword([]byte(passwordUser), bcrypt.DefaultCost)
        	if err != nil { log.Fatalf("Error hashing user password: %v", err) }
        	fmt.Printf("Hashed User Password for '%s': %s\n", passwordUser, string(hashedUserPassword))
        }
        ```

        Run this script: `go run generate_hashes.go` and copy the generated hashes.

    * Insert dummy data (users, roles, permissions, sites, user_roles, role_permissions) into your PostgreSQL database using the SQL `INSERT` statements provided previously. **Make sure to replace the password hash placeholders with the actual hashes you generated.**

6.  **Run the Application:**

    ```bash
    go run main.go
    ```

    The server should start and listen on the port specified in your `.env` file (default: `8080`).

---

## 3. API Endpoints

The API version is dynamically loaded from the `API_VERSION` environment variable (default: `v1`). All endpoints will be prefixed with `/api/{version}/`.

### 3.1. Authentication

#### 3.1.1. Login (`POST /api/{version}/login`)

Authenticates a user and returns an access token.

* **URL:** `http://localhost:8080/api/v1/login` (replace `v1` with your `API_VERSION`)
* **Method:** `POST`
* **Headers:**
    * `Content-Type: application/json`
* **Request Body (JSON):**
    ```json
    {
        "username": "admin",
        "password": "adminpassword"
    }
    ```
    *(Use `superadmin`/`superadminpassword`, `admin`/`adminpassword`, `support`/`supportpassword`, or `consumer`/`consumerpassword` from your dummy data)*
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": {
            "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "token_type": "Bearer",
            "expires_in": 900
        },
        "message": "Login successful"
    }
    ```
* **Error Response (400 Bad Request / 401 Unauthorized):**
    ```json
    {
        "status": 400,
        "data": null,
        "message": "Error message details (e.g., 'username and password cannot be empty', 'invalid credentials')"
    }
    ```
* **`curl` Example:**
    ```bash
    curl -X POST \
      http://localhost:8080/api/v1/login \
      -H 'Content-Type: application/json' \
      -d '{
        "username": "admin",
        "password": "adminpassword"
      }'
    ```

#### 3.1.2. Refresh Token (`POST /api/{version}/refresh`)

Obtains a new access token and refresh token using an existing refresh token. This endpoint is designed for mobile applications or long-lived sessions where the refresh token is securely stored.

* **URL:** `http://localhost:8080/api/v1/refresh` (replace `v1` with your `API_VERSION`)
* **Method:** `POST`
* **Headers:**
    * `Authorization: Bearer <YOUR_REFRESH_TOKEN_HERE>`
* **Request Body:** (No body required, refresh token is in header)
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": {
            "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        },
        "message": "Token refreshed successfully"
    }
    ```
* **Error Response (401 Unauthorized):**
    ```json
    {
        "status": 401,
        "data": null,
        "message": "Error message details (e.g., 'Authorization header required', 'Invalid refresh token')"
    }
    ```
* **`curl` Example:**
    ```bash
    # First, get a refresh token from the /login endpoint
    # Then, use it here:
    curl -X POST \
      http://localhost:8080/api/v1/refresh \
      -H 'Authorization: Bearer <YOUR_REFRESH_TOKEN_HERE>'
    ```

#### 3.1.3. Get My Profile (`GET /api/{version}/me`)

Retrieves the profile, roles, and hierarchical permissions of the currently authenticated user. This endpoint requires a valid access token.

* **URL:** `http://localhost:8080/api/v1/me` (replace `v1` with your `API_VERSION`)
* **Method:** `GET`
* **Headers:**
    * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`
* **Request Body:** (No body required)
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": {
            "profile": {
                "id": "11eebc99-9c0b-4ef8-bb6d-6bb9bd380a77",
                "username": "admin",
                "email": "admin@example.com",
                "is_active": true,
                "created_at": "2023-10-27T10:00:00Z",
                "updated_at": "2023-10-27T10:00:00Z"
            },
            "roles": [
                "admin"
            ],
            "permissions": [
                "cms:admin",
                "cms:admin:users:manage",
                "cms:admin:content:manage",
                "mobile:read"
            ]
        },
        "message": "User profile retrieved successfully"
    }
    ```
* **Error Response (401 Unauthorized / 404 Not Found / 500 Internal Server Error):**
    ```json
    {
        "status": 401,
        "data": null,
        "message": "Error message details (e.g., 'Invalid or expired token', 'User ID not found in context')"
    }
    ```
* **`curl` Example:**
    ```bash
    # First, get an access token from the /login endpoint
    # Then, use it here:
    curl -X GET \
      http://localhost:8080/api/v1/me \
      -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>'
    ```

#### 3.1.4. Logout (`POST /api/{version}/logout`)

Invalidates the current access token, effectively ending the user's session. The token's JTI (JWT ID) is blacklisted.

* **URL:** `http://localhost:8080/api/v1/logout` (replace `v1` with your `API_VERSION`)
* **Method:** `POST`
* **Headers:**
    * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`
* **Request Body:** (No body required)
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": null,
        "message": "Logged out successfully"
    }
    ```
* **Error Response (401 Unauthorized / 500 Internal Server Error):**
    ```json
    {
        "status": 401,
        "data": null,
        "message": "Error message details (e.g., 'Authorization header required', 'Failed to invalidate token')"
    }
    ```
* **`curl` Example:**
    ```bash
    # First, get an access token from the /login endpoint
    # Then, use it here:
    curl -X POST \
      http://localhost:8080/api/v1/logout \
      -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>'
    ```

### 3.2. User Management

All User Management endpoints require a valid access token in the `Authorization: Bearer` header. Specific permission requirements are noted below.

#### 3.2.1. Create User (`POST /api/{version}/users`)

Creates a new user in the system.

* **URL:** `http://localhost:8080/api/v1/users`
* **Method:** `POST`
* **Headers:**
    * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`
    * `Content-Type: application/json`
* **Required Permissions:** `user:create`
* **Request Body (JSON):**
    ```json
    {
        "username": "newuser",
        "email": "newuser@example.com",
        "password": "securepassword123",
        "is_active": true
    }
    ```
    * `is_active`: Optional. If not provided, defaults to `true`.
* **Success Response (201 Created):**
    ```json
    {
        "status": 201,
        "data": {
            "id": "a1b2c3d4-e5f6-7890-1234-567890abcdef",
            "username": "newuser",
            "email": "newuser@example.com",
            "is_active": true,
            "created_at": "2023-10-27T10:30:00Z",
            "updated_at": "2023-10-27T10:30:00Z"
        },
        "message": "User created successfully"
    }
    ```
* **Error Response (400 Bad Request / 403 Forbidden / 409 Conflict / 500 Internal Server Error):**
    ```json
    {
        "status": 403,
        "data": null,
        "message": "Forbidden: insufficient permissions"
    }
    ```
* **`curl` Example:**
    ```bash
    curl -X POST \
      http://localhost:8080/api/v1/users \
      -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>' \
      -H 'Content-Type: application/json' \
      -d '{
        "username": "newuser",
        "email": "newuser@example.com",
        "password": "securepassword123",
        "is_active": true
      }'
    ```

#### 3.2.2. Get All Users (`GET /api/{version}/users`)

Retrieves a list of all users in the system.

* **URL:** `http://localhost:8080/api/v1/users`
* **Method:** `GET`
* **Headers:**
    * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`
* **Required Permissions:** `user:read_all`
* **Request Body:** (No body required)
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": [
            {
                "id": "11eebc99-9c0b-4ef8-bb6d-6bb9bd380a77",
                "username": "superadmin",
                "email": "superadmin@example.com",
                "is_active": true,
                "created_at": "2023-10-27T10:00:00Z",
                "updated_at": "2023-10-27T10:00:00Z"
            },
            {
                "id": "21eebc99-9c0b-4ef8-bb6d-6bb9bd380a88",
                "username": "admin",
                "email": "admin@example.com",
                "is_active": true,
                "created_at": "2023-10-27T10:00:00Z",
                "updated_at": "2023-10-27T10:00:00Z"
            }
        ],
        "message": "Users retrieved successfully"
    }
    ```
* **Error Response (403 Forbidden / 500 Internal Server Error):**
    ```json
    {
        "status": 403,
        "data": null,
        "message": "Forbidden: insufficient permissions"
    }
    ```
* **`curl` Example:**
    ```bash
    curl -X GET \
      http://localhost:8080/api/v1/users \
      -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>'
    ```

#### 3.2.3. Get User by ID (`GET /api/{version}/users/:id`)

Retrieves a specific user's profile by their ID.

* **URL:** `http://localhost:8080/api/v1/users/{user_id}` (replace `{user_id}` with the user's UUID ID)
* **Method:** `GET`
* **Headers:**
    * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`
* **Required Permissions:** `user:read_all` OR `user:read_own` (if retrieving their own profile)
* **Request Body:** (No body required)
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": {
            "id": "11eebc99-9c0b-4ef8-bb6d-6bb9bd380a77",
            "username": "superadmin",
            "email": "superadmin@example.com",
            "is_active": true,
            "created_at": "2023-10-27T10:00:00Z",
            "updated_at": "2023-10-27T10:00:00Z"
        },
        "message": "User retrieved successfully"
    }
    ```
* **Error Response (400 Bad Request / 403 Forbidden / 404 Not Found / 500 Internal Server Error):**
    ```json
    {
        "status": 403,
        "data": null,
        "message": "Forbidden: insufficient permissions to read this user's profile"
    }
    ```
* **`curl` Example:**
    ```bash
    curl -X GET \
      http://localhost:8080/api/v1/users/11eebc99-9c0b-4ef8-bb6d-6bb9bd380a77 \
      -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>'
    ```

#### 3.2.4. Update User (`PUT /api/{version}/users/:id`)

Updates an existing user's profile.

* **URL:** `http://localhost:8080/api/v1/users/{user_id}` (replace `{user_id}` with the user's UUID ID)
* **Method:** `PUT`
* **Headers:**
    * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`
    * `Content-Type: application/json`
* **Required Permissions:** `user:update_all` OR `user:update_own` (if updating their own profile)
    * **Note:** `superadmin` can update anything. `admin` cannot update `superadmin`. `consumer` can only update their own password.
* **Request Body (JSON):**
    ```json
    {
        "username": "updateduser",
        "email": "updated@example.com",
        "password": "newsecurepassword",
        "is_active": false
    }
    ```
    * All fields are optional for partial updates.
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": {
            "id": "a1b2c3d4-e5f6-7890-1234-567890abcdef",
            "username": "updateduser",
            "email": "updated@example.com",
            "is_active": false,
            "created_at": "2023-10-27T10:30:00Z",
            "updated_at": "2023-10-27T10:45:00Z"
        },
        "message": "User updated successfully"
    }
    ```
* **Error Response (400 Bad Request / 403 Forbidden / 404 Not Found / 409 Conflict / 500 Internal Server Error):**
    ```json
    {
        "status": 403,
        "data": null,
        "message": "Forbidden: insufficient permissions to update this user's profile"
    }
    ```
* **`curl` Example:**
    ```bash
    curl -X PUT \
      http://localhost:8080/api/v1/users/11eebc99-9c0b-4ef8-bb6d-6bb9bd380a77 \
      -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>' \
      -H 'Content-Type: application/json' \
      -d '{
        "username": "updatedadmin",
        "is_active": false
      }'
    ```

#### 3.2.5. Delete User (`DELETE /api/{version}/users/:id`)

Deletes a user from the system.

* **URL:** `http://localhost:8080/api/v1/users/{user_id}` (replace `{user_id}` with the user's UUID ID)
* **Method:** `DELETE`
* **Headers:**
    * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`
* **Required Permissions:** `user:delete`
    * **Note:** `superadmin` can delete any user. `admin` cannot delete `superadmin`. Users cannot delete themselves via this endpoint.
* **Request Body:** (No body required)
* **Success Response (200 OK):**
    ```json
    {
        "status": 200,
        "data": null,
        "message": "User deleted successfully"
    }
    ```
* **Error Response (403 Forbidden / 404 Not Found / 500 Internal Server Error):**
    ```json
    {
        "status": 403,
        "data": null,
        "message": "Forbidden: insufficient permissions to delete users"
    }
    ```
* **`curl` Example:**
    ```bash
    curl -X DELETE \
      http://localhost:8080/api/v1/users/a1b2c3d4-e5f6-7890-1234-567890abcdef \
      -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>'
    ```


## 4. Changelog

### Version 0.1
* Initial release of the Authentication API.
* Implemented user login with JWT access and refresh tokens.
* Implemented token refresh functionality.
* Implemented user profile retrieval (`/me`) with authenticated access.
* Implemented token invalidation (logout) using a JTI blacklist.
* Integrated with PostgreSQL database for user, role, permission, and site management.
* Supports hierarchical permissions.
* Standardized JSON response format (`status`, `data`, `message`).

### Version 0.2
* Implemented User Managements (`Create`, `List`, `Update`, `Delete`)
* Split Middleware as Standalone Configuration
* Standardized JSON response format (`status`, `data`, `message`).
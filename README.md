# SSO Service - Authentication API Guide

A lightweight microservice designed for centralized authentication. As a core component of a modern, distributed system architecture, the SSO Service aims to streamline user authentication by providing a single point of entry for verifying user identities across multiple applications. Its microservice nature ensures high scalability, maintainability, and loose coupling, allowing for independent deployment and evolution.

## Table of Contents

1. Prerequisites
2. Setup and Run
3. API Endpoints
   * 3.1. Login (`POST /api/{version}/login`)
   * 3.2. Refresh Token (`POST /api/{version}/refresh`)
   * 3.3. Get My Profile (`GET /api/{version}/me`)
   * 3.4. Logout (`POST /api/{version}/logout`)
4. Changelog

## 1. Prerequisites

Before you begin, ensure you have the following installed:

* **Go (Golang)**: Version 1.18 or higher.
* **PostgreSQL**: Database server.
* **`git`**: For cloning the repository (if applicable).
* **`curl`** or a tool like Postman/Insomnia for testing API endpoints.

## 2. Setup and Run

Follow these steps to get the SSO Service up and running:

1. **Clone the repository** (if applicable, otherwise navigate to your project directory):

   ```
   git clone <your-repo-url>
   cd sso-service
   ```

2. **Install Go Modules:**

   ```
   go mod tidy
   ```

3. **Configure Environment Variables:**
   Create a `.env` file in the root directory of your project (`sso-service/`) with the following content. **Remember to replace placeholder values with your actual database credentials and a strong JWT secret.**

   ```
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

4. **Prepare PostgreSQL Database:**

   * Ensure your PostgreSQL server is running.

   * Create the database and user as specified in your `.env` file (e.g., `sso_db` and `sso_user`).

   * Connect to your PostgreSQL database and run the SQL schema commands to create all necessary tables. These commands were provided by the AI assistant in previous responses (e.g., look for the 'SSO Database SQL Schema' code block).

     * **Important:** You might need to run `CREATE EXTENSION IF NOT EXISTS "pgcrypto";` first.

     * **Important:** If you're updating an existing database, consider using `DROP TABLE IF EXISTS ... CASCADE;` commands before `CREATE TABLE` to ensure a clean slate, but **be cautious as this will delete all existing data.**

5. **Insert Dummy Data:**

   * Generate bcrypt hashes for your dummy user passwords. You can use a simple Go script for this (e.g., `generate_hashes.go` from previous instructions).

     ```
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

6. **Run the Application:**

   ```
   go run main.go
   ```

   The server should start and listen on the port specified in your `.env` file (default: `8080`).

## 3. API Endpoints

The API version is dynamically loaded from the `API_VERSION` environment variable (default: `v1`). All endpoints will be prefixed with `/api/{version}/`.

### 3.1. Login (`POST /api/{version}/login`)

Authenticates a user and returns an access token.

* **URL:** `http://localhost:8080/api/v1/login` (replace `v1` with your `API_VERSION`)

* **Method:** `POST`

* **Headers:**

  * `Content-Type: application/json`

* **Request Body (JSON):**

  ```
  {
      "username": "admin",
      "password": "adminpassword"
  }
  ```

  *(Use `admin`/`adminpassword` or `testuser`/`userpassword` from your dummy data)*

* **Success Response (200 OK):**

  ```
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

  ```
  {
      "status": 400,
      "data": null,
      "message": "Error message details (e.g., 'username and password cannot be empty', 'invalid credentials')"
  }
  ```

* **`curl` Example:**

  ```
  curl -X POST \
    http://localhost:8080/api/v1/login \
    -H 'Content-Type: application/json' \
    -d '{
      "username": "admin",
      "password": "adminpassword"
    }'
  ```

### 3.2. Refresh Token (`POST /api/{version}/refresh`)

Obtains a new access token and refresh token using an existing refresh token. This endpoint is designed for mobile applications or long-lived sessions where the refresh token is securely stored.

* **URL:** `http://localhost:8080/api/v1/refresh` (replace `v1` with your `API_VERSION`)

* **Method:** `POST`

* **Headers:**

  * `Authorization: Bearer <YOUR_REFRESH_TOKEN_HERE>`

* **Request Body:** (No body required, refresh token is in header)

* **Success Response (200 OK):**

  ```
  {
      "status": 200,
      "data": {
          "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
      },
      "message": "Token refreshed successfully"
  }
  ```

* **Error Response (401 Unauthorized):**

  ```
  {
      "status": 401,
      "data": null,
      "message": "Error message details (e.g., 'Authorization header required', 'Invalid refresh token')"
  }
  ```

* **`curl` Example:**

  ```
  # First, get a refresh token from the /login endpoint
  # Then, use it here:
  curl -X POST \
    http://localhost:8080/api/v1/refresh \
    -H 'Authorization: Bearer <YOUR_REFRESH_TOKEN_HERE>'
  ```

### 3.3. Get My Profile (`GET /api/{version}/me`)

Retrieves the profile, roles, and hierarchical permissions of the currently authenticated user. This endpoint requires a valid access token.

* **URL:** `http://localhost:8080/api/v1/me` (replace `v1` with your `API_VERSION`)

* **Method:** `GET`

* **Headers:**

  * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`

* **Request Body:** (No body required)

* **Success Response (200 OK):**

  ```
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

  ```
  {
      "status": 401,
      "data": null,
      "message": "Error message details (e.g., 'Invalid or expired token', 'User ID not found in context')"
  }
  ```

* **`curl` Example:**

  ```
  # First, get an access token from the /login endpoint
  # Then, use it here:
  curl -X GET \
    http://localhost:8080/api/v1/me \
    -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>'
  ```

### 3.4. Logout (`POST /api/{version}/logout`)

Invalidates the current access token, effectively ending the user's session. The token's JTI (JWT ID) is blacklisted.

* **URL:** `http://localhost:8080/api/v1/logout` (replace `v1` with your `API_VERSION`)

* **Method:** `POST`

* **Headers:**

  * `Authorization: Bearer <YOUR_ACCESS_TOKEN_HERE>`

* **Request Body:** (No body required)

* **Success Response (200 OK):**

  ```
  {
      "status": 200,
      "data": null,
      "message": "Logged out successfully"
  }
  ```

* **Error Response (401 Unauthorized / 500 Internal Server Error):**

  ```
  {
      "status": 401,
      "data": null,
      "message": "Error message details (e.g., 'Authorization header required', 'Failed to invalidate token')"
  }
  ```

* **`curl` Example:**

  ```
  # First, get an access token from the /login endpoint
  # Then, use it here:
  curl -X POST \
    http://localhost:8080/api/v1/logout \
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

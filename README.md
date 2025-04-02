# Film API

A RESTful API service for managing a film database, built with Go. This API provides endpoints for creating, reading, updating, and deleting film records, along with user authentication and permission management.

## Features

- **Film Management**: Full CRUD operations for films
- **Rich Film Data**: Support for genres, directors, actors, and ratings
- **User Authentication**: Secure user registration and authentication
- **Permission-based Access**: Role-based access control for API endpoints
- **Pagination & Filtering**: Advanced query options for film listings
- **CORS Support**: Configurable Cross-Origin Resource Sharing
- **Rate Limiting**: Customizable rate limiting for API endpoints

## Technical Stack

- **Language**: Go
- **Database**: PostgreSQL
- **Authentication**: Token-based authentication
- **Documentation**: OpenAPI/Swagger

## Getting Started

### Prerequisites

- Go (latest version)
- PostgreSQL
- Docker (optional)

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd film_api
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up the environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

### Database Setup

1. Create a PostgreSQL database
2. Run the migrations:
   ```bash
   # Using golang-migrate
   migrate -path ./migrations -database "postgres://your-connection-string" up
   ```

## API Documentation

### Authentication

All protected endpoints require a valid authentication token in the Authorization header:
```bash
Authorization: Bearer <your-token>
```

### Endpoints

#### Films

##### List Films
```http
GET /films
```

Query Parameters:
- `page` (int): Page number (default: 1)
- `page_size` (int): Results per page (default: 20)
- `title` (string): Filter by title
- `genres` (string): Filter by genres (comma-separated)
- `directors` (string): Filter by directors (comma-separated)
- `actors` (string): Filter by actors (comma-separated)
- `sort` (string): Sort field (-field for descending)

Example Response:
```json
{
  "films": [
    {
      "id": 1,
      "title": "Inception",
      "year": 2010,
      "runtime": "148 mins",
      "rating": 8.8,
      "description": "A mind-bending thriller",
      "image": "http://example.com/inception.jpg",
      "version": 1,
      "genres": ["Sci-Fi", "Thriller"],
      "directors": ["Christopher Nolan"],
      "actors": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"]
    }
  ],
  "metadata": {
    "current_page": 1,
    "page_size": 20,
    "first_page": 1,
    "last_page": 1,
    "total_records": 1
  }
}
```

#### User Management

##### Register User
```http
POST /users
```

Request Body:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "your-secure-password"
}
```

Response:
```json
{
  "user": {
    "id": 1,
    "created_at": "2024-04-02T14:30:00Z",
    "name": "John Doe",
    "email": "john@example.com",
    "activated": false
  },
  "activation_token": {
    "token": "ACTIVATION-TOKEN",
    "expiry": "2024-04-03T14:30:00Z"
  }
}
```

##### Activate User
```http
PUT /users/activate
```

Request Body:
```json
{
  "token": "ACTIVATION-TOKEN"
}
```

Response:
```json
{
  "user": {
    "id": 1,
    "created_at": "2024-04-02T14:30:00Z",
    "name": "John Doe",
    "email": "john@example.com",
    "activated": true
  }
}
```

##### Authentication
```http
POST /tokens/authentication
```

Request Body:
```json
{
  "email": "john@example.com",
  "password": "your-secure-password"
}
```

Response:
```json
{
  "authentication_token": {
    "token": "YOUR-AUTH-TOKEN",
    "expiry": "2024-04-03T14:30:00Z"
  }
}
```

#### Films (Protected Endpoints)

##### Create Film
```http
POST /films
Authorization: Bearer YOUR-AUTH-TOKEN
```

Request Body:
```json
{
  "title": "The Matrix",
  "year": 1999,
  "runtime": 136,
  "rating": 8.7,
  "description": "A computer hacker learns about the true nature of reality",
  "image": "http://example.com/matrix.jpg",
  "genres": ["Action", "Sci-Fi"],
  "directors": ["Lana Wachowski", "Lilly Wachowski"],
  "actors": ["Keanu Reeves", "Laurence Fishburne"]
}
```

Response:
```json
{
  "film": {
    "id": 2,
    "title": "The Matrix",
    "year": 1999,
    "runtime": "136 mins",
    "rating": 8.7,
    "description": "A computer hacker learns about the true nature of reality",
    "image": "http://example.com/matrix.jpg",
    "version": 1,
    "genres": ["Action", "Sci-Fi"],
    "directors": ["Lana Wachowski", "Lilly Wachowski"],
    "actors": ["Keanu Reeves", "Laurence Fishburne"]
  }
}
```

##### Update Film
```http
PUT /films/{id}
Authorization: Bearer YOUR-AUTH-TOKEN
```

Request Body:
```json
{
  "title": "The Matrix",
  "year": 1999,
  "runtime": 136,
  "rating": 9.0,
  "description": "Updated description",
  "image": "http://example.com/matrix.jpg",
  "genres": ["Action", "Sci-Fi"],
  "directors": ["Lana Wachowski", "Lilly Wachowski"],
  "actors": ["Keanu Reeves", "Laurence Fishburne"]
}
```

Response:
```json
{
  "film": {
    "id": 2,
    "title": "The Matrix",
    "year": 1999,
    "runtime": "136 mins",
    "rating": 9.0,
    "description": "Updated description",
    "image": "http://example.com/matrix.jpg",
    "version": 2,
    "genres": ["Action", "Sci-Fi"],
    "directors": ["Lana Wachowski", "Lilly Wachowski"],
    "actors": ["Keanu Reeves", "Laurence Fishburne"]
  }
}
```

##### Delete Film
```http
DELETE /films/{id}
Authorization: Bearer YOUR-AUTH-TOKEN
```

Response:
```json
{
  "message": "movie deleted successfully"
}
```

### Filtering and Pagination

The films listing endpoint (`GET /films`) supports various query parameters for filtering and pagination:

```http
GET /films?page=1&page_size=20&title=matrix&genres=action,sci-fi&directors=nolan&actors=keanu
```

Query Parameters:
- `page`: Page number (default: 1)
- `page_size`: Number of results per page (default: 20)
- `title`: Search by title (case-insensitive, partial match)
- `genres`: Filter by genres (comma-separated)
- `directors`: Filter by directors (comma-separated)
- `actors`: Filter by actors (comma-separated)
- `sort`: Sort results by field (prefix with - for descending order)
  - Allowed fields: id, title, year, runtime, rating

### Permissions

The API implements role-based access control with the following permissions:
- `films:read`: Required for viewing film details
- `films:write`: Required for creating, updating, and deleting films

These permissions are automatically assigned upon user activation and authentication.

## Error Handling

The API uses conventional HTTP response codes to indicate the success or failure of requests:

- `200 OK`: Successful request
- `201 Created`: Resource successfully created
- `400 Bad Request`: Invalid request (e.g., invalid parameters)
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Authenticated but not authorized
- `404 Not Found`: Resource not found
- `405 Method Not Allowed`: Invalid HTTP method
- `500 Internal Server Error`: Server error

Error Response Format:
```json
{
  "error": "Detailed error message"
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse. Limits can be configured via environment variables:

- `LIMITER_RPS`: Requests per second
- `LIMITER_BURST`: Maximum burst size
- `LIMITER_ENABLED`: Enable/disable rate limiting

## Development

### Running Locally

```bash
go run ./cmd/api
```

### Using Docker

```bash
docker-compose up --build
```

### Running Tests

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## User Management

The application includes functionality for managing users, including creating, retrieving, updating, and authenticating users.

### User Model
- **User Struct**: Represents a user with fields such as `ID`, `Name`, `Email`, `Password`, `Activated`, and `Version`.
- **Password Management**: Uses bcrypt for hashing passwords, with methods for setting and verifying passwords.

### User Operations
- **Insert**: Adds a new user to the database.
- **GetByEmail**: Retrieves a user by their email address.
- **Update**: Updates user details in the database.
- **GetForToken**: Retrieves a user based on a token, useful for authentication.

### Validation
- **Email Validation**: Ensures the email is provided and matches a valid format.
- **Password Validation**: Checks that the password is provided and meets length requirements.
- **User Validation**: Validates the user's name, email, and password.

### Error Handling
- Handles errors such as duplicate emails and record not found scenarios.

### Anonymous User
- Provides a concept of an anonymous user for cases where user authentication is not present.

## HTTP Response Handling

The application uses structured methods to handle HTTP responses, ensuring consistent status codes and error messages.

### Status Codes
- **200 OK**: Used for successful requests, such as retrieving or updating resources.
- **201 Created**: Used when a new resource is successfully created.
- **204 No Content**: Used when a resource is successfully deleted.
- **400 Bad Request**: Used when the request is malformed or contains invalid data.
- **403 Forbidden**: Used when access is denied, such as when CORS is not allowed.
- **404 Not Found**: Used when a requested resource cannot be found.
- **405 Method Not Allowed**: Used when an HTTP method is not supported for a resource.
- **500 Internal Server Error**: Used when the server encounters an unexpected condition.

### Error Handling
- The application provides helper functions to send error responses with appropriate status codes and messages.
- Common error responses include `serverErrorResponse`, `notFoundResponse`, `methodNotAllowedResponse`, and `badRequestResponse`.
- The `errorResponse` function is used to send custom error messages with a specified status code.

## API Usage Examples

Here are some examples of how to interact with the Film API using `curl` commands.

### Get List of Films

**Request:**

```bash
curl -i -H 'Accept: application/json' http://localhost:4000/films
```

**Response:**

```
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 123

{
  "films": [
    {
      "id": 1,
      "title": "Inception",
      "year": 2010,
      "runtime": "148 mins",
      "rating": 8.8,
      "description": "A mind-bending thriller",
      "image": "http://example.com/inception.jpg",
      "version": 1,
      "genres": ["Sci-Fi", "Thriller"],
      "directors": ["Christopher Nolan"],
      "actors": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"]
    }
  ],
  "metadata": {
    "current_page": 1,
    "page_size": 20,
    "first_page": 1,
    "last_page": 1,
    "total_records": 1
  }
}
```

### Create a New Film

**Request:**

```bash
curl -i -H 'Accept: application/json' -H 'Content-Type: application/json' -X POST -d '{
  "title": "The Matrix",
  "year": 1999,
  "runtime": 136,
  "rating": 8.7,
  "description": "A computer hacker learns about the true nature of reality",
  "image": "http://example.com/matrix.jpg",
  "genres": ["Action", "Sci-Fi"],
  "directors": ["Lana Wachowski", "Lilly Wachowski"],
  "actors": ["Keanu Reeves", "Laurence Fishburne"]
}' http://localhost:4000/films
```

**Response:**

```
HTTP/1.1 201 Created
Content-Type: application/json
Location: /v1/films/2
Content-Length: 123

{
  "film": {
    "id": 2,
    "title": "The Matrix",
    "year": 1999,
    "runtime": "136 mins",
    "rating": 8.7,
    "description": "A computer hacker learns about the true nature of reality",
    "image": "http://example.com/matrix.jpg",
    "version": 1,
    "genres": ["Action", "Sci-Fi"],
    "directors": ["Lana Wachowski", "Lilly Wachowski"],
    "actors": ["Keanu Reeves", "Laurence Fishburne"]
  }
}
```

### Get a Specific Film

**Request:**

```bash
curl -i -H 'Accept: application/json' http://localhost:4000/films/1
```

**Response:**

```
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 123

{
  "film": {
    "id": 1,
    "title": "Inception",
    "year": 2010,
    "runtime": "148 mins",
    "rating": 8.8,
    "description": "A mind-bending thriller",
    "image": "http://example.com/inception.jpg",
    "version": 1,
    "genres": ["Sci-Fi", "Thriller"],
    "directors": ["Christopher Nolan"],
    "actors": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"]
  }
}
```

### Update a Film

**Request:**

```bash
curl -i -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH -d '{
  "title": "Inception",
  "year": 2010,
  "runtime": 148,
  "rating": 9.0,
  "description": "A mind-bending thriller with a new rating",
  "image": "http://example.com/inception.jpg",
  "genres": ["Sci-Fi", "Thriller"],
  "directors": ["Christopher Nolan"],
  "actors": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"]
}' http://localhost:4000/films/1
```

**Response:**

```
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 123

{
  "film": {
    "id": 1,
    "title": "Inception",
    "year": 2010,
    "runtime": "148 mins",
    "rating": 9.0,
    "description": "A mind-bending thriller with a new rating",
    "image": "http://example.com/inception.jpg",
    "version": 2,
    "genres": ["Sci-Fi", "Thriller"],
    "directors": ["Christopher Nolan"],
    "actors": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"]
  }
}
```

### Delete a Film

**Request:**

```bash
curl -i -H 'Accept: application/json' -X DELETE http://localhost:4000/films/1
```

**Response:**

```
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 41

{
    "message": "movie deleted succesfully"
}
``` 
# Film API 🎬

A high-performance RESTful API for film discovery and recommendations, built with **Go** and **Neo4j** graph database. Features intelligent multi-factor recommendations, JWT authentication, and comprehensive film management.

## ✨ Key Features

### 🤖 Intelligent Recommendation Engine
- **Multi-dimensional scoring** algorithm analyzing 6 factors
- **Weighted matching**: Directors (5×), Actors (3×), Genres (1×)
- **Attribute similarity**: Certificate, Year range, Runtime
- **Personalized**: Scores multiplied by your ratings/priority
- **Graph-powered**: Leverages Neo4j relationships for fast traversal

### 🔐 Modern Authentication
- **JWT tokens**: Industry-standard, stateless authentication
- **Access tokens**: 1-hour lifetime for API calls
- **Refresh tokens**: 7-day lifetime with revocation capability
- **Zero database overhead**: Local JWT verification (10-50ms faster!)
- **Django-ready**: Standard format for easy integration

### 📊 Rich Film Database
- **8000+ films** auto-populated from IMDb dataset
- **Comprehensive metadata**: Genres, Directors, Actors, Ratings, Certificates
- **Graph relationships**: Efficient queries via Neo4j
- **Full-text search**: Filter by title, cast, crew
- **Advanced filtering**: Multi-criteria search with pagination

### 📝 Personal Watchlists
- **Priority system**: Rate films 1-10 for watch priority
- **Rating system**: Rate watched films to improve recommendations  
- **Watch tracking**: Mark films as watched with timestamps
- **Notes**: Add personal comments to films
- **User isolation**: Private, secure watchlists

### ⚡ Performance & Scalability
- **Stateless JWT**: No database queries for authentication
- **Graph database**: Optimized for relationship queries
- **Rate limiting**: Configurable protection against abuse
- **CORS support**: Secure cross-origin requests
- **Efficient pagination**: Handle large datasets smoothly

---

## 🛠 Technical Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.23+ |
| **Database** | Neo4j (Graph Database) |
| **Authentication** | JWT (golang-jwt/jwt/v5) |
| **Password Hashing** | bcrypt |
| **Rate Limiting** | golang.org/x/time/rate |
| **Containerisation** | Docker + Compose (scratch-based image) |
| **Architecture** | RESTful API with middleware chain |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or later
- Neo4j Database (Aura or local instance)

### Installation

```bash
# Clone repository
git clone <repository-url>
cd film_api

# Install dependencies
go mod tidy

# Configure environment
cp .env.example .env
# Edit .env with your Neo4j credentials and JWT secret
```

### Configuration

Create `.env` file:
```bash
# Neo4j
NEO4J_URI="neo4j+s://your-instance.databases.neo4j.io"
NEO4J_USERNAME="neo4j"
NEO4J_PASSWORD="your-password"

# JWT
JWT_SECRET="your-super-secret-key-min-32-chars"

# Server
APP_PORT=4000
APP_ENV=development
```

**Generate secure JWT secret:**
```bash
openssl rand -base64 32
```

### Run

```bash
# Build
go build -o api ./cmd/api

# Run
./api
```

Server starts on `http://localhost:4000`

---

## 🐳 Docker Self-Hosting

Run the entire stack (API + Neo4j) with a single command. The final API image is built from `scratch` — **under 15 MB**.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (v20+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2+)

### Setup

```bash
# 1. Clone & enter the project
git clone <repository-url>
cd film_api

# 2. Create your environment file
cp .env.example .env

# 3. Edit .env — at minimum change these:
#    NEO4J_PASSWORD=<choose-a-strong-password>
#    JWT_SECRET=<run: openssl rand -base64 32>

# 4. Launch everything
docker compose up -d
```

### What's Running

| Service | URL | Description |
|---------|-----|-------------|
| **API** | `http://localhost:4000` | Film API (all endpoints) |
| **Neo4j Browser** | `http://localhost:7474` | Database admin UI |

The API waits for Neo4j to be fully healthy before starting (`depends_on` + health check).

### Seeding Film Data (Optional)

The API auto-populates the database on first boot if a `filtered_films.csv` file is present. To seed data, mount your CSV into the container:

```yaml
# In docker-compose.yml, under the api service:
services:
  api:
    volumes:
      - ./filtered_films.csv:/filtered_films.csv:ro
```

Then restart:

```bash
docker compose up -d --force-recreate api
```

### Useful Commands

```bash
# View logs
docker compose logs -f api

# Rebuild after code changes
docker compose up -d --build

# Stop everything
docker compose down

# Stop and wipe all data (Neo4j database included)
docker compose down -v
```

---

## 📚 API Documentation

### Base URL
```
http://localhost:4000/v1
```

### Authentication Flow

#### 1. Register User
```http
POST /v1/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "created_at": "2025-11-23T12:00:00Z",
    "name": "John Doe",
    "email": "john@example.com",
    "activated": false
  },
  "activation_token": {
    "token": "ABCDEF123456",
    "expiry": "2025-11-24T12:00:00Z"
  }
}
```

#### 2. Activate Account
```http
PUT /v1/users/activate
Content-Type: application/json

{
  "token": "ABCDEF123456"
}
```

#### 3. Login (Get JWT)
```http
POST /v1/tokens/authentication
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "email": "john@example.com",
    "name": "John Doe",
    "activated": true
  }
}
```

Use access token in subsequent requests:
```
Authorization: Bearer YOUR_ACCESS_TOKEN
```

---

### Films Endpoints

#### List Films
```http
GET /v1/films?page=1&page_size=20&sort=-rating
Authorization: Bearer YOUR_TOKEN
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `page_size` (int): Results per page (default: 20, max: 100)
- `title` (string): Filter by title (partial match)
- `genres` (string): Filter by genres (comma-separated)
- `actors` (string): Filter by actors (comma-separated)
- `directors` (string): Filter by directors (comma-separated)
- `sort` (string): Sort field(s), prefix `-` for descending
  - Options: `id`, `title`, `year`, `runtime`, `rating`

**Example:**
```bash
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:4000/v1/films?title=inception&genres=sci-fi&sort=-rating&page_size=5"
```

**Response:**
```json
{
  "films": [
    {
      "id": 123,
      "imdb_id": "tt1375666",
      "title": "Inception",
      "year": 2010,
      "runtime": "148 mins",
      "certificate": "PG-13",
      "rating": 8.8,
      "description": "A mind-bending thriller...",
      "genres": ["Sci-Fi", "Thriller"],
      "directors": ["Christopher Nolan"],
      "actors": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"],
      "image": "https://m.media-amazon.com/..."
    }
  ],
  "metadata": {
    "current_page": 1,
    "page_size": 5,
    "first_page": 1,
    "last_page": 10,
    "total_records": 50
  }
}
```

#### Get Film by ID
```http
GET /v1/films/{id}
Authorization: Bearer YOUR_TOKEN
```

#### Create Film
```http
POST /v1/films
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "title": "The Matrix",
  "year": 1999,
  "runtime": 136,
  "rating": 8.7,
  "certificate": "R",
  "description": "A computer hacker learns...",
  "genres": ["Action", "Sci-Fi"],
  "directors": ["Lana Wachowski", "Lilly Wachowski"],
  "actors": ["Keanu Reeves", "Laurence Fishburne"]
}
```

#### Update Film
```http
PATCH /v1/films/{id}
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "rating": 9.0,
  "description": "Updated description"
}
```

#### Delete Film
```http
DELETE /v1/films/{id}
Authorization: Bearer YOUR_TOKEN
```

---

### Watchlist Endpoints

#### Add Film to Watchlist
```http
POST /v1/watchlist
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "film_id": 123,
  "notes": "Recommended by friend",
  "priority": 8
}
```

#### Rate a Film
```http
POST /v1/watchlist
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "film_id": 123,
  "rating": 9,
  "notes": "Masterpiece!",
  "priority": 10
}
```
*Rating automatically marks film as watched*

#### Get Your Watchlist
```http
GET /v1/watchlist?watched=false&sort=-priority
Authorization: Bearer YOUR_TOKEN
```

**Query Parameters:**
- `watched` (boolean): Filter by watched status
- `priority` (int): Filter by priority (1-10)
- `sort` (string): `id`, `added_at`, `priority`, `watched`

**Response:**
```json
{
  "watchlist": [
    {
      "id": 1,
      "user_id": 123,
      "film_id": 456,
      "added_at": "2025-11-23T10:00:00Z",
      "notes": "Must watch!",
      "priority": 10,
      "watched": false,
      "watched_at": null,
      "rating": null,
      "film": {
        "id": 456,
        "title": "Inception",
        ...
      }
    }
  ],
  "metadata": {...}
}
```

#### Update Watchlist Entry
```http
PATCH /v1/watchlist/{id}
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "rating": 10,
  "watched": true,
  "notes": "Absolutely incredible"
}
```

#### Remove from Watchlist
```http
DELETE /v1/watchlist/{id}
Authorization: Bearer YOUR_TOKEN
```

---

### 🤖 Recommendations Endpoint

#### Get Personalized Recommendations
```http
GET /v1/recommendations?limit=10
Authorization: Bearer YOUR_TOKEN
```

**How it works:**

The recommendation engine analyzes films in your watchlist and finds similar films using **6 scoring factors**:

1. **Directors** (5× weight) - Same director as your favorite films
2. **Actors** (3× weight) - Same actors
3. **Genres** (1× weight) - Same genres
4. **Certificate** (+2 points) - Same rating (PG-13, R, etc.)
5. **Year Range** (+2 points) - Released within 5 years
6. **Runtime** (+1 point) - Similar duration (±20 minutes)

**Score Formula:**
```
Total Score = (
  (Director matches × 5) +
  (Actor matches × 3) +
  (Genre matches × 1) +
  (Certificate match × 2) +
  (Year range match × 2) +
  (Runtime match × 1)
) × Your Rating/Priority
```

**Example:**

If you rated **The Dark Knight** (2008, PG-13, 152 mins, Nolan) as **10/10**:

**Inception** (2010, PG-13, 148 mins, Nolan) scores:
- Director: 5 × 10 = 50
- Certificate: 2 × 10 = 20
- Year: 2 × 10 = 20
- Runtime: 1 × 10 = 10
- **Total: 100 points** 🎯

**Response:**
```json
{
  "recommendations": [
    {
      "id": 789,
      "title": "Interstellar",
      "year": 2014,
      "runtime": "169 mins",
      "certificate": "PG-13",
      "rating": 8.6,
      "genres": ["Sci-Fi", "Drama"],
      "directors": ["Christopher Nolan"],
      "actors": ["Matthew McConaughey", "Anne Hathaway"],
      ...
    }
  ]
}
```

---

## 🐍 Django Integration

Perfect for Django web applications!

### Setup

```python
# settings.py
JWT_SECRET = os.getenv('JWT_SECRET')  # Same as Go API
JWT_ALGORITHM = 'HS256'
FILM_API_URL = 'http://localhost:4000/v1'
```

### Middleware

```python
# middleware.py
import jwt
from django.conf import settings

class FilmAPIAuthMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response
    
    def __call__(self, request):
        auth_header = request.META.get('HTTP_AUTHORIZATION', '')
        
        if auth_header.startswith('Bearer '):
            token = auth_header.split(' ')[1]
            try:
                # Verify JWT locally (no API call!)
                payload = jwt.decode(
                    token,
                    settings.JWT_SECRET,
                    algorithms=[settings.JWT_ALGORITHM]
                )
                request.user_id = payload['user_id']
                request.user_email = payload['email']
                request.user_activated = payload['activated']
            except jwt.ExpiredSignatureError:
                # Handle token refresh
                pass
            except jwt.InvalidTokenError:
                pass
        
        return self.get_response(request)
```

### Views

```python
# views.py
import requests
from django.conf import settings

def login_view(request):
    """Login via Film API and store JWT"""
    response = requests.post(
        f'{settings.FILM_API_URL}/tokens/authentication',
        json={
            'email': request.POST['email'],
            'password': request.POST['password']
        }
    )
    
    if response.status_code == 201:
        data = response.json()
        # Store tokens in session
        request.session['access_token'] = data['access_token']
        request.session['refresh_token'] = data['refresh_token']
        return redirect('dashboard')
    
    return render(request, 'login.html', {'error': 'Invalid credentials'})

def get_films(request):
    """Fetch films from API"""
    headers = {
        'Authorization': f"Bearer {request.session.get('access_token')}"
    }
    
    response = requests.get(
        f'{settings.FILM_API_URL}/films',
        headers=headers,
        params={'page_size': 20, 'sort': '-rating'}
    )
    
    films = response.json().get('films', [])
    return render(request, 'films.html', {'films': films})

def get_recommendations(request):
    """Get personalized recommendations"""
    headers = {
        'Authorization': f"Bearer {request.session.get('access_token')}"
    }
    
    response = requests.get(
        f'{settings.FILM_API_URL}/recommendations',
        headers=headers,
        params={'limit': 10}
    )
    
    recommendations = response.json().get('recommendations', [])
    return render(request, 'recommendations.html', {
        'recommendations': recommendations
    })
```

---

## 🔐 Security Features

- ✅ **JWT Authentication**: Industry-standard, cryptographically secure
- ✅ **bcrypt Password Hashing**: Slow hash function resistant to brute-force
- ✅ **Token Expiry**: Access tokens expire in 1 hour
- ✅ **Refresh Token Revocation**: Logout via database deletion
- ✅ **CORS Protection**: Configurable trusted origins
- ✅ **Rate Limiting**: Prevents API abuse
- ✅ **Input Validation**: Comprehensive data validation
- ✅ **SQL Injection Safe**: Parameterized Cypher queries

---

## ⚡ Performance

### JWT vs Database Tokens

**Request Flow:**

```
Before (Database Tokens):
Client → API → Auth → Neo4j Query (10-50ms) → Process
Total: ~50-100ms

After (JWT):
Client → API → Auth → JWT Verify (1ms) → Process
Total: ~10-20ms
```

**At 10,000 requests/second:**
- Database Tokens: 100,000-500,000 DB queries/sec 😰
- JWT: 0 DB queries for auth 🚀

---

## 📊 Database Schema (Neo4j)

### Nodes
- `User` - Users with credentials
- `Film` - Films with metadata
- `Genre` - Film genres
- `Actor` - Actors
- `Director` - Directors
- `Watchlist` - User watchlist entries
- `Token` - Activation tokens
- `RefreshToken` - JWT refresh tokens (for revocation)
- `Permission` - User permissions

### Relationships
```cypher
(Film)-[:HAS_GENRE]->(Genre)
(Film)-[:HAS_ACTOR]->(Actor)
(Film)-[:HAS_DIRECTOR]->(Director)
(User)-[:HAS_WATCHLIST]->(Watchlist)-[:FOR_FILM]->(Film)
(RefreshToken)-[:BELONGS_TO]->(User)
(User)-[:HAS_PERMISSION]->(Permission)
```

---

## 🧪 Testing

```bash
# Run test script
chmod +x test_api.sh
./test_api.sh
```

Tests include:
- User registration & activation
- JWT authentication
- Film listing & filtering
- Watchlist management
- Recommendations

---

## ⚙️ Configuration

**Environment Variables:**

| Variable | Description | Default |
|----------|-------------|---------|
| `NEO4J_URI` | Neo4j connection URI | Required |
| `NEO4J_USERNAME` | Neo4j username | `neo4j` |
| `NEO4J_PASSWORD` | Neo4j password | Required |
| `JWT_SECRET` | JWT signing secret (32+ chars) | Required |
| `APP_PORT` | Server port | `4000` |
| `APP_ENV` | Environment | `development` |
| `LIMITER_RPS` | Rate limit (requests/sec) | `2` |
| `LIMITER_BURST` | Rate limit burst | `4` |
| `CORS_TRUSTED_ORIGINS` | Allowed CORS origins | Empty |

---

## 🚦 Error Handling

**HTTP Status Codes:**

| Code | Meaning |
|------|---------|
| `200 OK` | Success |
| `201 Created` | Resource created |
| `400 Bad Request` | Invalid input |
| `401 Unauthorized` | Missing/invalid token |
| `403 Forbidden` | Insufficient permissions |
| `404 Not Found` | Resource not found |
| `429 Too Many Requests` | Rate limit exceeded |
| `500 Internal Server Error` | Server error |

**Error Response:**
```json
{
  "error": "Detailed error message"
}
```

**Validation Errors:**
```json
{
  "error": {
    "email": "must be a valid email address",
    "password": "must be at least 8 characters"
  }
}
```

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

---

## 📝 License

MIT License - see LICENSE file for details

---

## 🙏 Acknowledgments

- Film data from IMDb Top Films dataset
- Built with [Neo4j](https://neo4j.com/) Graph Database
- Authentication powered by [golang-jwt](https://github.com/golang-jwt/jwt)
- Inspired by modern microservices architecture

---

## 📞 Support

For issues and questions:
- Open an issue on GitHub
- Check existing documentation
- Review API examples above

---

**Built with ❤️ using Go and Neo4j**

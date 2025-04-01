#!/bin/bash

# Test public routes (no auth required)
echo "Testing public routes..."
echo "1. Healthcheck (should succeed):"
curl -i localhost:8080/v1/healthcheck
echo -e "\n"

# Test authentication
echo "Testing authentication..."

echo "2. Create user (should succeed):"
CREATE_USER_RESPONSE=$(curl -i -X POST localhost:8080/v1/user -H "Content-Type: application/json" -d '{
    "name": "testuser1",
    "email": "test1@example.com",
    "password": "securepassword123"
}')

echo "$CREATE_USER_RESPONSE"
echo -e "\n"

# Extract activation token from create user response
ACTIVATION_TOKEN=$(echo "$CREATE_USER_RESPONSE" | grep -oP '(?<="token":")[^"]+')

echo "3. Activate user (should succeed):"
curl -i -X PUT localhost:8080/v1/users/activated -H "Content-Type: application/json" -d '{
    "token": "'$ACTIVATION_TOKEN'"
}'
echo -e "\n"

# Get authentication token for user 1
echo "4. Get authentication token for user 1 (should succeed):"
TOKEN1=$(curl -s -X POST localhost:8080/v1/tokens/authentication -H "Content-Type: application/json" -d '{
    "email": "test1@example.com",
    "password": "securepassword123"
}' | jq -r '.authentication_token.token')

echo "Testing protected routes..."
echo "5. Get films without auth (should fail):"
curl -i localhost:8080/v1/films
echo -e "\n"

# Create user 2 for additional tests
echo "6. Create user 2 (should succeed):"
CREATE_USER2_RESPONSE=$(curl -i -X POST localhost:8080/v1/user -H "Content-Type: application/json" -d '{
    "name": "testuser2",
    "email": "test2@example.com",
    "password": "securepassword123"
}')

echo "$CREATE_USER2_RESPONSE"
echo -e "\n"

# Extract activation token for user 2
ACTIVATION_TOKEN2=$(echo "$CREATE_USER2_RESPONSE" | grep -oP '(?<="token":")[^"]+')

echo "7. Activate user 2 (should succeed):"
curl -i -X PUT localhost:8080/v1/users/activated -H "Content-Type: application/json" -d '{
    "token": "'$ACTIVATION_TOKEN2'"
}'
echo -e "\n"

# Get authentication token for user 2
echo "8. Get authentication token for user 2 (should succeed):"
TOKEN2=$(curl -s -X POST localhost:8080/v1/tokens/authentication -H "Content-Type: application/json" -d '{
    "email": "test2@example.com",
    "password": "securepassword123"
}' | jq -r '.authentication_token.token')

echo "9. Get films with auth (should succeed):"
curl -i -H "Authorization: Bearer $TOKEN1" localhost:8080/v1/films
echo -e "\n"

# Create film with user 1
echo "10. Create film with user 1 (should succeed):"
CREATE_FILM_RESPONSE=$(curl -i -X POST localhost:8080/v1/films -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN1" -d '{
    "title": "Test Film",
    "year": 2025,
    "runtime": "120 mins",
    "genres": ["Action", "Adventure"],
    "directors": ["Director Name"],
    "actors": ["Actor Name"],
    "rating": 8.5,
    "description": "Test film description",
    "image": "https://example.com/image.jpg"
}')

echo "$CREATE_FILM_RESPONSE"
echo -e "\n"

# Extract film ID from create response
FILM_ID=$(echo "$CREATE_FILM_RESPONSE" | grep -oP '(?<="id":)[0-9]+')

echo "11. Get specific film without auth (should fail):"
curl -i localhost:8080/v1/films/$FILM_ID
echo -e "\n"

# Test getting film with user 1's token
echo "12. Get specific film with user 1's auth (should succeed):"
curl -i -H "Authorization: Bearer $TOKEN1" localhost:8080/v1/films/$FILM_ID
echo -e "\n"

# Test getting film with user 2's token (should succeed)
echo "13. Get specific film with user 2's auth (should succeed):"
curl -i -H "Authorization: Bearer $TOKEN2" localhost:8080/v1/films/$FILM_ID
echo -e "\n"

# Test updating film with user 1's token
echo "14. Update film with user 1's auth (should succeed):"
curl -i -X PATCH localhost:8080/v1/films/$FILM_ID -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN1" -d '{
    "title": "Updated Test Film"
}'
echo -e "\n"

# Test updating film with user 2's token (should fail)
echo "15. Update film with user 2's auth (should fail):"
curl -i -X PATCH localhost:8080/v1/films/$FILM_ID -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN2" -d '{
    "title": "Another Update"
}'
echo -e "\n"

# Test deleting film with user 1's token
echo "16. Delete film with user 1's auth (should succeed):"
curl -i -X DELETE localhost:8080/v1/films/$FILM_ID -H "Authorization: Bearer $TOKEN1"
echo -e "\n"

# Test deleting film with user 2's token (should fail)
echo "17. Delete film with user 2's auth (should fail):"
curl -i -X DELETE localhost:8080/v1/films/$FILM_ID -H "Authorization: Bearer $TOKEN2"
echo -e "\n"

# Test using invalid token
echo "18. Try to use invalid token (should fail):"
curl -i -X GET localhost:8080/v1/films -H "Authorization: Bearer invalidtoken"
echo -e "\n"

# Test using malformed token
echo "19. Try to use malformed token (should fail):"
curl -i -X GET localhost:8080/v1/films -H "Authorization: Bearer invalid token"
echo -e "\n"

# Test using wrong auth scheme
echo "20. Try to use wrong auth scheme (should fail):"
curl -i -X GET localhost:8080/v1/films -H "Authorization: Basic $TOKEN1"
echo -e "\n"

echo "All tests completed!"
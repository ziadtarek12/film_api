# #!/bin/bash

# # Test the healthcheck endpoint
# echo "Testing healthcheck endpoint..."
# curl -i -X GET http://localhost:8080/v1/healthcheck

# Test creating a new film
echo -e "\n\nTesting create film endpoint..."
curl -i -X POST http://localhost:8080/v1/films \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Godfather",
    "year": 1972,
    "runtime": "175 mins",
    "genres": ["Crime", "Drama"],
    "directors": ["Francis Ford Coppola"],
    "actors": ["Marlon Brando", "Al Pacino", "James Caan"],
    "rating": 9.2,
    "description": "The aging patriarch of an organized crime dynasty transfers control of his clandestine empire to his reluctant son.",
    "image": "https://example.com/godfather.jpg"
  }'

# # Test getting a film by ID (replace {id} with an actual ID after creating a film)
# echo -e "\n\nTesting get film endpoint..."
# curl -i -X GET http://localhost:8080/v1/films/8

# # Test updating a film
# echo -e "\n\nTesting update film endpoint..."
# curl -i -X PATCH http://localhost:8080/v1/films/8 \
#   -H "Content-Type: application/json" \
#   -d '{
#     "title": "The Godfather: Part II",
#     "year": 1974,
#     "runtime": "202 mins",
#     "genres": ["Crime", "Drama"],
#     "directors": ["Francis Ford Coppola"],
#     "actors": ["Al Pacino", "Robert De Niro", "Robert Duvall"],
#     "rating": 9.0,
#     "description": "The early life and career of Vito Corleone in 1920s New York City is portrayed, while his son, Michael, expands and tightens his grip on the family crime syndicate.",
#     "image": "https://example.com/godfather2.jpg"
#   }'

# curl -i -X PATCH http://localhost:8080/v1/films/8 \
#   -H "Content-Type: application/json" \
#   -d '{
#     "title": "The Godfather",
#     "year": 1974,
#     "runtime": "202 mins",
#     "genres": ["Crime", "Drama"],
#     "directors": ["Francis Ford Coppola"],
#     "actors": ["Al Pacino", "Robert De Niro", "Robert Duvall"],
#     "rating": 9.0,
#     "description": "The early life and career of Vito Corleone in 1920s New York City is portrayed, while his son, Michael, expands and tightens his grip on the family crime syndicate.",
#     "image": "https://example.com/godfather2.jpg"
# }'


# echo -e "\n\nTesting delete film endpoint..."
# curl -X DELETE localhost:8080/v1/films/8


# # Add a test for race condition in updateFilmHandler
# echo "Testing race condition in updateFilmHandler..."

# # Create a film to test with
# FILM_ID=$(curl -s -X POST http://localhost:8080/v1/films -H "Content-Type: application/json" -d '{"title": "Race Test", "year": 2023, "runtime": "120 mins", "genres": ["Action"], "directors": ["John Doe"], "actors": ["Jane Doe"], "rating": 7.5, "description": "Test film for race condition", "image": "http://example.com/image.jpg"}' | jq -r '.film.id')

# # Function to update film
# update_film() {
#     curl -s -X PATCH http://localhost:8080/v1/films/$FILM_ID -H "Content-Type: application/json" -d '{"title": "Updated Title '$1'"}'
# }

# # Run multiple updates concurrently
# update_film 1 &
# update_film 2 &
# update_film 3 &
# update_film 4 &

# # Wait for all updates to complete
# wait

# Fetch the film to check the final state
# FINAL_TITLE=$(curl -s http://localhost:8080/v1/films/$FILM_ID | jq -r '.film.title')

# # Check if the final title is as expected
# if [[ $FINAL_TITLE != "Updated Title 4" ]]; then
#     echo "Race condition detected: Final title is '$FINAL_TITLE'"
# else
#     echo "No race condition detected: Final title is '$FINAL_TITLE'"
# fi

# # Clean up
# curl -s -X DELETE http://localhost:8080/v1/films/$FILM_ID
#!/bin/bash

# Create SQLite database and run migrations
echo "Setting up SQLite database..."

# Remove old database if exists
rm -f database.sqlite

# Create new database and run migration
sqlite3 database.sqlite < migrations/000008_create_sqlite_tables.up.sql

echo "SQLite database setup complete!"
echo "Database file: database.sqlite"

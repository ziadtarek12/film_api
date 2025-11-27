package main

import (
	"filmapi.zeyadtarek.net/internals/tokens"
)

// verifyJWT is a helper method to verify JWT tokens
func (app *application) verifyJWT(tokenString string) (*tokens.Claims, error) {
	return tokens.VerifyJWT(tokenString, app.config.jwt.secret)
}

// generateAccessToken creates a new JWT access token
func (app *application) generateAccessToken(userID int64, email string, activated bool) (string, error) {
	return tokens.GenerateJWT(
		userID,
		email,
		activated,
		"access",
		app.config.jwt.accessTokenTTL,
		app.config.jwt.secret,
	)
}

// generateRefreshToken creates a new JWT refresh token
func (app *application) generateRefreshToken(userID int64, email string, activated bool) (string, error) {
	return tokens.GenerateJWT(
		userID,
		email,
		activated,
		"refresh",
		app.config.jwt.refreshTokenTTL,
		app.config.jwt.secret,
	)
}

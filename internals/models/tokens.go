package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserId    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userId int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserId: userId,
		Expiry: time.Now().UTC().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	Driver neo4j.DriverWithContext
}

func (model TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = model.Insert(token)
	return token, err
}

func (model TokenModel) Insert(token *Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)
			WHERE id(u) = $user_id
			CREATE (t:Token {
				hash: $hash,
				expiry: datetime($expiry),
				scope: $scope
			})
			MERGE (t)-[:BELONGS_TO]->(u)
			RETURN id(t)
		`

		params := map[string]any{
			"user_id": token.UserId,
			"hash":    []byte(token.Hash),
			"expiry":  token.Expiry,
			"scope":   token.Scope,
		}

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

func (model TokenModel) DeleteAllForUser(scope string, userID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)<-[:BELONGS_TO]-(t:Token)
			WHERE id(u) = $user_id AND t.scope = $scope
			DETACH DELETE t
		`

		params := map[string]any{
			"user_id": userID,
			"scope":   scope,
		}

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// InsertJWTRefresh stores a JWT refresh token hash in the database for revocation tracking
func (model TokenModel) InsertJWTRefresh(userID int64, jwtToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// Hash the JWT token
	hash := sha256.Sum256([]byte(jwtToken))

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)
			WHERE id(u) = $user_id
			CREATE (t:RefreshToken {
				hash: $hash,
				created_at: datetime()
			})
			MERGE (t)-[:BELONGS_TO]->(u)
			RETURN id(t)
		`

		params := map[string]any{
			"user_id": userID,
			"hash":    hash[:],
		}

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// ValidateJWTRefresh checks if a JWT refresh token exists and is not revoked
func (model TokenModel) ValidateJWTRefresh(userID int64, jwtToken string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Hash the JWT token
	hash := sha256.Sum256([]byte(jwtToken))

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)<-[:BELONGS_TO]-(t:RefreshToken)
			WHERE id(u) = $user_id AND t.hash = $hash
			RETURN count(t) > 0 AS exists
		`

		params := map[string]any{
			"user_id": userID,
			"hash":    hash[:],
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return false, err
		}

		if result.Next(ctx) {
			record := result.Record()
			exists, _ := record.Get("exists")
			return exists.(bool), nil
		}

		return false, nil
	})

	if err != nil {
		return false, err
	}

	return result.(bool), nil
}

// RevokeRefreshToken removes a JWT refresh token from the database (logout)
func (model TokenModel) RevokeRefreshToken(userID int64, jwtToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// Hash the JWT token
	hash := sha256.Sum256([]byte(jwtToken))

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (t:RefreshToken {hash: $hash})-[:BELONGS_TO]->(u:User)
			WHERE id(u) = $user_id
			DETACH DELETE t
		`

		params := map[string]any{
			"user_id": userID,
			"hash":    hash[:],
		}

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

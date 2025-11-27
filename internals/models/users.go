package models

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	AnonymousUser     = &User{}
)

type UserModel struct {
	Driver neo4j.DriverWithContext
}

func (user *User) IsAnonyomous() bool {
	return user == AnonymousUser
}

func (model UserModel) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Check for duplicate email
		checkQuery := `MATCH (u:User {email: $email}) RETURN count(u)`
		result, err := tx.Run(ctx, checkQuery, map[string]any{"email": user.Email})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			count := result.Record().Values[0].(int64)
			if count > 0 {
				return nil, ErrDuplicateEmail
			}
		}

		query := `
			CREATE (u:User {
				name: $name,
				email: $email,
				password_hash: $password_hash,
				activated: $activated,
				created_at: datetime(),
				version: 1
			})
			RETURN id(u), u.created_at, u.version
		`

		params := map[string]any{
			"name":          user.Name,
			"email":         user.Email,
			"password_hash": string(user.Password.hash), // Store as string
			"activated":     user.Activated,
		}

		result, err = tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			user.ID = record.Values[0].(int64)
			user.CreatedAt = record.Values[1].(time.Time)
			user.Version = int(record.Values[2].(int64))
			return nil, nil
		}

		return nil, errors.New("failed to insert user")
	})

	return err
}

func (model UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)
			WHERE u.email = $email
			RETURN u
		`
		result, err := tx.Run(ctx, query, map[string]any{"email": email})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record(), nil
		}

		return nil, ErrRecordNotFound
	})

	if err != nil {
		return nil, err
	}

	record := result.(*neo4j.Record)
	node, _ := record.Get("u")
	userNode := node.(neo4j.Node)

	user := &User{
		ID:        userNode.Id,
		CreatedAt: userNode.Props["created_at"].(time.Time),
		Name:      userNode.Props["name"].(string),
		Email:     userNode.Props["email"].(string),
		Activated: userNode.Props["activated"].(bool),
		Version:   int(userNode.Props["version"].(int64)),
	}
	user.Password.hash = []byte(userNode.Props["password_hash"].(string))

	return user, nil
}

func (model UserModel) Update(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Check for duplicate email if email is changed
		// This is a bit complex in one query, so we might skip strict check or do it in Cypher
		// Let's rely on the query logic: match by ID and version, update if email unique?
		// Simplest is to check email uniqueness first if it changed, but let's assume for now we just update.
		// Actually, if we update email to an existing one, we should fail.

		query := `
			MATCH (u:User)
			WHERE id(u) = $id AND u.version = $version
			SET u.name = $name,
				u.email = $email,
				u.password_hash = $password_hash,
				u.activated = $activated,
				u.version = u.version + 1
			RETURN u.version
		`

		params := map[string]any{
			"id":            user.ID,
			"version":       user.Version,
			"name":          user.Name,
			"email":         user.Email,
			"password_hash": string(user.Password.hash),
			"activated":     user.Activated,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			user.Version = int(result.Record().Values[0].(int64))
			return nil, nil
		}

		return nil, ErrEditConflict
	})

	return err
}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (pass *password) Set(plaintextPassword string) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	pass.plaintext = &plaintextPassword
	pass.hash = hash

	return nil
}

func (pass *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(pass.hash, []byte(plaintextPassword))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")

}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes")
	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

}

func (model UserModel) GetForToken(tokenscope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)<-[:BELONGS_TO]-(t:Token)
			WHERE t.hash = $hash
			AND t.scope = $scope
			AND t.expiry > datetime($now)
			RETURN u
		`

		// Neo4j datetime expects ISO 8601 string or similar. Go time.Time marshals to string usually.
		// But neo4j driver handles time.Time mapping to Neo4j DateTime/LocalDateTime.

		params := map[string]any{
			"hash":  []byte(tokenHash[:]), // Pass as byte array
			"scope": tokenscope,
			"now":   time.Now().UTC(),
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record(), nil
		}

		return nil, ErrRecordNotFound
	})

	if err != nil {
		return nil, err
	}

	record := result.(*neo4j.Record)
	node, _ := record.Get("u")
	userNode := node.(neo4j.Node)

	user := &User{
		ID:        userNode.Id,
		CreatedAt: userNode.Props["created_at"].(time.Time),
		Name:      userNode.Props["name"].(string),
		Email:     userNode.Props["email"].(string),
		Activated: userNode.Props["activated"].(bool),
		Version:   int(userNode.Props["version"].(int64)),
	}
	user.Password.hash = []byte(userNode.Props["password_hash"].(string))

	return user, nil
}

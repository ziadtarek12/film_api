package models

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Permissions []string

func (permissions Permissions) Include(code string) bool {
	for i := range permissions {
		if code == permissions[i] {
			return true
		}
	}

	return false
}

type PermissionModel struct {
	Driver neo4j.DriverWithContext
}

func (model PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)-[:HAS_PERMISSION]->(p:Permission)
			WHERE id(u) = $user_id
			RETURN collect(p.code)
		`
		result, err := tx.Run(ctx, query, map[string]any{"user_id": userID})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}

		return []any{}, nil
	})

	if err != nil {
		return nil, err
	}

	permissionsRaw := result.([]any)
	permissions := make(Permissions, len(permissionsRaw))
	for i, p := range permissionsRaw {
		permissions[i] = p.(string)
	}

	return permissions, nil
}

func (model PermissionModel) AddForUser(userID int64, codes ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)
			WHERE id(u) = $user_id
			UNWIND $codes as code
			MERGE (p:Permission {code: code})
			MERGE (u)-[:HAS_PERMISSION]->(p)
		`

		params := map[string]any{
			"user_id": userID,
			"codes":   codes,
		}

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

package models

import (
"context"
"time"

"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (m FilmModel) GetRecommendations(id int64) ([]*Film, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	query := `
		MATCH (target:Film {id: $id})
		CALL db.index.vector.queryNodes('movie_plots_index', 50, target.embedding) YIELD node AS recommend, score AS vectorScore
		WHERE recommend.id <> $id
		MATCH (recommend)
		OPTIONAL MATCH (recommend)-[:DIRECTED_BY]->(d:Director)<-[:DIRECTED_BY]-(target)
		OPTIONAL MATCH (recommend)-[:ACTED_IN]->(a:Actor)<-[:ACTED_IN]-(target)
		OPTIONAL MATCH (recommend)-[:IN_GENRE]->(g:Genre)<-[:IN_GENRE]-(target)
		WITH recommend, vectorScore,
			 count(DISTINCT d) * 5.0 AS directorScore,
			 count(DISTINCT a) * 3.0 AS actorScore,
			 count(DISTINCT g) * 1.0 AS genreScore
		WITH recommend, vectorScore + directorScore + actorScore + genreScore AS totalScore
		ORDER BY totalScore DESC
		RETURN recommend, totalScore
	`

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, query, map[string]interface{}{
"id": id,
})
	if err != nil {
		return nil, err
	}

	var recommendations []*Film
	for result.Next(ctx) {
		record := result.Record()
		node, _, err := neo4j.GetRecordValue[neo4j.Node](record, "recommend")
		if err != nil {
			return nil, err
		}

		film := &Film{
			ID:            node.Props["id"].(int64),
			Title:         node.Props["title"].(string),
			OriginalTitle: node.Props["original_title"].(string),
			Year:          int32(node.Props["year"].(int64)),
			Description:   node.Props["description"].(string),
		}

		recommendations = append(recommendations, film)
	}

	if err = result.Err(); err != nil {
		return nil, err
	}

	return recommendations, nil
}

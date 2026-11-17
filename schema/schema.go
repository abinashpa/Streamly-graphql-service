package schema

import (
	"streamly/graphql-service/resolvers"

	"github.com/graphql-go/graphql"
)

func BuildSchema(client *resolvers.Client) (graphql.Schema, error) {
	mediaItemType := CreateMediaItemType(client)
	searchResultType := CreateSearchResultType(mediaItemType)
	rootQuery := CreateRootQuery(client, mediaItemType, searchResultType)
	rootMutation := CreateRootMutation(client)

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
}

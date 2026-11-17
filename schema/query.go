package schema

import (
	"streamly/graphql-service/resolvers"

	"github.com/graphql-go/graphql"
)

func CreateRootQuery(client *resolvers.Client, mediaItemType *graphql.Object, searchResultType *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// Feed query
			"feed": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(mediaItemType))),
				Description: "Get paginated media feed items (cached in Redis)",
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{
						Type:         MediaTypeEnum,
						DefaultValue: "all",
					},
					"page": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 1,
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 20,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					mediaType, _ := p.Args["type"].(string)
					page, _ := p.Args["page"].(int)
					limit, _ := p.Args["limit"].(int)
					return client.FetchFeed(mediaType, page, limit)
				},
			},

			// Trending query
			"trending": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(mediaItemType))),
				Description: "Get trending media items sorted by popularity",
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{
						Type:         MediaTypeEnum,
						DefaultValue: "all",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					mediaType, _ := p.Args["type"].(string)
					return client.FetchTrending(mediaType)
				},
			},

			// Single media item query
			"media": &graphql.Field{
				Type:        mediaItemType,
				Description: "Get single media details by ID, including nested comments and uploader profile",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)
					return client.FetchMedia(id)
				},
			},

			// Search query
			"search": &graphql.Field{
				Type:        graphql.NewNonNull(searchResultType),
				Description: "Search media using OpenSearch multi-match fuzzy search",
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"type": &graphql.ArgumentConfig{
						Type:         MediaTypeEnum,
						DefaultValue: "all",
					},
					"page": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 1,
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 20,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					query, _ := p.Args["query"].(string)
					mediaType, _ := p.Args["type"].(string)
					page, _ := p.Args["page"].(int)
					limit, _ := p.Args["limit"].(int)
					return client.SearchMedia(query, mediaType, page, limit)
				},
			},

			// Autocomplete suggestions
			"suggestions": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
				Description: "Fast OpenSearch edge n-gram autocomplete suggestions",
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 6,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					query, _ := p.Args["query"].(string)
					limit, _ := p.Args["limit"].(int)
					return client.FetchSuggestions(query, limit)
				},
			},

			// User channel profile
			"user": &graphql.Field{
				Type:        UserType,
				Description: "Get public channel profile by user ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)
					return client.FetchUser(id)
				},
			},

			// Authenticated user profile
			"me": &graphql.Field{
				Type:        UserType,
				Description: "Get currently authenticated user from Bearer token",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					token, _ := p.Context.Value("Authorization").(string)
					return client.FetchMe(token)
				},
			},
		},
	})
}

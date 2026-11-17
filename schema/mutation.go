package schema

import (
	"streamly/graphql-service/resolvers"

	"github.com/graphql-go/graphql"
)

func CreateRootMutation(client *resolvers.Client) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			// Record view
			"recordView": &graphql.Field{
				Type:        graphql.NewNonNull(ViewPayloadType),
				Description: "Increment media view counter atomically",
				Args: graphql.FieldConfigArgument{
					"mediaId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					mediaID, _ := p.Args["mediaId"].(string)
					views, err := client.RecordView(mediaID)
					if err != nil {
						return nil, err
					}
					return map[string]interface{}{
						"id":    mediaID,
						"views": views,
					}, nil
				},
			},

			// Toggle like
			"toggleLike": &graphql.Field{
				Type:        graphql.NewNonNull(LikePayloadType),
				Description: "Toggle like or unlike on media item",
				Args: graphql.FieldConfigArgument{
					"mediaId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"userId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					mediaID, _ := p.Args["mediaId"].(string)
					userID, _ := p.Args["userId"].(string)
					liked, err := client.ToggleLike(mediaID, userID)
					if err != nil {
						return nil, err
					}
					return map[string]interface{}{
						"liked": liked,
					}, nil
				},
			},

			// Add comment
			"addComment": &graphql.Field{
				Type:        graphql.NewNonNull(CommentType),
				Description: "Post a new comment on a media item",
				Args: graphql.FieldConfigArgument{
					"mediaId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"content": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"userName": &graphql.ArgumentConfig{
						Type:         graphql.String,
						DefaultValue: "Streamly User",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					mediaID, _ := p.Args["mediaId"].(string)
					content, _ := p.Args["content"].(string)
					userName, _ := p.Args["userName"].(string)
					return client.AddComment(mediaID, content, userName)
				},
			},
		},
	})
}

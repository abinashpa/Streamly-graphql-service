package schema

import (
	"streamly/graphql-service/resolvers"

	"github.com/graphql-go/graphql"
)

var MediaTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "MediaType",
	Values: graphql.EnumValueConfigMap{
		"ALL": &graphql.EnumValueConfig{
			Value: "all",
		},
		"VIDEO": &graphql.EnumValueConfig{
			Value: "video",
		},
		"SHORT": &graphql.EnumValueConfig{
			Value: "short",
		},
		"MUSIC": &graphql.EnumValueConfig{
			Value: "music",
		},
	},
})

var CommentType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Comment",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"catalog_id":  &graphql.Field{Type: graphql.String},
		"user_id":     &graphql.Field{Type: graphql.String},
		"user_name":   &graphql.Field{Type: graphql.String},
		"user_avatar": &graphql.Field{Type: graphql.String},
		"content":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"likes":       &graphql.Field{Type: graphql.Int},
		"created_at":  &graphql.Field{Type: graphql.String},
	},
})

var UserType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id":           &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"username":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"email":        &graphql.Field{Type: graphql.String},
		"display_name": &graphql.Field{Type: graphql.String},
		"bio":          &graphql.Field{Type: graphql.String},
		"avatar_url":   &graphql.Field{Type: graphql.String},
		"banner_url":   &graphql.Field{Type: graphql.String},
		"subscribers":  &graphql.Field{Type: graphql.Int},
		"created_at":   &graphql.Field{Type: graphql.String},
	},
})

func CreateMediaItemType(client *resolvers.Client) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "MediaItem",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"id":            &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
				"title":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"description":   &graphql.Field{Type: graphql.String},
				"media_type":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"artist":        &graphql.Field{Type: graphql.String},
				"duration":      &graphql.Field{Type: graphql.Float},
				"stream_url":    &graphql.Field{Type: graphql.String},
				"thumbnail_url": &graphql.Field{Type: graphql.String},
				"uploader_id":   &graphql.Field{Type: graphql.String},
				"uploader_name": &graphql.Field{Type: graphql.String},
				"tags":          &graphql.Field{Type: graphql.String},
				"views":         &graphql.Field{Type: graphql.Int},
				"likes":         &graphql.Field{Type: graphql.Int},
				"created_at":    &graphql.Field{Type: graphql.String},

				// Nested Resolver: Fetch comments for this media item
				"comments": &graphql.Field{
					Type: graphql.NewList(CommentType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if source, ok := p.Source.(resolvers.MediaItem); ok {
							return client.FetchComments(source.ID)
						}
						return []resolvers.Comment{}, nil
					},
				},

				// Nested Resolver: Fetch uploader profile
				"uploader": &graphql.Field{
					Type: UserType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if source, ok := p.Source.(resolvers.MediaItem); ok && source.UploaderID != "" {
							return client.FetchUser(source.UploaderID)
						}
						return nil, nil
					},
				},
			}
		}),
	})
}

func CreateSearchResultType(mediaType *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "SearchResult",
		Fields: graphql.Fields{
			"total": &graphql.Field{Type: graphql.Int},
			"took":  &graphql.Field{Type: graphql.Int},
			"page":  &graphql.Field{Type: graphql.Int},
			"limit": &graphql.Field{Type: graphql.Int},
			"items": &graphql.Field{Type: graphql.NewList(mediaType)},
		},
	})
}

var ViewPayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ViewPayload",
	Fields: graphql.Fields{
		"id":    &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"views": &graphql.Field{Type: graphql.Int},
	},
})

var LikePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "LikePayload",
	Fields: graphql.Fields{
		"liked": &graphql.Field{Type: graphql.Boolean},
	},
})

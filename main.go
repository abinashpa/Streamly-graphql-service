package main

import (
	"context"
	"log"
	"net/http"

	"streamly/graphql-service/config"
	"streamly/graphql-service/resolvers"
	"streamly/graphql-service/schema"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

const playgroundHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset=utf-8 />
  <meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
  <title>Streamly GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="https://api.dicebear.com/7.x/identicon/svg?seed=streamly" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root">
    <style>
      body {
        background-color: #0f0f13;
        font-family: Open Sans, sans-serif;
        margin: 0;
        overflow: hidden;
      }
      #root {
        height: 100vh;
      }
    </style>
  </div>
  <script>
    window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '/graphql',
        settings: {
          'editor.theme': 'dark',
          'editor.cursorShape': 'line',
          'tracing.hideTracingResponse': true
        }
      })
    })
  </script>
</body>
</html>`

type GraphQLRequestBody struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func main() {
	cfg := config.LoadConfig()
	client := resolvers.NewClient(cfg)

	gqlSchema, err := schema.BuildSchema(client)
	if err != nil {
		log.Fatalf("Failed to build GraphQL schema: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "graphql-service",
		})
	})

	// GraphQL Playground (Interactive Browser IDE)
	r.GET("/graphql", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, playgroundHTML)
	})
	r.GET("/playground", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, playgroundHTML)
	})

	// GraphQL Execution Endpoint
	r.POST("/graphql", func(c *gin.Context) {
		var req GraphQLRequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body format"})
			return
		}

		// Inject Authorization header into context for protected queries like `me`
		authHeader := c.GetHeader("Authorization")
		ctx := context.WithValue(c.Request.Context(), "Authorization", authHeader)

		result := graphql.Do(graphql.Params{
			Schema:         gqlSchema,
			RequestString:  req.Query,
			VariableValues: req.Variables,
			OperationName:  req.OperationName,
			Context:        ctx,
		})

		if len(result.Errors) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"data":   result.Data,
				"errors": result.Errors,
			})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	log.Printf("Streamly GraphQL Service running on port %s", cfg.Port)
	log.Printf("Interactive GraphQL Playground: http://localhost:%s/graphql", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("GraphQL server failed: %v", err)
	}
}

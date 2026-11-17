package resolvers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"streamly/graphql-service/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Media Item Representation
type MediaItem struct {
	ID           string  `json:"id"`
	MediaID      string  `json:"media_id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	MediaType    string  `json:"media_type"`
	Artist       string  `json:"artist"`
	Duration     float64 `json:"duration"`
	StreamURL    string  `json:"stream_url"`
	ThumbnailURL string  `json:"thumbnail_url"`
	UploaderID   string  `json:"uploader_id"`
	UploaderName string  `json:"uploader_name"`
	Tags         string  `json:"tags"`
	Views        int64   `json:"views"`
	Likes        int64   `json:"likes"`
	CreatedAt    string  `json:"created_at"`
}

type Comment struct {
	ID         string `json:"id"`
	CatalogID  string `json:"catalog_id"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
	Content    string `json:"content"`
	Likes      int64  `json:"likes"`
	CreatedAt  string `json:"created_at"`
}

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
	BannerURL   string `json:"banner_url"`
	Subscribers int64  `json:"subscribers"`
	CreatedAt   string `json:"created_at"`
}

type SearchResult struct {
	Total int         `json:"total"`
	Took  int         `json:"took"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Items []MediaItem `json:"items"`
}

func (c *Client) dialGRPC(target string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// FetchFeed queries catalog-service via gRPC with REST fallback
func (c *Client) FetchFeed(mediaType string, page, limit int) ([]MediaItem, error) {
	conn, err := c.dialGRPC(c.cfg.CatalogServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Items []MediaItem
		}
		err = conn.Invoke(ctx, "/catalog.CatalogService/GetFeed", map[string]interface{}{
			"media_type": mediaType,
			"page":       page,
			"limit":      limit,
		}, &out)
		if err == nil {
			return out.Items, nil
		}
		log.Printf("gRPC FetchFeed error (%v), falling back to REST", err)
	}

	url := fmt.Sprintf("%s/api/v1/catalog/feed?type=%s&page=%d&limit=%d", c.cfg.CatalogServiceURL, mediaType, page, limit)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []MediaItem `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// FetchTrending queries catalog-service via gRPC with REST fallback
func (c *Client) FetchTrending(mediaType string) ([]MediaItem, error) {
	conn, err := c.dialGRPC(c.cfg.CatalogServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Items []MediaItem
		}
		err = conn.Invoke(ctx, "/catalog.CatalogService/GetTrending", map[string]interface{}{
			"media_type": mediaType,
		}, &out)
		if err == nil {
			return out.Items, nil
		}
	}

	url := fmt.Sprintf("%s/api/v1/catalog/trending?type=%s", c.cfg.CatalogServiceURL, mediaType)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []MediaItem `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// FetchMedia queries catalog-service or media-service via gRPC
func (c *Client) FetchMedia(id string) (*MediaItem, error) {
	conn, err := c.dialGRPC(c.cfg.CatalogServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Item MediaItem
		}
		err = conn.Invoke(ctx, "/catalog.CatalogService/GetCatalogItem", map[string]string{"id": id}, &out)
		if err == nil && out.Item.ID != "" {
			return &out.Item, nil
		}
	}

	// Try media-service via gRPC
	mediaConn, err := c.dialGRPC(c.cfg.MediaServiceGRPCAddr)
	if err == nil {
		defer mediaConn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Media MediaItem
		}
		err = mediaConn.Invoke(ctx, "/media.MediaService/GetMedia", map[string]string{"id": id}, &out)
		if err == nil && out.Media.ID != "" {
			return &out.Media, nil
		}
	}

	// Fallback to REST
	url := fmt.Sprintf("%s/api/v1/catalog/media/%s", c.cfg.CatalogServiceURL, id)
	resp, err := c.httpClient.Get(url)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		var item MediaItem
		if err := json.NewDecoder(resp.Body).Decode(&item); err == nil {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("media not found")
}

// FetchComments queries catalog-service via gRPC
func (c *Client) FetchComments(id string) ([]Comment, error) {
	conn, err := c.dialGRPC(c.cfg.CatalogServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Comments []Comment
		}
		err = conn.Invoke(ctx, "/catalog.CatalogService/GetComments", map[string]string{"media_id": id}, &out)
		if err == nil {
			return out.Comments, nil
		}
	}

	url := fmt.Sprintf("%s/api/v1/catalog/media/%s/comments", c.cfg.CatalogServiceURL, id)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return []Comment{}, nil
	}
	defer resp.Body.Close()

	var result struct {
		Comments []Comment `json:"comments"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.Comments, nil
}

// RecordView records view via gRPC
func (c *Client) RecordView(id string) (int64, error) {
	conn, err := c.dialGRPC(c.cfg.CatalogServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Views int64
		}
		err = conn.Invoke(ctx, "/catalog.CatalogService/RecordView", map[string]string{"media_id": id}, &out)
		if err == nil {
			return out.Views, nil
		}
	}

	url := fmt.Sprintf("%s/api/v1/catalog/media/%s/view", c.cfg.CatalogServiceURL, id)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Views int64 `json:"views"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.Views, nil
}

// ToggleLike toggles like via gRPC
func (c *Client) ToggleLike(id, userID string) (bool, error) {
	conn, err := c.dialGRPC(c.cfg.CatalogServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Liked bool
		}
		err = conn.Invoke(ctx, "/catalog.CatalogService/ToggleLike", map[string]string{
			"media_id": id,
			"user_id":  userID,
		}, &out)
		if err == nil {
			return out.Liked, nil
		}
	}

	url := fmt.Sprintf("%s/api/v1/catalog/media/%s/like", c.cfg.CatalogServiceURL, id)
	body, _ := json.Marshal(map[string]string{"user_id": userID})
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		Liked bool `json:"liked"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.Liked, nil
}

// AddComment posts comment via gRPC
func (c *Client) AddComment(id, content, userName string) (*Comment, error) {
	conn, err := c.dialGRPC(c.cfg.CatalogServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var out struct {
			Comment Comment
		}
		err = conn.Invoke(ctx, "/catalog.CatalogService/AddComment", map[string]string{
			"media_id":  id,
			"content":   content,
			"user_name": userName,
		}, &out)
		if err == nil && out.Comment.ID != "" {
			return &out.Comment, nil
		}
	}

	url := fmt.Sprintf("%s/api/v1/catalog/media/%s/comments", c.cfg.CatalogServiceURL, id)
	payload := map[string]string{
		"content":   content,
		"user_name": userName,
	}
	body, _ := json.Marshal(payload)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Comment Comment `json:"comment"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Comment, nil
}

// SearchMedia searches OpenSearch via search-service gRPC
func (c *Client) SearchMedia(query, mediaType string, page, limit int) (*SearchResult, error) {
	conn, err := c.dialGRPC(c.cfg.SearchServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var out SearchResult
		err = conn.Invoke(ctx, "/search.SearchService/Search", map[string]interface{}{
			"query":      query,
			"media_type": mediaType,
			"page":       page,
			"limit":      limit,
		}, &out)
		if err == nil {
			return &out, nil
		}
		log.Printf("gRPC Search error (%v), falling back to REST", err)
	}

	url := fmt.Sprintf("%s/api/v1/search?q=%s&type=%s&page=%d&limit=%d", c.cfg.SearchServiceURL, query, mediaType, page, limit)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FetchSuggestions retrieves autocomplete terms via search-service gRPC
func (c *Client) FetchSuggestions(query string, limit int) ([]string, error) {
	conn, err := c.dialGRPC(c.cfg.SearchServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var out struct {
			Suggestions []string
		}
		err = conn.Invoke(ctx, "/search.SearchService/Suggest", map[string]interface{}{
			"query": query,
			"limit": limit,
		}, &out)
		if err == nil {
			return out.Suggestions, nil
		}
	}

	url := fmt.Sprintf("%s/api/v1/search/suggest?q=%s&limit=%d", c.cfg.SearchServiceURL, query, limit)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return []string{}, nil
	}
	defer resp.Body.Close()

	var result struct {
		Suggestions []string `json:"suggestions"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.Suggestions, nil
}

// FetchUser queries user-service via gRPC
func (c *Client) FetchUser(id string) (*User, error) {
	conn, err := c.dialGRPC(c.cfg.UserServiceGRPCAddr)
	if err == nil {
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var out struct {
			User User
		}
		err = conn.Invoke(ctx, "/user.UserService/GetUser", map[string]string{"id": id}, &out)
		if err == nil && out.User.ID != "" {
			return &out.User, nil
		}
	}

	url := fmt.Sprintf("%s/api/v1/users/%s", c.cfg.UserServiceURL, id)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// FetchMe retrieves session user
func (c *Client) FetchMe(token string) (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/me", c.cfg.UserServiceURL), nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unauthorized")
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

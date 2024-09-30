package api

import (
	"encoding/json"
	"fmt"
	"kasm-stress-test/internal/models"
)

// GetUserInfo retrieves user information about a specific user
func (c *Client) GetUserInfo(username string) (*models.User, error) {
	postBody := map[string]interface{}{
		"api_key":        c.config.APIKey,
		"api_key_secret": c.config.APISecret,
		"target_user": map[string]string{
			"username": username,
		},
	}

	body, err := c.apiRequest("/get_user", postBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	var response models.Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info response: %w", err)
	}

	return &response.User, nil
}

// GetUserImages retrieves the images available to a specific user
func (c *Client) GetUserImages(userID string) ([]models.Image, error) {
	postBody := map[string]interface{}{
		"api_key":        c.config.APIKey,
		"api_key_secret": c.config.APISecret,
	}

	body, err := c.apiRequest("get_images", postBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get user images: %w", err)
	}

	var response models.ImageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal images response: %w", err)
	}

	return response.Images, nil
}

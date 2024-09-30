package api

import (
	"encoding/json"
	"fmt"
	"time"

	"kasm-stress-test/internal/models"
	"kasm-stress-test/internal/utils"
)

// RequestKasm creates a new Kasm session
func (c *Client) RequestKasm(userID, imageID string) (*models.Kasm, error) {
	respBody, err := c.apiRequest("request_kasm", map[string]interface{}{
		"user_id":        userID,
		"image_id":       imageID,
		"enable_sharing": false,
	},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to request Kasm: %w", err)
	}

	var kasm models.Kasm
	if err := json.Unmarshal(respBody, &kasm); err != nil {
		utils.Error("Failed to unmarshal Kasm response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal Kasm response: %w", err)
	}

	if kasm.KasmID == "" {
		utils.Error("Kasm ID is empty in the response")
		return nil, fmt.Errorf("received empty Kasm ID from API")
	}

	utils.Info("Successfully created Kasm with ID: %s", kasm.KasmID)

	return &kasm, nil
}

// GetKasmStatus retrieves the status of a Kasm session
func (c *Client) GetKasmStatus(kasmID, user_id string) (*models.KasmStatus, error) {
	respBody, err := c.apiRequest("get_kasm_status", map[string]interface{}{
		"user_id": user_id,
		"kasm_id": kasmID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Kasm status: %w", err)
	}

	var status models.KasmStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kasm status: %w", err)
	}

	return &status, nil
}

// ExecCommand executes a command in a Kasm session
func (c *Client) ExecCommand(kasmID, userID, command string) (*models.CommandResult, error) {
	requestBody := map[string]interface{}{
		"kasm_id": kasmID,
		"user_id": userID,
		"exec_config": map[string]string{
			"cmd": command,
		},
	}

	respBody, err := c.apiRequest("exec_command_kasm", requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var result models.CommandResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command execution response: %w", err)
	}

	return &result, nil
}

// DestroyKasm destroys a Kasm session
func (c *Client) DestroyKasm(kasmID, userID string) error {
	respBody, err := c.apiRequest("destroy_kasm", map[string]interface{}{
		"kasm_id": kasmID,
		"user_id": userID,
	})

	if err != nil {
		return fmt.Errorf("failed to destroy Kasm: %w", err)
	}

	// If the response is empty, it means the Kasm was destroyed successfully
	if len(respBody) == 0 || string(respBody) == "{}" {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse destroy Kasm response: %w", err)
	}

	if errMsg, ok := result["error_message"].(string); ok && errMsg != "" {
		return fmt.Errorf("failed to destroy Kasm: %s", errMsg)
	}

	// If we get here, there was a non-empty response without an error message
	return fmt.Errorf("unexpected response when destroying Kasm: %s", string(respBody))
}

// WaitForKasmReady waits for a Kasm session to be in the "running" state
func (c *Client) WaitForKasmReady(kasmID, image_id string, timeout time.Duration) error {
	start := time.Now()
	for {
		status, err := c.GetKasmStatus(kasmID, image_id)
		if err != nil {
			return fmt.Errorf("failed to get Kasm status: %w", err)
		}

		if status.Kasm.OperationalStatus == "running" {
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for Kasm to be ready")
		}

		time.Sleep(5 * time.Second)
	}
}

// GetAutoscalingStatus retrieves the current autoscaling status
func (c *Client) GetAutoscalingStatus() (*models.AutoscalingStatus, error) {
	respBody, err := c.apiRequest("get_autoscaling_status", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get autoscaling status: %w", err)
	}

	var result models.AutoscalingStatus
	if err := handleAPIResponse(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to handle autoscaling status response: %w", err)
	}

	return &result, nil
}

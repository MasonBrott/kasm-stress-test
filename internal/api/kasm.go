package api

import (
	"encoding/json"
	"fmt"
	"strings"
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
func (c *Client) ExecCommand(kasmID, userID, command string) error {
	requestBody := map[string]interface{}{
		"kasm_id": kasmID,
		"user_id": userID,
		"exec_config": map[string]string{
			"cmd": command,
		},
	}

	_, err := c.apiRequest("exec_command_kasm", requestBody)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// DestroyKasm destroys a Kasm session
func (c *Client) DestroyKasm(kasmID, userID string) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		respBody, err := c.apiRequest("destroy_kasm", map[string]interface{}{
			"kasm_id": kasmID,
			"user_id": userID,
		})

		if err == nil {
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

		utils.Error("Attempt %d to destroy Kasm %s failed: %v. Retrying...", i+1, kasmID, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return fmt.Errorf("failed to destroy Kasm after %d attempts", maxRetries)
}

// WaitForKasmReady waits for a Kasm session to be in the "running" state
func (c *Client) WaitForKasmReady(kasmID, image_id string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		status, err := c.GetKasmStatus(kasmID, image_id)
		if err != nil {
			if strings.Contains(err.Error(), "This session is currently requested") {
				utils.Info("Kasm %s is still in requested state. Waiting...", kasmID)
				time.Sleep(10 * time.Second)
				continue
			}
		}

		if status.Kasm.OperationalStatus == "running" {
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for Kasm to be ready")
		}

		utils.Info("Kasm %s status: %s. Waiting... (%s)", kasmID, status.Kasm.OperationalStatus, time.Since(start))
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("timeout waiting for Kasm %s to be ready", kasmID)
}

// GetAutoscalingStatus retrieves the current autoscaling status
func (c *Client) GetAutoscalingStatus() (*models.AutoscalingStatus, error) {
	// Placeholder for future implementation
	return nil, nil
}

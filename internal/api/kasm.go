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
	requestedTime := time.Time{}
	maxRequestedTime := 3 * time.Minute // Maximum time to wait in "requested" state
	lastNotificationTime := time.Time{}
	notificationInterval := 30 * time.Second

	for time.Since(start) < timeout {
		status, err := c.GetKasmStatus(kasmID, image_id)
		if err != nil {
			utils.Error("Failed to get Kasm status: %v", err)
			time.Sleep(15 * time.Second)
			continue
		}

		if status.Kasm.OperationalStatus == "running" {
			return nil
		}

		if status.ErrorMessage == "This session is currently requested." {
			if requestedTime.IsZero() {
				requestedTime = time.Now()
				utils.Console("Kasm %s is in requested state\n", kasmID)
			} else if time.Since(requestedTime) > maxRequestedTime {
				return fmt.Errorf("Kasm %s stuck in 'requested' state for too long", kasmID)
			}
		} else {
			requestedTime = time.Time{} // Reset if not in "requested" state
		}

		if time.Since(lastNotificationTime) >= notificationInterval {
			utils.Console("Waiting for Kasm %s - Status: Requested, Progress: %d%%, Time elapsed: %s\n",
				kasmID, status.OperationalProgress, time.Since(start).Round(time.Second))
			lastNotificationTime = time.Now()
		}

		utils.Info("Kasm %s status: Requested. Message: %s. Progress: %d%%. Waiting... (%s)",
			kasmID, status.OperationalMessage,
			status.OperationalProgress, time.Since(start))

		time.Sleep(30 * time.Second)
	}
	return fmt.Errorf("timeout waiting for Kasm %s to be ready", kasmID)
}

// GetAutoscalingStatus retrieves the current autoscaling status
func (c *Client) GetAutoscalingStatus() (*models.AutoscalingStatus, error) {
	// Placeholder for future implementation
	return nil, nil
}

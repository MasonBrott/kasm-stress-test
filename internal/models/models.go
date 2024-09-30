package models

import "time"

// User represents a Kasm user
type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// Image represents a Kasm Workspace image
type Image struct {
	FriendlyName string `json:"friendly_name"`
	ImageID      string `json:"image_id"`
}

// Response represents the API response for user info
type Response struct {
	User User `json:"user"`
}

// ImagesResponse represents the API response for user images
type ImageResponse struct {
	Images []Image `json:"images"`
}

// Kasm represents a Kasm session
type Kasm struct {
	KasmID       string `json:"kasm_id"`
	Status       string `json:"status"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	SessionToken string `json:"session_token"`
	KasmURL      string `json:"kasm_url"`
}

// KasmStatus represents the status of a Kasm session
type KasmStatus struct {
	Kasm struct {
		KasmID            string `json:"kasm_id"`
		OperationalStatus string `json:"operational_status"`
		ContainerID       string `json:"container_id"`
		UserID            string `json:"user_id"`
		ImageID           string `json:"image_id"`
		StartDate         string `json:"start_date"`
		ExpirationDate    string `json:"expiration_date"`
		ContainerIP       string `json:"container_ip"`
	} `json:"kasm"`
	CurrentTime string `json:"current_time"`
	KasmURL     string `json:"kasm_url"`
}

// CommandResult represents the result of an executed command
type CommandResult struct {
	KasmID     string `json:"kasm_id"`
	Command    string `json:"command"`
	ExitCode   int    `json:"exit_code"`
	Output     string `json:"output"`
	ExecutedAt string `json:"executed_at"`
}

// StressTestResult represents the result of a stress test
type StressTestResult struct {
	Username         string
	TotalKasms       int
	SuccessfulKasms  int
	FailedKasms      int
	AverageStartTime time.Duration
	TotalDuration    time.Duration
	Errors           []string
	KasmResults      []KasmResult
}

// KasmResult stores individual results for each Kasm instance
type KasmResult struct {
	KasmNumber     int
	KasmID         string
	StartTime      time.Duration
	ExecutionError string
}

// AutoscalingStatus represents the status of the autoscaling system
type AutoscalingStatus struct {
	CurrentNodes int     `json:"current_nodes"`
	DesiredNodes int     `json:"desired_nodes"`
	PendingNodes int     `json:"pending_nodes"`
	MaxNodes     int     `json:"max_nodes"`
	CurrentLoad  float64 `json:"current_load"`
}

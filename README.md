# Kasm Stress Test Tool

This tool is designed to stress test Kasm Workspaces by creating multiple Kasm instances, executing commands, and then destroying them. It's particularly useful for testing the autoscaling capabilities of a Kasm deployment.

## Features

- Create multiple Kasm instances for a specified user
- Execute commands on each Kasm instance
- Test Kasm's autoscaling capabilities
- Detailed logging and error reporting
- Configurable via command-line flags and configuration file

## Prerequisites

- Go 1.16 or higher
- Access to a Kasm Workspaces deployment
- API key and secret for the Kasm API

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/kasm-stress-test.git
   ```
2. Navigate to the project directory:
   ```
   cd kasm-stress-test
   ```
3. Build the project:
   ```
   go build -o kasm-stress-test cmd/kasm-stress-test/main.go
   ```

## Configuration

Create a `.kasm-stress-test.json` file in your home directory with the following structure:

```
json
{
"api_key": "your-api-key",
"api_secret": "your-api-secret",
"api_host": "https://your-kasm-host.com/api/public",
"default_image_id": "your-default-image-id",
"log_level": "info",
"timeout_seconds": 300
}
```

## Usage

Run the stress test with the following command:

```
./kasm-stress-test -u username@example.com -kasm-range 1-5
```

Command-line flags:
- `-u`: Username to use for the test (can be specified multiple times for multiple users)
- `-kasm-range`: Range of Kasm instances to create (e.g., 1-5 will create 5 instances)

## Output

The tool will provide detailed output about each Kasm instance created, including:
- Start time
- Execution status
- Any errors encountered

At the end of the test, a summary will be displayed showing:
- Total number of Kasm instances created
- Number of successful and failed instances
- Average start time
- Total test duration
# Bulk Mailer

A simple bulk email sending service with rate limiting and a web API for control.

## Features

*   Bulk email sending from a CSV file.
*   Rate limiting to control the sending speed.
*   Web API to start, stop, pause, resume, and monitor the sending process.
*   Asynchronous task processing using Asynq.
*   Redis for task queuing and state management.

## Getting Started

### Prerequisites

*   Go 1.18 or higher
*   Redis

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/your-username/bulkmailer.git
    ```
2.  Change into the project directory:
    ```bash
    cd bulkmailer
    ```
3.  Install dependencies:
    ```bash
    go mod tidy
    ```

## Usage

1.  **Run the application:**

    ```bash
    go run cmd/main.go
    ```

    By default, the application will connect to a Redis server at `127.0.0.1:6379` and look for a `recipients.csv` file in the project root.

2.  **Use the API endpoints:**

    *   `GET /start`: Starts the email sending process.
    *   `GET /stop`: Stops the email sending process.
    *   `GET /pause`: Pauses the email sending process.
    *   `GET /resume`: Resumes the email sending process.
    *   `GET /monitor`: Shows the current status of the email sending process.

## Configuration

The application can be configured using the following environment variables. If an environment variable is not set, the default value will be used.

*   `REDIS_ADDR`: The address of the Redis server (default: `127.0.0.1:6379`).
*   `CSV_PATH`: The path to the CSV file containing the recipient information (default: `recipients.csv`).
*   `HTTP_ADDR`: The address for the HTTP control API to listen on (default: `:8080`).

### CSV File Format

The CSV file should have the following format:

```csv
email,subject,body
test1@example.com,Hello,This is a test email.
test2@example.com,Another one,This is another test email.
```

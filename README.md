## Key Implementation Details

### Configuration

- For simplicity, this project does not use external configuration files.
- Server settings (like port number) are hardcoded in the `main.go` file.

### Persistence

- The server uses a hybrid approach for data storage:
  1. In-memory map for fast read and write operations.
  2. File-based backend for persistence across server restarts.

### Concurrency Control

- The project approach to manage concurrent access:
  - A `sync.RWMutex` in `fileDB` protects access to the in-memory map.
  - Named mutexes (stored in a `sync.Map`) prevent duplicate command executions.

## Usage

1. Run the server:
   ```
   go run *.go
   ```

2. Send POST requests to `http://localhost:8080/cmd` with JSON payloads:
   ```json
   {
     "cmd": "example",
     "id": 12
   }
   ```
   
## NOTE: to test concurrency change sleep to bigger value, but less that server read timeout.
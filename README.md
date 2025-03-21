```markdown
# Download Manager

A powerful, terminal-based download manager written in Go, featuring multi-part downloads, bandwidth limiting, and an interactive TUI.

## Design Overview

This download manager is built with a modular architecture that separates concerns between user interface, download management, queue management, and bandwidth control. The application leverages Go's concurrency model to handle multiple downloads simultaneously while providing a responsive terminal user interface.

### Core Components

#### User Interface (`internal/ui/`)

The user interface is built using [Bubble Tea](https://github.com/charmbracelet/bubbletea), a Go framework for building terminal user interfaces based on The Elm Architecture. The UI provides:

- Interactive download management
- Real-time progress visualization
- Queue management
- Bandwidth control settings
- Error notifications and status updates

The UI is designed with a component-based approach, where different views (add download, queue management, settings, etc.) are implemented as separate components that can be composed together.

#### Download Engine (`internal/downloads/`)

The download engine is responsible for:

- Handling HTTP requests and responses
- Supporting multi-part downloads through HTTP Range headers
- Managing download chunks and reassembling them
- Handling connection errors and timeouts
- Tracking download progress

Multi-part downloading is implemented by:

1. Making an initial HEAD request to check if the server supports range requests (`Accept-Ranges` header)
2. Dividing the download into multiple chunks
3. Creating separate goroutines for each chunk
4. Downloading chunks in parallel
5. Writing chunks to the correct file offsets using a synchronized file writer

#### Queue Management (`internal/queues/`)

The queue management system provides:

- CRUD operations for download queues
- Prioritization of downloads
- Persistence of queue state
- Scheduling downloads based on queue priority
- Pausing and resuming queues

The Queue Manager acts as the coordinator between the UI and the download engine, ensuring that user commands are properly executed and that download state is maintained.

#### Bandwidth Control (`internal/bandwidthlimit/`)

Bandwidth limiting is implemented using Go's `golang.org/x/time/rate` package, which provides a token bucket rate limiter. This allows:

- Setting global bandwidth limits
- Per-download bandwidth allocation
- Dynamic adjustment of bandwidth limits
- Fair sharing of available bandwidth between active downloads

The bandwidth limiter wraps standard IO operations to enforce the configured rate limits without blocking the UI.

### Data Flow

1. User initiates a download through the UI
2. The request is passed to the Queue Manager
3. Queue Manager creates a download entry and assigns it to a queue
4. Download Engine checks if the server supports range requests
5. If ranges are supported, the download is split into multiple chunks
6. Each chunk is downloaded in parallel, subject to bandwidth limits
7. Progress is reported back to the UI in real-time
8. Completed chunks are assembled into the final file
9. Queue Manager updates the download status

### Error Handling

The download manager implements robust error handling:

- Network disconnections are detected and reported
- Failed downloads can be resumed from the point of failure
- Timeouts are configurable to quickly detect connection issues
- Detailed error reporting helps diagnose download problems

### Events System

An event-driven architecture allows components to communicate without tight coupling:

- UI components subscribe to relevant events
- Download Engine publishes progress and error events
- Queue Manager publishes state change events
- Events are processed asynchronously to keep the UI responsive

## Technical Details

- **Language**: Go
- **UI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Bandwidth Control**: `golang.org/x/time/rate`
- **Concurrency**: Go routines and channels
- **File I/O**: Synchronized file access for concurrent chunk writing
- **HTTP Client**: Custom-configured HTTP client with timeout settings
```

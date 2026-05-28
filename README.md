# NAS AI Advisor

An intelligent AI-powered advisor system for NAS operating systems, inspired by Synology's AI Advisor functionality.

## Features

- **Natural Language Q&A**: Ask questions about system features and configurations in natural language
- **Intelligent Recommendations**: Get personalized recommendations based on your usage patterns
- **Fault Diagnosis**: Diagnose common system issues with AI assistance
- **Knowledge Base**: Searchable knowledge base with articles on various topics
- **Context-Aware Conversations**: Maintains conversation context for follow-up questions
- **Web Interface**: Modern, responsive web UI for easy interaction
- **REST API**: Full-featured API for integration with other systems
- **Docker Support**: Easy deployment with Docker and Docker Compose

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Web Interface                          │
│                     (HTML/CSS/JS)                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      REST API Layer                         │
│                    (Gin Framework)                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     AI Advisor Core                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Query       │  │  Recommend   │  │  Diagnose    │      │
│  │  Handler     │  │  Engine      │  │  Engine      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│   LLM Client     │ │  Knowledge Base  │ │   Conversation   │
│   (Ollama)       │ │  (JSON Files)    │ │   Manager        │
└──────────────────┘ └──────────────────┘ └──────────────────┘
```

## Prerequisites

- Go 1.21 or higher
- Ollama (for local LLM inference)
- Docker (optional, for containerized deployment)

## Quick Start

### 1. Install Ollama

```bash
# macOS/Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull llama2
```

### 2. Clone and Build

```bash
git clone https://github.com/yourusername/nas-ai-advisor.git
cd nas-ai-advisor

# Build the application
go build -o server ./cmd/server
```

### 3. Configure

Create a `config.json` file:

```json
{
  "llm": {
    "base_url": "http://localhost:11434",
    "model": "llama2",
    "timeout": 30,
    "max_tokens": 2048,
    "temperature": 0.7
  },
  "knowledge_path": "./knowledge",
  "max_messages": 50,
  "cleanup_interval": 60,
  "max_conversation_age": 24
}
```

### 4. Run

```bash
./server -config config.json
```

The server will start on `http://localhost:8080`.

### 5. Access Web Interface

Open your browser and navigate to `http://localhost:8080`.

## Docker Deployment

### Using Docker Compose

```bash
docker-compose up -d
```

This will start both the AI Advisor and Ollama services.

### Manual Docker Build

```bash
# Build the image
docker build -t nas-ai-advisor .

# Run the container
docker run -d \
  -p 8080:8080 \
  -v ./config:/app/config \
  -v ./knowledge:/app/knowledge \
  --name nas-ai-advisor \
  nas-ai-advisor
```

## API Usage

### Query

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I set up RAID 5?",
    "user_id": "user123"
  }'
```

### Get Recommendations

```bash
curl -X POST http://localhost:8080/api/v1/recommendations \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "limit": 5
  }'
```

### Run Diagnostics

```bash
curl -X POST http://localhost:8080/api/v1/diagnose \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "symptoms": ["slow performance", "disk errors"]
  }'
```

### System Status

```bash
curl http://localhost:8080/api/v1/status
```

## Knowledge Base

The knowledge base is stored in the `knowledge/` directory as JSON files. Each file represents an article:

```json
{
  "id": "article-id",
  "title": "Article Title",
  "content": "Article content in markdown...",
  "category": "category-name",
  "tags": ["tag1", "tag2"]
}
```

### Adding Articles

1. Create a new JSON file in the `knowledge/` directory
2. Follow the article format above
3. Restart the server or use the API to add articles

### Categories

- `storage`: Storage management, RAID, disks
- `network`: Network configuration, VPN, remote access
- `backup`: Backup strategies, snapshots, cloud backup
- `docker`: Docker containers, applications
- `security`: Security hardening, permissions, firewall
- `performance`: Performance optimization, monitoring

## Development

### Project Structure

```
nas-ai-advisor/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   └── aiadvisor/
│       ├── advisor.go        # Core advisor logic
│       ├── api.go            # REST API handlers
│       ├── conversation.go   # Conversation management
│       ├── knowledge.go      # Knowledge base management
│       ├── llm.go            # LLM client integration
│       ├── models.go         # Data models
│       └── advisor_test.go   # Unit tests
├── knowledge/                # Knowledge base articles
├── web/
│   └── static/
│       └── index.html        # Web interface
├── docs/
│   └── API.md                # API documentation
├── config.json               # Configuration file
├── Dockerfile                # Docker build file
├── docker-compose.yml        # Docker Compose configuration
├── go.mod                    # Go module file
└── README.md                 # This file
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./internal/aiadvisor/

# Run specific test
go test -run TestQuery ./internal/aiadvisor/
```

### Building for Production

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Build for ARM (Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o server ./cmd/server
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `llm.base_url` | string | `http://localhost:11434` | Ollama API base URL |
| `llm.model` | string | `llama2` | LLM model to use |
| `llm.timeout` | int | `30` | Request timeout in seconds |
| `llm.max_tokens` | int | `2048` | Maximum tokens in response |
| `llm.temperature` | float | `0.7` | Response creativity (0.0-1.0) |
| `knowledge_path` | string | `./knowledge` | Path to knowledge base directory |
| `max_messages` | int | `50` | Maximum messages per conversation |
| `cleanup_interval` | int | `60` | Cleanup interval in minutes |
| `max_conversation_age` | int | `24` | Maximum conversation age in hours |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Synology AI Advisor
- Built with [Gin](https://github.com/gin-gonic/gin) web framework
- Uses [Ollama](https://ollama.ai/) for local LLM inference

## Support

For support, please open an issue on GitHub or contact the maintainers.

## Roadmap

- [ ] Streaming responses for real-time interaction
- [ ] Multi-language support
- [ ] Voice input/output
- [ ] Integration with NAS monitoring systems
- [ ] Custom knowledge base editor
- [ ] User authentication and authorization
- [ ] Rate limiting and abuse prevention
- [ ] WebSocket support for real-time updates
- [ ] Mobile application
- [ ] Plugin system for extensibility

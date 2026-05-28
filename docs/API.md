# NAS AI Advisor API Documentation

## Overview

The NAS AI Advisor provides a REST API for interacting with the AI-powered assistant system. This API enables natural language queries, intelligent recommendations, and system diagnostics.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. In production, implement proper authentication mechanisms.

## Endpoints

### Health Check

**GET** `/health`

Check the health status of the API service.

#### Response

```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

---

### Query

**POST** `/query`

Send a natural language query to the AI advisor.

#### Request Body

```json
{
  "query": "How do I set up RAID 5?",
  "conversation_id": "optional-conversation-id",
  "user_id": "user123"
}
```

#### Response

```json
{
  "answer": "To set up RAID 5, follow these steps...",
  "conversation_id": "conv-abc123",
  "sources": ["Storage Management Guide"],
  "suggestions": ["How to monitor RAID health?", "What is RAID 6?"]
}
```

#### Error Responses

- **400 Bad Request**: Invalid request body
- **500 Internal Server Error**: Server processing error

---

### Recommendations

**POST** `/recommendations`

Get personalized recommendations based on user history.

#### Request Body

```json
{
  "user_id": "user123",
  "limit": 5
}
```

#### Response

```json
{
  "recommendations": [
    {
      "id": "rec-storage-pool",
      "title": "Optimize Storage Pool",
      "description": "Based on your interest in storage, consider optimizing your storage pool configuration for better performance.",
      "category": "storage",
      "relevance": 0.8,
      "action": "Navigate to Storage Manager"
    }
  ]
}
```

---

### Diagnose

**POST** `/diagnose`

Run diagnostics on system issues.

#### Request Body

```json
{
  "user_id": "user123",
  "symptoms": ["slow performance", "high CPU usage"],
  "system_info": "OS: Linux\nCPU: 4 cores\nRAM: 8GB"
}
```

#### Response

```json
{
  "issue": "System diagnostic analysis",
  "severity": "medium",
  "causes": ["See detailed analysis below"],
  "solutions": ["The system is experiencing slow performance..."]
}
```

#### Severity Levels

- `low`: Minor issues, no immediate action required
- `medium`: Moderate issues, recommended to address soon
- `high`: Significant issues, should be addressed promptly
- `critical`: Critical issues, immediate action required

---

### System Status

**GET** `/status`

Get the current system status.

#### Response

```json
{
  "status": "running",
  "knowledge_ok": true,
  "llm_ok": true,
  "active_conversations": 5,
  "uptime": "2h30m15s"
}
```

---

### Knowledge Base

#### List Articles

**GET** `/knowledge/articles`

Get all articles in the knowledge base.

#### Response

```json
{
  "articles": [
    {
      "id": "storage-management",
      "title": "Storage Management Guide",
      "content": "## Storage Pool Management...",
      "category": "storage",
      "tags": ["storage", "raid", "disk"]
    }
  ]
}
```

#### Get Article

**GET** `/knowledge/articles/:id`

Get a specific article by ID.

#### Response

```json
{
  "id": "storage-management",
  "title": "Storage Management Guide",
  "content": "## Storage Pool Management...",
  "category": "storage",
  "tags": ["storage", "raid", "disk"]
}
```

#### Create Article

**POST** `/knowledge/articles`

Create a new knowledge base article.

#### Request Body

```json
{
  "id": "new-article",
  "title": "New Article Title",
  "content": "Article content here...",
  "category": "general",
  "tags": ["tag1", "tag2"]
}
```

#### Response

```json
{
  "id": "new-article",
  "title": "New Article Title",
  "content": "Article content here...",
  "category": "general",
  "tags": ["tag1", "tag2"]
}
```

#### Delete Article

**DELETE** `/knowledge/articles/:id`

Delete an article from the knowledge base.

#### Response

```json
{
  "message": "Article deleted"
}
```

#### Get Categories

**GET** `/knowledge/categories`

Get all unique categories in the knowledge base.

#### Response

```json
{
  "categories": ["storage", "network", "backup", "docker", "security"]
}
```

---

### Conversations

#### Get Conversation

**GET** `/conversations/:id`

Get a conversation by ID.

#### Response

```json
{
  "id": "conv-abc123",
  "user_id": "user123",
  "messages": [
    {
      "id": "msg-1",
      "role": "user",
      "content": "How do I set up RAID?",
      "timestamp": "2024-01-15T10:30:00Z"
    },
    {
      "id": "msg-2",
      "role": "assistant",
      "content": "To set up RAID...",
      "timestamp": "2024-01-15T10:30:05Z"
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:05Z"
}
```

#### Get Conversation Messages

**GET** `/conversations/:id/messages`

Get all messages in a conversation.

#### Response

```json
{
  "messages": [
    {
      "id": "msg-1",
      "role": "user",
      "content": "How do I set up RAID?",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ]
}
```

#### Delete Conversation

**DELETE** `/conversations/:id`

Delete a conversation.

#### Response

```json
{
  "message": "Conversation deleted"
}
```

## Error Handling

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

## Rate Limiting

Currently, there are no rate limits implemented. In production, consider implementing rate limiting to prevent abuse.

## CORS

The API supports Cross-Origin Resource Sharing (CORS) with the following headers:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

## WebSocket Support

WebSocket support is not currently implemented. Consider adding WebSocket support for real-time streaming responses in future versions.

## Examples

### cURL Examples

#### Send a Query

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I set up a backup?",
    "user_id": "user123"
  }'
```

#### Get Recommendations

```bash
curl -X POST http://localhost:8080/api/v1/recommendations \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "limit": 3
  }'
```

#### Run Diagnostics

```bash
curl -X POST http://localhost:8080/api/v1/diagnose \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "symptoms": ["slow performance", "disk errors"],
    "system_info": "OS: Linux\nCPU: 4 cores"
  }'
```

#### Get System Status

```bash
curl http://localhost:8080/api/v1/status
```

## SDKs and Libraries

### JavaScript/Node.js

```javascript
const axios = require('axios');

const API_BASE = 'http://localhost:8080/api/v1';

async function query(question, userId) {
  const response = await axios.post(`${API_BASE}/query`, {
    query: question,
    user_id: userId
  });
  return response.data;
}

// Usage
query('How do I set up RAID?', 'user123')
  .then(response => console.log(response.answer))
  .catch(error => console.error(error));
```

### Python

```python
import requests

API_BASE = 'http://localhost:8080/api/v1'

def query(question, user_id):
    response = requests.post(f'{API_BASE}/query', json={
        'query': question,
        'user_id': user_id
    })
    return response.json()

# Usage
response = query('How do I set up RAID?', 'user123')
print(response['answer'])
```

## Deployment

### Docker

```bash
docker-compose up -d
```

### Manual Deployment

1. Build the application:
   ```bash
   go build -o server ./cmd/server
   ```

2. Run the server:
   ```bash
   ./server -config config.json
   ```

## Configuration

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

## Troubleshooting

### Common Issues

1. **LLM Service Unavailable**
   - Check if Ollama is running
   - Verify the base_url in config
   - Check network connectivity

2. **Knowledge Base Not Loading**
   - Verify knowledge_path exists
   - Check file permissions
   - Ensure JSON files are valid

3. **High Memory Usage**
   - Reduce max_messages
   - Decrease max_conversation_age
   - Monitor with system status endpoint

## Support

For issues and questions, please open an issue on the GitHub repository.

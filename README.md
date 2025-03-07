# OpenAI API Mock Service

A tiny mock service that emulates OpenAI's API for embeddings and chat completions. 

Usefull to complex pipelines and flows load testing.

100% compatible with the OpenAI API and allows configurable response times.

## Features

- Fully compatible with OpenAI API endpoints
- Supports both embeddings and chat completions
- Configurable response latency (min/max)
- Supports multiple models
- Simple configuration via JSON file or command-line flags

## Installation

### Prerequisites

- Go 1.16 or higher (for building from source)
- Docker and Docker Compose (for Docker deployment)

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/aimock.git
cd aimock

# Build the binary
go build -o aimock

# Run the service
./aimock
```

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build and run with Docker directly
docker build -t aimock .
docker run -p 8080:8080 aimock
```

## Usage

### Starting the service

```bash
# Run with default settings (port 8080, no latency)
./aimock

# Run with custom port
./aimock -port 3000

# Run with simulated latency
./aimock -min-latency 100ms -max-latency 300ms

# Run with configuration file
./aimock -config config.json
```

### Configuration

You can configure the service using a JSON configuration file:

```json
{
  "port": 8080,
  "min_latency": "200ms",
  "max_latency": "500ms",
  "models": {
    "embedding": ["text-embedding-ada-002", "text-embedding-3-small", "text-embedding-3-large"],
    "chat": ["gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"]
  }
}
```

### API Endpoints

The service implements the following OpenAI API endpoints:

#### Chat Completions

```
POST /v1/chat/completions
```

Example request:

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Hello!"
    }
  ]
}
```

#### Embeddings

```
POST /v1/embeddings
```

Example request:

```json
{
  "model": "text-embedding-ada-002",
  "input": ["The food was delicious and the service was excellent."]
}
```

#### Models

```
GET /v1/models
```

## Using with OpenAI Client Libraries

You can use this mock service with any OpenAI client library by setting the base URL to point to your mock service:

### Python

```python
import openai

client = openai.OpenAI(
    api_key="dummy-api-key",  # Any value works
    base_url="http://localhost:8080/v1"  # Point to your mock service
)

# Chat completion
completion = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello!"}
    ]
)
print(completion.choices[0].message.content)

# Embeddings
embedding = client.embeddings.create(
    model="text-embedding-ada-002",
    input="The food was delicious and the service was excellent."
)
print(embedding.data[0].embedding[:5])  # Print first 5 values
```

### Node.js

```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: 'dummy-api-key', // Any value works
  baseURL: 'http://localhost:8080/v1', // Point to your mock service
});

// Chat completion
const completion = await openai.chat.completions.create({
  model: 'gpt-3.5-turbo',
  messages: [
    { role: 'system', content: 'You are a helpful assistant.' },
    { role: 'user', content: 'Hello!' },
  ],
});
console.log(completion.choices[0].message.content);

// Embeddings
const embedding = await openai.embeddings.create({
  model: 'text-embedding-ada-002',
  input: 'The food was delicious and the service was excellent.',
});
console.log(embedding.data[0].embedding.slice(0, 5)); // Print first 5 values
```

## License

MIT
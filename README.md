# Micro A2A

Experimental repo for A2A support within Micro

## Example

Request

```bash
# non-stream
curl -X POST http://localhost:8081/AgentDebo -d '{"jsonrpc":"2.0","id":1,"method":"tasks/send","params":{"id":"de38c76d-d54c-436c-8b9f-4c2703648d64","message":{"role":"user","parts":[{"type":"text","text":"what is the current time"}]},"metadata":{}}}' | jq
# stream
curl -X GET http://localhost:8081/AgentDebo/stream -d '{"jsonrpc":"2.0","id":1,"method":"tasks/sendSubscribe","params":{"id":"de38c76d-d54c-436c-8b9f-4c2703648d64","message":{"role":"user","parts":[{"type":"text","text":"what is the current time"}]},"metadata":{}}}' | jq
```

Response
```bash
# non-stream
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "id": "5d5d6b49-185a-4899-a2ad-9b679db4265d",
    "sessionId": "55085a75-f51f-42e9-a99e-84a2a7adcf54",
    "status": {
      "state": "completed"
    },
    "artifacts": [
      {
        "name": "what is the current time",
        "parts": [
          {
            "type": "text",
            "text": "2025-05-09 20:59:13.053657194 +0300 MSK m=+274.850514452"
          }
        ],
        "index": 0
      }
    ]
  }
}

# stream
{
  "data": {
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
      "id": "0aba5e6d-738e-4b5b-8877-ebe52c8c774e",
      "artifact": {
        "name": "time ticks every 1 second",
        "parts": [
          {
            "type": "text",
            "text": "2025-05-10 17:16:20.988482028 +0300 MSK m=+77.666015930"
          }
        ],
        "index": 0
      }
    }
  }
}
{
  "data": {
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
      "id": "e2c5f0e6-caf3-41ee-8d84-c233b2ed3e33",
      "artifact": {
        "name": "time ticks every 1 second",
        "parts": [
          {
            "type": "text",
            "text": "2025-05-10 17:16:21.987688392 +0300 MSK m=+78.665222272"
          }
        ],
        "index": 0
      }
    }
  }
}
```

# Stashr

An in-memory key/value store with TTL support, exposed via HTTP/REST and gRPC.

## Building

```bash
go build ./cmd/stashr
```

## Running

```bash
./stashr
```

The server starts two listeners:

| Protocol | Address |
|----------|---------|
| HTTP     | `:8080` |
| gRPC     | `:9090` |

Stop with `Ctrl+C` for graceful shutdown.

## HTTP/REST API

### Set a key

```
PUT /keys/{key}
Content-Type: application/json

{"value": "...", "ttl_seconds": 60}
```

`ttl_seconds` is optional. Omit it or set to `0` for no expiration.

### Get a key

```
GET /keys/{key}
```

Returns `200` with `{"value": "..."}` or `404` if not found.

### Delete a key

```
DELETE /keys/{key}
```

Returns `{"deleted": true}` or `{"deleted": false}`.

---

## gRPC API

The service is defined in `proto/stashr.proto` and exposes three RPCs:

| RPC    | Request fields             | Response fields      |
|--------|----------------------------|----------------------|
| Get    | `key`                      | `value`, `found`     |
| Set    | `key`, `value`, `ttl_seconds` | _(empty)_         |
| Delete | `key`                      | `deleted`            |

gRPC server reflection is enabled, so tools like `grpcurl` work out of the box.

---

## Usage Examples

### curl

```bash
# set a key (no expiry)
curl -X PUT http://localhost:8080/keys/greeting \
  -H 'Content-Type: application/json' \
  -d '{"value": "hello world"}'

# set a key with 30-second TTL
curl -X PUT http://localhost:8080/keys/session \
  -H 'Content-Type: application/json' \
  -d '{"value": "abc123", "ttl_seconds": 30}'

# get a key
curl http://localhost:8080/keys/greeting
# => {"value":"hello world"}

# delete a key
curl -X DELETE http://localhost:8080/keys/greeting
# => {"deleted":true}
```

### grpcurl

```bash
# set
grpcurl -plaintext -d '{"key":"color","value":"blue"}' \
  localhost:9090 stashr.KVStore/Set

# get
grpcurl -plaintext -d '{"key":"color"}' \
  localhost:9090 stashr.KVStore/Get
# => {"value": "blue", "found": true}

# delete
grpcurl -plaintext -d '{"key":"color"}' \
  localhost:9090 stashr.KVStore/Delete
# => {"deleted": true}
```

### Python (HTTP)

```python
import requests

base = "http://localhost:8080/keys"

# set a key with a 60-second TTL
requests.put(f"{base}/user:42", json={"value": "Alice", "ttl_seconds": 60})

# get
resp = requests.get(f"{base}/user:42")
if resp.ok:
    print(resp.json()["value"])  # Alice

# delete
resp = requests.delete(f"{base}/user:42")
print(resp.json()["deleted"])  # True
```

### Python (gRPC)

Install the client libraries first:

```bash
pip install grpcio grpcio-tools
python -m grpc_tools.protoc -Iproto --python_out=. --grpc_python_out=. proto/stashr.proto
```

This generates `stashr_pb2.py` and `stashr_pb2_grpc.py` in the current directory.

```python
import grpc
import stashr_pb2
import stashr_pb2_grpc

channel = grpc.insecure_channel("localhost:9090")
stub = stashr_pb2_grpc.KVStoreStub(channel)

# set
stub.Set(stashr_pb2.SetRequest(key="lang", value="python", ttl_seconds=120))

# get
resp = stub.Get(stashr_pb2.GetRequest(key="lang"))
if resp.found:
    print(resp.value)  # python

# delete
resp = stub.Delete(stashr_pb2.DeleteRequest(key="lang"))
print(resp.deleted)  # True

channel.close()
```

### Go (HTTP)

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	base := "http://localhost:8080/keys"

	// set
	body, _ := json.Marshal(map[string]any{"value": "world", "ttl_seconds": 60})
	req, _ := http.NewRequest(http.MethodPut, base+"/hello", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	// get
	resp, _ := http.Get(base + "/hello")
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	fmt.Println(result["value"]) // world

	// delete
	req, _ = http.NewRequest(http.MethodDelete, base+"/hello", nil)
	http.DefaultClient.Do(req)
}
```

### Go (gRPC)

```go
package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "stashr/pb"
)

func main() {
	conn, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewKVStoreClient(conn)
	ctx := context.Background()

	// set with 2-minute TTL
	client.Set(ctx, &pb.SetRequest{Key: "color", Value: "green", TtlSeconds: 120})

	// get
	resp, _ := client.Get(ctx, &pb.GetRequest{Key: "color"})
	if resp.Found {
		fmt.Println(resp.Value) // green
	}

	// delete
	del, _ := client.Delete(ctx, &pb.DeleteRequest{Key: "color"})
	fmt.Println(del.Deleted) // true
}
```

## Testing

```bash
go test ./store/... -v
```

## Project Structure

```
stashr/
├── cmd/stashr/main.go     # entry point, starts HTTP + gRPC servers
├── proto/stashr.proto      # gRPC service definition
├── pb/                     # generated protobuf Go code
├── store/store.go          # core in-memory store with TTL
├── store/store_test.go     # unit tests
├── server/http.go          # REST handler (stdlib router)
└── server/grpc.go          # gRPC server implementation
```

## Dan's Note

This project is shamelessly vibe-coded as a means of testing Claude. I am pretty impressed with what it was capable of doing, and after reviewing all of the code I can state that this is a pretty legitimate product. I may iterate on this as needed.
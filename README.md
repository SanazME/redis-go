# redis-go
This is build your own Redis project which creates a Redis server:
- A TCP server that bind on and listen to port 6379 (the default port that Redis uses) 
- Handle concurrent Redis clients
- Implement RESP protocol to support the following commands:
  - [`PING`](https://redis.io/commands/ping/)
  - [`ECHO`](https://redis.io/commands/echo/)
  - [`SET`](https://redis.io/commands/set/) with support for `Expiry` flag
  - [`GET`](https://redis.io/commands/get/)

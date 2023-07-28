### Go chat

A basic single-channel websocket server for communication between connected clients.

## Server

Run with `make start`.

By default it starts on port `4001`, which can be overridden with `GO_CHAT_PORT`.

## Client

For testing only!

Run with `make test`, insert name and start writing messages. Message is sent on pressing Enter.

By default it uses `localhost:4001` as host, which can be overriden with `GO_CHAT_SERVER_HOST`.

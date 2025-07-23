// Package websocket is a lightweight yet robust client-only
// implementation of the WebSocket protocol (RFC 6455).
//
// It focuses on continuous asynchronous reading of text/binary
// messages, and enables occasional writing.
//
// It is designed primarily for ease of use and availability at scale.
// Additional design goals: reliability, maintainability, and efficiency.
//
// How does this package optimize for availability at scale?
//  1. In-memory map of active clients, keyed by (a secure hash of)
//     their ID, to minimize the number of open connections per app
//  2. Preemptively switch connections before each anticipated
//     disconnection, to prevent downtime during reconnections
//  3. Fast detection and recovery from unexpected disconnections
//  4. Idiomatic, minimalistic, and modern code patterns
//
// Note A: optimization 1 relies on Go channels to dispatch and
// potentially fan-out messages efficiently and reliably.
//
// Note B: optimization 2 requires careful balancing of optimization 1
// with ensuring state isolation, correct and efficient garbage collection,
// and ensuring that users of this package do not receive duplicate copies
// of messages while a client temporarily has an extra connection.
//
// Note C: WebSocket [extensions] and [subprotocols] are not supported yet.
//
// [extensions]: https://www.iana.org/assignments/websocket/websocket.xhtml#extension-name
// [subprotocols]: https://www.iana.org/assignments/websocket/websocket.xhtml#subprotocol-name
package websocket

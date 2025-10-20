# Timpani

[![Go Reference](https://pkg.go.dev/badge/github.com/tzrikka/timpani.svg)](https://pkg.go.dev/github.com/tzrikka/timpani)
[![Go Report Card](https://goreportcard.com/badge/github.com/tzrikka/timpani)](https://goreportcard.com/report/github.com/tzrikka/timpani)

Timpani is a [Temporal](https://temporal.io/) worker that sends API calls and receives asynchronous event notifications to/from various well-known third-party services.

API calls are wrapped and exposed as Temporal [activities](https://docs.temporal.io/activities) for robustness and durable execution in larger workflows.

Event listeners are similarly reliable and scalable, and support multiple technologies: HTTP webhooks, [WebSocket](https://en.wikipedia.org/wiki/WebSocket) connections, and [Pub/Sub](https://cloud.google.com/pubsub/docs/overview) subscriptions. They may be passive and stateless receivers with a static configuration on the remote service's side, or semi-active subscribers that renew their subscription from time to time, or stateful clients maintaining a 2-way streaming connection with the remote service.

For example:

- Discord: WebSocket client
- Gmail: Google Cloud Pub/Sub subscriber
- Jira: stateful HTTP webhook (with periodic subscription renewals)
- Slack: stateless HTTP webhook / WebSocket client

## Dependencies

- [Temporal](https://temporal.io/)
- [Thrippy](https://github.com/tzrikka/thrippy)
  (preferably with a secrets manager, e.g. [HashiCorp Vault](https://developer.hashicorp.com/vault))

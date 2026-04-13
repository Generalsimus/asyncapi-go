<div align="center">
  <h1>🚀 AsyncAPI-Go</h1>
  <p><b>Code-First AsyncAPI v3 Documentation & Embedded UI for Go</b></p>
  
  <a href="https://pkg.go.dev/github.com/Generalsimus/asyncapi-go"><img src="https://pkg.go.dev/badge/github.com/Generalsimus/asyncapi-go.svg" alt="Go Reference"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://www.asyncapi.com/docs/specifications/v3.0.0"><img src="https://img.shields.io/badge/AsyncAPI-v3.0.0-45D093?logo=asyncapi" alt="AsyncAPI v3.0.0"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.18%2B-00ADD8?logo=go" alt="Go 1.18+"></a>
</div>

<br/>

Writing and maintaining massive AsyncAPI YAML files by hand is tedious and error-prone. In fast-moving event-driven architectures, documentation often drifts from the actual codebase.

**AsyncAPI-Go** solves this. It is a zero-dependency library that uses Go reflection to instantly generate valid **AsyncAPI v3** documentation directly from your Go structs. Better yet, it includes a built-in HTTP handler to serve the beautiful, interactive `@asyncapi/react-component` UI right from your microservice.

No YAML. No drift. Just your code.

<!-- For maximum impact, add a screenshot of the generated UI here! -->
<!-- !AsyncAPI-Go UI Screenshot -->

## ✨ Why use this?

*   **Code-First Generation:** Automatically maps Go primitives, `time.Time`, and complex nested structs to valid JSON Schema.
*   **Embedded UI Server:** Serve world-class API documentation directly from your Go app with a single line of code.
*   **Smart Struct Tags:** Extracts clean `description` tags for the UI, while preserving critical backend tags (`validate`, `db`, `json`) in a custom `x-go-tags` extension for total visibility.
*   **Robust Security Schemes:** Native builder methods for Kafka SASL, JWT Bearer tokens, HTTP Basic Auth, and API Keys.
*   **Zero Dependencies:** Built entirely with the standard library.

---

## 📦 Installation

```bash
go get github.com/Generalsimus/asyncapi-go
```

---

## 🚀 Quick Start

Get a fully documented Kafka event stream and a beautiful UI running in under 40 lines of code.

First, define your event payload as a standard Go struct.

```go
func main() {
    doc := asyncapi.NewAsyncAPI("User Pipeline", "1.0.0", "Kafka Event Driven Service")

    // Define a Secure Kafka Broker
    doc.AddServer("kafka-prod", "kafka.internal.net:9092", "kafka", "Prod Cluster").
        SetKafkaSASL("SCRAM-SHA-256 Authentication")

    // Define the Topic (Channel) and Payload
    signupChan := doc.AddChannel("user.events.signup", "Emitted when a new user registers")
    signupChan.AddMessage(&UserSignup{})

    // The Server publishes to this topic
    doc.AddOperation(signupChan, "send").SetKafkaConsumerGroup("user-group")

    http.HandleFunc("/docs", doc.Handler)
    http.ListenAndServe(":8080", nil)
}
```

### 2. RabbitMQ (Message Queuing / AMQP)
Perfect for task processing and worker queues.

```go
type ProcessPayment struct {
    PaymentID string  `json:"payment_id"`
    Amount    float64 `json:"amount"`
}

func main() {
    doc := asyncapi.NewAsyncAPI("Payment Gateway", "1.2.0", "RabbitMQ Task Queue")

    // Define an AMQP Server with Basic Auth
    doc.AddServer("rmq-prod", "amqp.example.com:5672", "amqp", "Main RabbitMQ Node").
        SetBasicAuth("Requires RabbitMQ username and password")

    paymentChan := doc.AddChannel("payments.process", "Queue to handle incoming payments")
    paymentChan.AddMessage(&ProcessPayment{})

    // The Worker consumes from this queue
    doc.AddOperation(paymentChan, "receive")

    http.HandleFunc("/docs", doc.Handler)
    http.ListenAndServe(":8080", nil)
}
```

### 3. WebSockets (Real-time Client/Server)
Perfect for live market data, chat applications, and mobile clients.

```go
type LiveTicker struct {
    Symbol string  `json:"symbol"`
    Price  float64 `json:"price"`
}

func main() {
    doc := asyncapi.NewAsyncAPI("Market Data API", "2.0.0", "WebSocket Streaming Server")

    // Define a Secure WebSocket Server
    doc.AddServer("ws-gateway", "wss://api.example.com/ws", "wss", "Public WebSocket Gateway").
        SetJWTAuth("Requires Bearer Token during connection handshake")

    tickerChan := doc.AddChannel("market.ticker.live", "Real-time price updates")
    tickerChan.AddMessage(&LiveTicker{})

    // The Server pushes events to the Client
    doc.AddOperation(tickerChan, "receive")

    http.HandleFunc("/docs", doc.Handler)
    http.ListenAndServe(":8080", nil)
}
```

### 4. HTTP / RPC (Request & Reply Pattern)
You can seamlessly map Request-Reply patterns natively, which is vital for B2B Webhooks or RPC calls over message brokers.

```go
// 1. Define Request and Response Channels
reqChan := doc.AddChannel("v1/orders/request", "Submit an order")
reqChan.AddMessage(&OrderRequest{})

resChan := doc.AddChannel("v1/orders/response", "Receive the order status")
resChan.AddMessage(&OrderResponse{})

// 2. Link the POST request directly to the expected reply
doc.AddHttpOperation(reqChan, "send", "POST").SetReply(resChan)
```

---

## 🔒 Supported Security Architectures

The framework provides fluent builder methods to attach standard security protocols directly to your servers, rendering beautiful, standard-compliant UI badges:

* `SetKafkaSASL(description string)`: Generates `scramSha256` Kafka properties.
* `SetJWTAuth(description string)`: Generates HTTP Bearer configuration.
* `SetAPIKeyAuth(headerName, description string)`: Generates HTTP API Key configuration.
* `SetBasicAuth(description string)`: Generates standard HTTP Username/Password configuration.

---

## 🤝 Contributing
Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change. 

If you are adding support for a new protocol binding (e.g., MQTT or SQS), please ensure you test the UI rendering via the `doc.Handler`.

## 📄 License
[MIT](https://choosealicense.com/licenses/mit/)

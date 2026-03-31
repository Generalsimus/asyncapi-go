package asyncapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// ==========================================
// AST STRUCTURES (Abstract Syntax Tree)
// ==========================================

type AsyncAPIDocument struct {
	AsyncAPI   string                `json:"asyncapi"`
	Info       AsyncAPIInfo          `json:"info"`
	Servers    map[string]*Server    `json:"servers,omitempty"`
	Channels   map[string]*Channel   `json:"channels,omitempty"`
	Operations map[string]*Operation `json:"operations,omitempty"`
	Components *Components           `json:"components,omitempty"`
}

type AsyncAPIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type Server struct {
	Host        string            `json:"host"`
	Protocol    string            `json:"protocol"`
	Description string            `json:"description,omitempty"`
	Bindings    map[string]any    `json:"bindings,omitempty"`
	Security    []*SecurityScheme `json:"security,omitempty"`
}

type SecurityScheme struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	In          string `json:"in,omitempty"`
	Name        string `json:"name,omitempty"`
}

type Components struct {
	Messages        map[string]*Message        `json:"messages,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

type Channel struct {
	Address     string              `json:"address"`
	Description string              `json:"description,omitempty"`
	Messages    map[string]*Message `json:"messages,omitempty"`
}

type Operation struct {
	Action   string          `json:"action"` // "send" or "receive"
	Channel  *ChannelRef     `json:"channel"`
	Bindings map[string]any  `json:"bindings,omitempty"`
	Reply    *OperationReply `json:"reply,omitempty"`
}

type OperationReply struct {
	Channel *ChannelRef `json:"channel"`
}

type ChannelRef struct {
	Ref string `json:"$ref"`
}

type Message struct {
	Name           string         `json:"name,omitempty"`
	Title          string         `json:"title,omitempty"`
	Payload        map[string]any `json:"payload"`
	ChannelAddress string         `json:"-"`
}

// ==========================================
// BUILDER METHODS
// ==========================================

// NewAsyncAPI initializes a new AsyncAPI v3 document.
func NewAsyncAPI(title, version, description string) *AsyncAPIDocument {
	return &AsyncAPIDocument{
		AsyncAPI: "3.0.0",
		Info: AsyncAPIInfo{
			Title:       title,
			Version:     version,
			Description: description,
		},
		Servers:    make(map[string]*Server),
		Channels:   make(map[string]*Channel),
		Operations: make(map[string]*Operation),
	}
}

// GenerateHtml returns the full HTML string containing the AsyncAPI React component.
func (doc *AsyncAPIDocument) GenerateHtml() string {
	specJSON, _ := json.Marshal(doc)
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>AsyncAPI Viewer</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;700&display=swap" rel="stylesheet"> 
    <link rel="stylesheet" href="https://unpkg.com/@asyncapi/react-component@3.1.0/styles/default.min.css">
    <style> 
        body { 
            margin: 0; 
            padding: 0; 
            background-color: #0d1117;
            font-family: 'Inter', sans-serif; 
        }  
    </style>
</head>
<body id="go-async-api"> 
    <script src="https://unpkg.com/@asyncapi/react-component@3.1.0/browser/standalone/index.js"></script>
    <script>
        AsyncApiStandalone.render(
            { schema: %s }, 
            document.getElementById('go-async-api')
        );
    </script>
</body>
</html>`, string(specJSON))
}

// Handler is a standard HTTP handler function to serve the AsyncAPI UI.
func (doc *AsyncAPIDocument) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, doc.GenerateHtml())
}

// AddServer registers a new message broker or web server to the documentation.
func (doc *AsyncAPIDocument) AddServer(serverKey, host, protocol, description string) *Server {
	server := &Server{
		Host:        host,
		Protocol:    protocol,
		Description: description,
	}
	doc.Servers[serverKey] = server
	return server
}

// SetKafkaSASL configures SCRAM-SHA-256 authentication for a Kafka server.
func (s *Server) SetKafkaSASL(description string) *Server {
	scheme := &SecurityScheme{
		Type:        "scramSha256",
		Description: description,
	}
	s.Security = append(s.Security, scheme)
	return s
}

// SetJWTAuth configures standard HTTP Bearer (JWT) authentication.
func (s *Server) SetJWTAuth(description string) *Server {
	scheme := &SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: description,
	}
	s.Security = append(s.Security, scheme)
	return s
}

// SetAPIKeyAuth configures API Key authentication injected via headers.
func (s *Server) SetAPIKeyAuth(headerName, description string) *Server {
	scheme := &SecurityScheme{
		Type:        "httpApiKey",
		In:          "header",
		Name:        headerName,
		Description: description,
	}
	s.Security = append(s.Security, scheme)
	return s
}

// SetBasicAuth configures standard HTTP Basic Authentication (Username/Password).
func (s *Server) SetBasicAuth(description string) *Server {
	scheme := &SecurityScheme{
		Type:        "http",
		Scheme:      "basic",
		Description: description,
	}
	s.Security = append(s.Security, scheme)
	return s
}

// AddChannel registers a new topic, queue, or endpoint.
func (doc *AsyncAPIDocument) AddChannel(address, description string) *Channel {
	channel := &Channel{
		Address:     address,
		Description: description,
		Messages:    make(map[string]*Message),
	}
	doc.Channels[strings.ReplaceAll(address, "/", "-")] = channel
	return channel
}

// AddMessage reflects on a Go struct to automatically generate the payload schema.
func (c *Channel) AddMessage(payload any) *Message {
	t := reflect.TypeOf(payload)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	if c.Messages == nil {
		c.Messages = make(map[string]*Message)
	}
	msg := &Message{
		Name:           name,
		Payload:        Generate(reflect.TypeOf(payload), ""),
		ChannelAddress: c.Address,
	}

	c.Messages[name] = msg
	return msg
}

// AddOperation links an action (send or receive) to a registered channel.
func (doc *AsyncAPIDocument) AddOperation(channel *Channel, action string) *Operation {
	op := &Operation{
		Action: action,
		Channel: &ChannelRef{
			Ref: fmt.Sprintf("#/channels/%s", strings.ReplaceAll(channel.Address, "/", "-")),
		},
	}

	operationKey := fmt.Sprintf("%s-%s", action, strings.ReplaceAll(channel.Address, "/", "-"))
	doc.Operations[operationKey] = op
	return op
}

// SetReply links a response channel to the current operation to establish a Request-Reply pattern.
func (op *Operation) SetReply(replyChannel *Channel) *Operation {
	op.Reply = &OperationReply{
		Channel: &ChannelRef{
			Ref: fmt.Sprintf("#/channels/%s", strings.ReplaceAll(replyChannel.Address, "/", "-")),
		},
	}
	return op
}

// AddHttpOperation links an action and defines the HTTP method binding.
func (doc *AsyncAPIDocument) AddHttpOperation(channel *Channel, action string, httpMethod string) *Operation {
	op := doc.AddOperation(channel, action)
	op.Bindings = map[string]any{
		"http": map[string]any{
			"method": strings.ToUpper(httpMethod),
		},
	}
	return op
}

// Generate recursively builds a JSON schema map from a reflected Go type.
func Generate(t reflect.Type, description string) map[string]any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	schema := map[string]any{}

	switch t.String() {
	case "string":
		schema["type"] = "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		schema["type"] = "integer"
	case "float32", "float64":
		schema["type"] = "number"
	case "bool":
		schema["type"] = "boolean"
	case "time.Time":
		schema["type"] = "string"
		schema["format"] = "date-time"
	case "uuid.UUID":
		schema["type"] = "string"
		schema["format"] = "uuid"
	default:
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			schema["type"] = "array"
			schema["items"] = Generate(t.Elem(), "")

		case reflect.Struct:
			properties := make(map[string]any)
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if !field.IsExported() {
					continue
				}
				v := strings.Split(field.Tag.Get("json"), ",")[0]

				if v == "-" {
					continue
				}
				cleanDesc := field.Tag.Get("description")

				propSchema := Generate(field.Type, cleanDesc)
				propSchema["x-go-tags"] = string(field.Tag)
				if v == "" {
					properties[field.Name] = propSchema
				} else {
					properties[v] = propSchema
				}
			}

			schema["type"] = "object"
			schema["properties"] = properties
			schema["description"] = description
		default:
			schema["type"] = "object"
		}
	}

	if description != "" {
		schema["description"] = description
	}
	return schema
}

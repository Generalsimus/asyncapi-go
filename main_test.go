package asyncapi

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewAsyncAPI(t *testing.T) {
	doc := NewAsyncAPI("Test API", "1.0.0", "A test description")

	if doc.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", doc.Info.Title)
	}
	if doc.AsyncAPI != "3.0.0" {
		t.Errorf("Expected AsyncAPI version '3.0.0', got '%s'", doc.AsyncAPI)
	}
}

func TestBuilderMethods(t *testing.T) {
	doc := NewAsyncAPI("Test API", "1.0.0", "")

	// Test Adding Server
	server := doc.AddServer("prod", "localhost:9092", "kafka", "Prod Broker")
	if server.Host != "localhost:9092" {
		t.Errorf("Expected server host 'localhost:9092', got '%s'", server.Host)
	}
	if len(doc.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(doc.Servers))
	}

	// Test Adding Channel
	channel := doc.AddChannel("test/topic", "A test topic")
	if len(doc.Channels) != 1 {
		t.Errorf("Expected 1 channel in document")
	}
	if channel.Description != "A test topic" {
		t.Errorf("Expected channel description 'A test topic'")
	}
}

func TestServerSecuritySchemes(t *testing.T) {
	doc := NewAsyncAPI("Test API", "1.0.0", "")

	kafkaServer := doc.AddServer("kafka", "localhost", "kafka", "").SetKafkaSASL("Kafka Auth")
	if len(kafkaServer.Security) == 0 || kafkaServer.Security[0].Type != "scramSha256" {
		t.Errorf("Failed to set Kafka SASL security")
	}

	jwtServer := doc.AddServer("jwt", "localhost", "http", "").SetJWTAuth("JWT Auth")
	if len(jwtServer.Security) == 0 || jwtServer.Security[0].Scheme != "bearer" {
		t.Errorf("Failed to set JWT security")
	}

	apiKeyServer := doc.AddServer("apikey", "localhost", "http", "").SetAPIKeyAuth("X-API-Key", "API Key Auth")
	if len(apiKeyServer.Security) == 0 || apiKeyServer.Security[0].Type != "httpApiKey" || apiKeyServer.Security[0].Name != "X-API-Key" {
		t.Errorf("Failed to set API Key security")
	}

	basicServer := doc.AddServer("basic", "localhost", "http", "").SetBasicAuth("Basic Auth")
	if len(basicServer.Security) == 0 || basicServer.Security[0].Scheme != "basic" {
		t.Errorf("Failed to set Basic security")
	}
}

func TestChannelAndMessages(t *testing.T) {
	doc := NewAsyncAPI("Test API", "1.0.0", "")
	channel := doc.AddChannel("test/topic", "A test topic")

	msg := channel.AddMessage(&TestStruct{}) // Tests pointer unwrapping
	if msg.Name != "TestStruct" {
		t.Errorf("Expected message name 'TestStruct' (unwrapped pointer), got '%s'", msg.Name)
	}
}

func TestOperations(t *testing.T) {
	doc := NewAsyncAPI("Test API", "1.0.0", "")
	reqChan := doc.AddChannel("req", "")
	resChan := doc.AddChannel("res", "")

	op := doc.AddOperation(reqChan, "send").SetReply(resChan)
	if op.Reply == nil || !strings.Contains(op.Reply.Channel.Ref, "res") {
		t.Errorf("Failed to set operation reply")
	}

	httpOp := doc.AddHttpOperation(reqChan, "receive", "POST")
	if httpOp.Bindings["http"].(map[string]any)["method"] != "POST" {
		t.Errorf("Failed to set HTTP operation bindings")
	}
}

func TestGeneratePrimitives(t *testing.T) {
	// Test String
	strSchema := Generate(reflect.TypeOf(""), "")
	if strSchema["type"] != "string" {
		t.Errorf("Expected type 'string', got '%v'", strSchema["type"])
	}

	// Test Integer
	intSchema := Generate(reflect.TypeOf(100), "")
	if intSchema["type"] != "integer" {
		t.Errorf("Expected type 'integer', got '%v'", intSchema["type"])
	}

	// Test Boolean
	boolSchema := Generate(reflect.TypeOf(true), "")
	if boolSchema["type"] != "boolean" {
		t.Errorf("Expected type 'boolean', got '%v'", boolSchema["type"])
	}

	// Test Float
	floatSchema := Generate(reflect.TypeOf(3.14), "")
	if floatSchema["type"] != "number" {
		t.Errorf("Expected type 'number', got '%v'", floatSchema["type"])
	}
}

type TestStruct struct {
	VisibleField string `json:"visible"`
	HiddenField  string `json:"-"`
	NoJsonTag    float64
}

func TestGenerateComplexTypes(t *testing.T) {
	// Test Slice/Array
	sliceSchema := Generate(reflect.TypeOf([]int{}), "")
	if sliceSchema["type"] != "array" {
		t.Errorf("Expected type 'array', got '%v'", sliceSchema["type"])
	}
	if sliceSchema["items"].(map[string]any)["type"] != "integer" {
		t.Errorf("Expected array items to be 'integer'")
	}

	// Test Struct and Struct Pointers
	structSchema := Generate(reflect.TypeOf(&TestStruct{}), "A test struct")
	if structSchema["type"] != "object" {
		t.Errorf("Expected type 'object' for struct, got '%v'", structSchema["type"])
	}

	props := structSchema["properties"].(map[string]any)
	if _, ok := props["visible"]; !ok {
		t.Errorf("Expected 'visible' property based on json tag")
	}
	if _, ok := props["HiddenField"]; ok {
		t.Errorf("Expected 'HiddenField' to be ignored due to json:'-'")
	}
	if _, ok := props["NoJsonTag"]; !ok {
		t.Errorf("Expected 'NoJsonTag' to be present using field name")
	}
}

func TestHTMLGenerationAndHandler(t *testing.T) {
	doc := NewAsyncAPI("Test API", "1.0.0", "A test description")
	html := doc.GenerateHtml()
	if !strings.Contains(html, "AsyncApiStandalone.render") {
		t.Errorf("HTML does not contain the AsyncAPI React component rendering logic")
	}

	// Test HTTP Handler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	doc.Handler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

// ==========================================
// ADVANCED COMPLEX TEST CASES
// ==========================================

type NestedStruct struct {
	FieldA string `json:"field_a"`
	FieldB int    `json:"field_b"`
}

type AdvancedStruct struct {
	ID        string         `json:"id,omitempty"`
	Tags      []string       `json:"tags"`
	Metadata  map[string]any `json:"metadata"`
	Nested    NestedStruct   `json:"nested"`
	NestedPtr *NestedStruct  `json:"nested_ptr"`
	Ignored   string         `json:"-"`
}

func TestGenerateAdvancedNestedTypes(t *testing.T) {
	schema := Generate(reflect.TypeOf(AdvancedStruct{}), "Advanced description")

	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("Failed to cast properties to map")
	}

	// Check ID (omitempty parsing)
	if _, ok := props["id"]; !ok {
		t.Errorf("Expected 'id' property to be parsed successfully, handling 'omitempty'")
	}

	// Check Tags (Slice of strings)
	tags := props["tags"].(map[string]any)
	if tags["type"] != "array" || tags["items"].(map[string]any)["type"] != "string" {
		t.Errorf("Failed to correctly parse string slice")
	}

	// Check Metadata (Map fallback should default to 'object')
	meta := props["metadata"].(map[string]any)
	if meta["type"] != "object" {
		t.Errorf("Expected map to fallback to 'object', got '%v'", meta["type"])
	}

	// Check Nested Struct
	nested := props["nested"].(map[string]any)
	if nested["type"] != "object" {
		t.Errorf("Expected nested struct to be 'object', got '%v'", nested["type"])
	}
	nestedProps := nested["properties"].(map[string]any)
	if _, ok := nestedProps["field_a"]; !ok {
		t.Errorf("Failed to parse nested struct internal properties")
	}

	// Check Nested Pointer Struct (Reflection unwrapping)
	nestedPtr := props["nested_ptr"].(map[string]any)
	if nestedPtr["type"] != "object" {
		t.Errorf("Expected nested pointer struct to be correctly unwrapped to 'object', got '%v'", nestedPtr["type"])
	}
}

func TestComprehensiveDocumentBuilder(t *testing.T) {
	doc := NewAsyncAPI("Massive System", "4.0.0", "Integration test document")

	// Add Multiple Servers
	doc.AddServer("kafka-1", "broker1:9092", "kafka", "Broker 1")
	doc.AddServer("kafka-2", "broker2:9092", "kafka", "Broker 2").SetKafkaSASL("SASL")
	doc.AddServer("http-api", "api.example.com", "http", "API Gateway").SetJWTAuth("JWT")
	if len(doc.Servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(doc.Servers))
	}

	// Add Multiple Channels & Link Operations
	ch1 := doc.AddChannel("system/metrics", "Metrics output")
	ch2 := doc.AddChannel("system/alerts", "Alerts output")

	op1 := doc.AddOperation(ch1, "send")
	doc.AddHttpOperation(ch2, "receive", "POST")
	op1.SetReply(ch2)

	if len(doc.Operations) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(doc.Operations))
	}
	if !strings.Contains(op1.Channel.Ref, "system-metrics") {
		t.Errorf("Incorrect channel ref in operation: %s", op1.Channel.Ref)
	}
	if op1.Reply == nil || !strings.Contains(op1.Reply.Channel.Ref, "system-alerts") {
		t.Errorf("Operation reply ref formatted incorrectly")
	}
}

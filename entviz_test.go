package entviz

import (
	"encoding/json"
	"testing"
)

func TestJsFieldComment(t *testing.T) {
	field := jsField{
		Name:    "test_field",
		Type:    "string",
		Comment: "这是一个测试字段",
	}

	data, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("Failed to marshal jsField: %v", err)
	}

	expected := `{"name":"test_field","type":"string","comment":"这是一个测试字段"}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}

	var unmarshaled jsField
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal jsField: %v", err)
	}

	if unmarshaled.Comment != field.Comment {
		t.Errorf("Expected comment %s, got %s", field.Comment, unmarshaled.Comment)
	}
}

func TestJsFieldEmptyComment(t *testing.T) {
	field := jsField{
		Name:    "test_field",
		Type:    "string",
		Comment: "",
	}

	data, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("Failed to marshal jsField: %v", err)
	}

	expected := `{"name":"test_field","type":"string","comment":""}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

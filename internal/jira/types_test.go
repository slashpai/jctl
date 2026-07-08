package jira

import "testing"

func TestErrorResponse_Summary(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrorResponse
		wantSub  string
	}{
		{
			name:    "error messages only",
			err:     ErrorResponse{ErrorMessages: []string{"not found", "bad request"}},
			wantSub: "not found",
		},
		{
			name:    "field errors only",
			err:     ErrorResponse{Errors: map[string]string{"summary": "is required"}},
			wantSub: "summary: is required",
		},
		{
			name:    "both messages and field errors",
			err:     ErrorResponse{ErrorMessages: []string{"general error"}, Errors: map[string]string{"field": "invalid"}},
			wantSub: "general error",
		},
		{
			name:    "empty",
			err:     ErrorResponse{},
			wantSub: "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Summary()
			if !contains(got, tt.wantSub) {
				t.Errorf("Summary() = %q, want substring %q", got, tt.wantSub)
			}
		})
	}
}

func TestAdfText(t *testing.T) {
	result := adfText("hello world")

	if result["type"] != "doc" {
		t.Errorf("expected type 'doc', got %v", result["type"])
	}
	if result["version"] != 1 {
		t.Errorf("expected version 1, got %v", result["version"])
	}

	content, ok := result["content"].([]map[string]any)
	if !ok || len(content) != 1 {
		t.Fatal("expected content array with 1 paragraph")
	}
	if content[0]["type"] != "paragraph" {
		t.Errorf("expected paragraph type, got %v", content[0]["type"])
	}

	innerContent, ok := content[0]["content"].([]map[string]any)
	if !ok || len(innerContent) != 1 {
		t.Fatal("expected inner content with 1 text node")
	}
	if innerContent[0]["text"] != "hello world" {
		t.Errorf("expected text 'hello world', got %v", innerContent[0]["text"])
	}
}

func TestExtractADFText(t *testing.T) {
	tests := []struct {
		name string
		input any
		want  string
	}{
		{name: "nil input", input: nil, want: ""},
		{name: "non-map input", input: "plain string", want: ""},
		{
			name: "simple paragraph",
			input: map[string]any{
				"type": "doc", "version": 1,
				"content": []any{
					map[string]any{
						"type": "paragraph",
						"content": []any{
							map[string]any{"type": "text", "text": "Hello world"},
						},
					},
				},
			},
			want: "Hello world",
		},
		{
			name: "multiple paragraphs",
			input: map[string]any{
				"type": "doc", "version": 1,
				"content": []any{
					map[string]any{
						"type": "paragraph",
						"content": []any{
							map[string]any{"type": "text", "text": "First"},
						},
					},
					map[string]any{
						"type": "paragraph",
						"content": []any{
							map[string]any{"type": "text", "text": "Second"},
						},
					},
				},
			},
			want: "First\nSecond",
		},
		{
			name: "bullet list",
			input: map[string]any{
				"type": "doc", "version": 1,
				"content": []any{
					map[string]any{
						"type": "bulletList",
						"content": []any{
							map[string]any{
								"type": "listItem",
								"content": []any{
									map[string]any{
										"type": "paragraph",
										"content": []any{
											map[string]any{"type": "text", "text": "Item one"},
										},
									},
								},
							},
							map[string]any{
								"type": "listItem",
								"content": []any{
									map[string]any{
										"type": "paragraph",
										"content": []any{
											map[string]any{"type": "text", "text": "Item two"},
										},
									},
								},
							},
						},
					},
				},
			},
			want: "  • Item one\n  • Item two",
		},
		{
			name: "empty doc",
			input: map[string]any{
				"type": "doc", "version": 1,
				"content": []any{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractADFText(tt.input)
			if got != tt.want {
				t.Errorf("ExtractADFText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	t.Run("valid JSON error", func(t *testing.T) {
		body := []byte(`{"errorMessages":["Something went wrong"],"errors":{}}`)
		err := parseError(body, 400)
		if err == nil {
			t.Fatal("expected error")
		}
		if !contains(err.Error(), "Something went wrong") {
			t.Errorf("expected error message in output, got: %s", err.Error())
		}
		if !contains(err.Error(), "400") {
			t.Errorf("expected status code in output, got: %s", err.Error())
		}
	})

	t.Run("invalid JSON falls back to raw body", func(t *testing.T) {
		body := []byte("Internal Server Error")
		err := parseError(body, 500)
		if err == nil {
			t.Fatal("expected error")
		}
		if !contains(err.Error(), "Internal Server Error") {
			t.Errorf("expected raw body in output, got: %s", err.Error())
		}
	})
}

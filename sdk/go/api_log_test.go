package pan302plugin

import "testing"

func TestLogFieldsStringifiesStruct(t *testing.T) {
	type detail struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
	}
	type payload struct {
		Count  int      `json:"count"`
		OK     bool     `json:"ok"`
		Detail detail   `json:"detail"`
		Tags   []string `json:"tags"`
	}

	fields := logFields(payload{
		Count: 3,
		OK:    true,
		Detail: detail{
			Path: "/movies/a.mkv",
			Size: 1024,
		},
		Tags: []string{"movie", "new"},
	})

	if fields["count"] != "3" {
		t.Fatalf("expected count to be stringified, got %q", fields["count"])
	}
	if fields["ok"] != "true" {
		t.Fatalf("expected ok to be stringified, got %q", fields["ok"])
	}
	if fields["detail"] != `{"path":"/movies/a.mkv","size":1024}` {
		t.Fatalf("expected detail to be JSON, got %q", fields["detail"])
	}
	if fields["tags"] != `["movie","new"]` {
		t.Fatalf("expected tags to be JSON, got %q", fields["tags"])
	}
}

func TestLogFieldsStringifiesMapValues(t *testing.T) {
	fields := logFields(map[string]any{
		"count": 3,
		"ok":    true,
		"nested": map[string]any{
			"path": "/movies/a.mkv",
		},
	})

	if fields["count"] != "3" {
		t.Fatalf("expected count to be stringified, got %q", fields["count"])
	}
	if fields["ok"] != "true" {
		t.Fatalf("expected ok to be stringified, got %q", fields["ok"])
	}
	if fields["nested"] != `{"path":"/movies/a.mkv"}` {
		t.Fatalf("expected nested to be JSON, got %q", fields["nested"])
	}
}

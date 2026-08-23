package validation

import "testing"

func TestValidation(t *testing.T) {
	if ID("bad id") == nil || ID("plot_01") != nil {
		t.Fatal("id")
	}
	if Required("name", "") == nil || Required("name", "ok") != nil {
		t.Fatal("required")
	}
	if Length("x", "abc", 1, 2) == nil || Length("x", "abc", 1, 3) != nil {
		t.Fatal("length")
	}
	if URL("not-url") == nil || URL("https://example.com") != nil {
		t.Fatal("url")
	}
	if Enum("state", "bad", "open", "closed") == nil {
		t.Fatal("enum")
	}
}

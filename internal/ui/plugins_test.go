package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestFileHash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tailwind.css")
	hashOf := func(content string) string {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		hash, err := fileHash(path)
		if err != nil {
			t.Fatal(err)
		}
		return hash
	}

	first := hashOf(".grid{display:grid}")
	if again := hashOf(".grid{display:grid}"); again != first {
		t.Errorf("fileHash changed for unchanged content: %q != %q", again, first)
	}
	if changed := hashOf(".grid{display:grid}.grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}"); changed == first {
		t.Errorf("fileHash(%q) did not change when the content changed", first)
	}
	if _, err := fileHash(filepath.Join(t.TempDir(), "missing.css")); err == nil {
		t.Error("fileHash of a missing file: got nil error")
	}
}

func TestRenderHtml(t *testing.T) {
	data := htmlData{
		TailwindHash: "0123abcd",
		OutputFiles: []api.OutputFile{
			{Path: filepath.Join(distDir, "index-AAAA.js")},
			{Path: filepath.Join(distDir, "index-AAAA.js.map")},
			{Path: filepath.Join(distDir, "index-BBBB.css")},
		},
	}
	html, err := renderHtml(public("index.tmpl.html"), data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<link rel="stylesheet" href="/static/tailwind.css?v=0123abcd">`,
		`<link rel="stylesheet" href="/static/index-BBBB.css">`,
		`<script type="module" src="/static/index-AAAA.js" defer></script>`,
	} {
		if !strings.Contains(string(html), want) {
			t.Errorf("index.html lacks %s:\n%s", want, html)
		}
	}
	if strings.Contains(string(html), "index-AAAA.js.map") {
		t.Errorf("index.html links the source map:\n%s", html)
	}
}

func TestWriteHtml(t *testing.T) {
	dir := t.TempDir()
	tailwindPath := filepath.Join(dir, "tailwind.css")
	htmlPath := filepath.Join(dir, "index.html")
	write := func() string {
		t.Helper()
		if err := writeHtml(public("index.tmpl.html"), htmlPath, tailwindPath, nil); err != nil {
			t.Fatal(err)
		}
		html, err := os.ReadFile(htmlPath)
		if err != nil {
			t.Fatal(err)
		}
		return string(html)
	}

	if html, want := write(), `href="/static/tailwind.css"`; !strings.Contains(html, want) {
		t.Errorf("index.html lacks %s when tailwind.css is missing:\n%s", want, html)
	}

	if err := os.WriteFile(tailwindPath, []byte(".grid{display:grid}"), 0o600); err != nil {
		t.Fatal(err)
	}
	hash, err := fileHash(tailwindPath)
	if err != nil {
		t.Fatal(err)
	}
	if html, want := write(), `href="/static/tailwind.css?v=`+hash+`"`; !strings.Contains(html, want) {
		t.Errorf("index.html lacks %s:\n%s", want, html)
	}
}

package service

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Links a mermaid source into the live editor at https://mermaid.live.
//
// The editor keeps its whole state in the URL fragment, so the diagram travels
// in the link and is never uploaded anywhere — which is what makes this safe to
// put in an exported document that may leave the machine.
//
// The fragment is `pako:<payload>`: the editor's own State as JSON, zlib
// deflated, then url-safe base64. The frontend builds the same thing in the
// browser (see frontend/composables/useMermaidLive.ts); this is the copy the
// export needs, because an export is produced server-side.
const mermaidLiveBase = "https://mermaid.live/view#pako:"

// mermaidLiveState mirrors the live editor's State (mermaid-live-editor,
// src/lib/types.d.ts). Note that `mermaid` is a JSON *string* there, not an
// object, and that the config matches the editor's default so the diagram is
// drawn the way a fresh editor would draw it rather than in this app's theme.
type mermaidLiveState struct {
	Code          string `json:"code"`
	Mermaid       string `json:"mermaid"`
	UpdateDiagram bool   `json:"updateDiagram"`
	Rough         bool   `json:"rough"`
	PanZoom       bool   `json:"panZoom"`
}

// MermaidLiveURL returns a link that opens code in the live editor, or an empty
// string if it cannot be built — a caller writing an export should simply omit
// the line rather than emit a broken link.
func MermaidLiveURL(code string) string {
	if code == "" {
		return ""
	}
	state, err := json.Marshal(mermaidLiveState{
		Code:          code,
		Mermaid:       "{\n  \"theme\": \"default\"\n}",
		UpdateDiagram: true,
		PanZoom:       true,
	})
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	// zlib, not raw deflate: pako.deflate produces the zlib wrapper, and the
	// editor's inflate expects it.
	zw, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return ""
	}
	if _, err := zw.Write(state); err != nil {
		return ""
	}
	if err := zw.Close(); err != nil {
		return ""
	}
	return mermaidLiveBase + base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

// mermaidLiveLine is the markdown an export appends under a diagram.
func mermaidLiveLine(code string) string {
	url := MermaidLiveURL(code)
	if url == "" {
		return ""
	}
	return fmt.Sprintf("[Open this diagram in mermaid.live](%s)\n", url)
}

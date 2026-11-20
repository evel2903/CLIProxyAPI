package helps

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/MoeForge/moego/ojson"
)

func claudeTestImageBlock(t *testing.T, w, h int) string {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(`{"type":"image","source":{"type":"base64","media_type":"image/png","data":%q}}`, base64.StdEncoding.EncodeToString(buf.Bytes()))
}

func claudeTestImageMessages(blocks []string) []byte {
	return []byte(`{"messages":[{"role":"user","content":[` + strings.Join(blocks, ",") + `]}]}`)
}

func TestCapClaudeImageDimensionsDoc(t *testing.T) {
	blocks := []string{claudeTestImageBlock(t, 2600, 40)}
	for range claudeManyImageThreshold - 1 {
		blocks = append(blocks, claudeTestImageBlock(t, 10, 10))
	}

	// At the threshold the request is not a many-image request: leave it alone.
	if CapClaudeImageDimensionsDoc(ojson.NewDoc(claudeTestImageMessages(blocks))) {
		t.Fatal("images must not be resized at the many-image threshold")
	}

	// One more image, nested in a tool_result, crosses it.
	toolResult := `{"type":"tool_result","tool_use_id":"toolu_1","content":[` + claudeTestImageBlock(t, 30, 3000) + `]}`
	doc := ojson.NewDoc(claudeTestImageMessages(append(blocks, toolResult)))
	if !CapClaudeImageDimensionsDoc(doc) {
		t.Fatal("expected oversized images to be resized")
	}
	for path, want := range map[string][2]int{
		"messages.0.content.0":            {2000, 30},
		"messages.0.content.20.content.0": {20, 2000},
		"messages.0.content.1":            {10, 10},
	} {
		raw, err := base64.StdEncoding.DecodeString(doc.Get(path + ".source.data").String())
		if err != nil {
			t.Fatalf("%s: bad base64: %v", path, err)
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
		if err != nil || cfg.Width != want[0] || cfg.Height != want[1] {
			t.Fatalf("%s: got %dx%d (err %v), want %dx%d", path, cfg.Width, cfg.Height, err, want[0], want[1])
		}
	}
}

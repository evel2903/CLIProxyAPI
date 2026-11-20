package helps

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"

	"github.com/MoeForge/moego/ojson"
	log "github.com/sirupsen/logrus"
)

// Anthropic caps every image side once a request carries more than
// claudeManyImageThreshold images (documented as 2000px; the API currently
// reports 3000px). Clients such as Codex replay each viewed screenshot at full
// resolution, so long sessions cross both limits and every later turn fails.
const (
	claudeManyImageThreshold = 20
	claudeManyImageMaxSide   = 2000
)

// CapClaudeImageDimensionsDoc downscales base64 images whose longest side
// exceeds claudeManyImageMaxSide when doc carries more than
// claudeManyImageThreshold images, including images nested in tool_result
// blocks. Images that cannot be decoded (e.g. WebP) are left unchanged.
// ponytail: re-decodes oversized images on every replay; cache resized data by hash if latency shows up.
func CapClaudeImageDimensionsDoc(doc *ojson.Doc) bool {
	if doc == nil || doc.Len() == 0 {
		return false
	}
	var paths []string
	for i, message := range doc.Get("messages").Array() {
		for j, block := range message.Get("content").Array() {
			switch block.Get("type").String() {
			case "image":
				paths = append(paths, fmt.Sprintf("messages.%d.content.%d", i, j))
			case "tool_result":
				for k, inner := range block.Get("content").Array() {
					if inner.Get("type").String() == "image" {
						paths = append(paths, fmt.Sprintf("messages.%d.content.%d.content.%d", i, j, k))
					}
				}
			}
		}
	}
	if len(paths) <= claudeManyImageThreshold {
		return false
	}
	resized := 0
	for _, path := range paths {
		source := doc.Get(path + ".source")
		if source.Get("type").String() != "base64" {
			continue
		}
		data, mediaType, ok := downscaleBase64Image(source.Get("data").String(), claudeManyImageMaxSide)
		if !ok {
			continue
		}
		_ = doc.Set(path+".source.data", data)
		_ = doc.Set(path+".source.media_type", mediaType)
		resized++
	}
	if resized > 0 {
		log.Debugf("claude images: downscaled %d of %d image(s) to %dpx for a many-image request", resized, len(paths), claudeManyImageMaxSide)
	}
	return resized > 0
}

// downscaleBase64Image returns the image re-encoded with its longest side at
// maxSide, or ok=false when it already fits or cannot be decoded. JPEG stays
// JPEG; everything else is re-encoded as PNG.
func downscaleBase64Image(encoded string, maxSide int) (data, mediaType string, ok bool) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || max(cfg.Width, cfg.Height) <= maxSide {
		return "", "", false
	}
	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", "", false
	}
	small := downscaleImage(img, maxSide)
	var buf bytes.Buffer
	mediaType = "image/png"
	if format == "jpeg" {
		mediaType = "image/jpeg"
		err = jpeg.Encode(&buf, small, &jpeg.Options{Quality: 90})
	} else {
		err = png.Encode(&buf, small)
	}
	if err != nil {
		return "", "", false
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), mediaType, true
}

// downscaleImage shrinks img so its longest side is maxSide, averaging each
// source box into one destination pixel (premultiplied RGBA, so alpha is
// averaged correctly).
func downscaleImage(img image.Image, maxSide int) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	src := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(src, src.Bounds(), img, b.Min, draw.Src)
	longest := max(w, h)
	nw, nh := max(w*maxSide/longest, 1), max(h*maxSide/longest, 1)
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		y0, y1 := y*h/nh, max((y+1)*h/nh, y*h/nh+1)
		for x := 0; x < nw; x++ {
			x0, x1 := x*w/nw, max((x+1)*w/nw, x*w/nw+1)
			var sum [4]int
			for sy := y0; sy < y1; sy++ {
				row := src.Pix[sy*src.Stride+x0*4 : sy*src.Stride+x1*4]
				for i, v := range row {
					sum[i%4] += int(v)
				}
			}
			n := (y1 - y0) * (x1 - x0)
			px := dst.Pix[y*dst.Stride+x*4:]
			for c := range 4 {
				px[c] = uint8(sum[c] / n)
			}
		}
	}
	return dst
}

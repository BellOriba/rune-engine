package ascii

import   (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}

	for y := range 4 {
		for x := range 4 {
			if y < 2 {
				img.Set(x, y, white)
			} else {
				img.Set(x, y, black)
			}
		}
	}

	tests := []struct {
		name string
		targetWidth int
		expectedRows int
		expectedChar string
	}{
		{
			name: "Standard 4-wide",
			targetWidth: 4,
			expectedRows: 2,
			expectedChar: "@",
		},
		{
			name: "Downsampled 2-wide",
			targetWidth: 2,
			expectedRows: 1,
			expectedChar: "@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConverter(Options{TargetWidth: tt.targetWidth, Mode: "plain"})
			result := conv.Convert(img)

			result = strings.TrimSuffix(result, "\n")
			rows := strings.Split(result, "\n")
			if len(rows) != tt.expectedRows {
				t.Errorf("expected %d rows, got %d", tt.expectedRows, len(rows))
			}

			if !strings.Contains(rows[0], tt.expectedChar) {
				t.Errorf("expected row to contain %q, got %q", tt.expectedChar, rows[0])
			}
		})
	}
}

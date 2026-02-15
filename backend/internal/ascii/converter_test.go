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

func TestConvertANSI(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	conv := NewConverter(Options{TargetWidth: 1, Mode: "ansi"})
	result := conv.Convert(img)

	expectedPrefix := "\x1b[38;2;255;0;0m"
	if !strings.HasPrefix(result, expectedPrefix) {
		t.Errorf("esperado prefixo ANSI %q, obtido %q", expectedPrefix, result)
	}
}

func TestDecode(t *testing.T) {
	invalidData := strings.NewReader("not-an-image")
	_, _, err := Decode(invalidData)
	if err == nil {
		t.Error("esperado erro ao decodificar dados inválidos, obtido nil")
	}
}

func TestTargetHeightCalculation(t *testing.T) {
	tests := []struct {
		name string
		w, h, targetW int
		expectedH int
	}{
		{"Imagem Quadrada", 100, 100, 100, 55},
		{"Imagem Horizontal (Wide)", 200, 100, 100, 27},
		{"Garantia de Altura Mínima", 1000, 10, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tt.w, tt.h))
			conv := NewConverter(Options{TargetWidth: tt.targetW})

			result := conv.Convert(img)
			result = strings.TrimSuffix(result, "\n")
			rows := strings.Split(result, "\n")

			if len(rows) != tt.expectedH {
				t.Errorf("%s: esperado altura %d, obtido %d", tt.name, tt.expectedH, len(rows))
			}
		})
	}
}

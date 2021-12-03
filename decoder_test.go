package qoi

import (
	"image"
	"image/png"
	"os"
	"reflect"
	"testing"
)

func TestDecoder(t *testing.T) {
	f, err := os.Open("testdata/dice.qoi")
	if err != nil {
		t.Fatal(err)
	}
	img, err := Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	pf, err := os.Open("testdata/dice.png")
	if err != nil {
		t.Fatal(err)
	}
	pi, err := png.Decode(pf)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := len(pi.(*image.NRGBA).Pix), len(img.(*image.NRGBA).Pix); want != got {
		t.Fatalf("len mismatch. want: %d, got: %d", want, got)
	}

	if !reflect.DeepEqual(img.Bounds(), pi.Bounds()) {
		t.Errorf("%v != %v", img.Bounds(), pi.Bounds())
	}
	bounds := pi.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if want, got := pi.At(x, y), img.At(x, y); !reflect.DeepEqual(want, got) {
				t.Fatalf("differing pixel at (%d,%d). want: %#v, got: %#v", x, y, want, got)
			}
		}
	}
}

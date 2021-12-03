package qoi

import (
	"bufio"
	"encoding/binary"
	"image"
	"image/color"
	"io"
)

func Encode(w io.Writer, img image.Image) error {
	b := img.Bounds()
	desc := desc{
		Magic:  magicBytes,
		Width:  uint32(b.Max.X - b.Min.X),
		Height: uint32(b.Max.Y - b.Min.Y),
	}
	if err := binary.Write(w, binary.BigEndian, desc); err != nil {
		return err
	}

	enc := NewEncoder(w)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if err := enc.Encode(img.At(x, y)); err != nil {
				return err
			}
		}
	}
	if err := enc.Finish(); err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, uint32(0)) // padding
}

type Encoder struct {
	w *bufio.Writer

	prev color.NRGBA
	run  int
	seen [64]color.NRGBA

	err error
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:    bufio.NewWriter(w),
		prev: transparent,
	}
}

func (e *Encoder) Encode(c color.Color) error {
	px := color.NRGBA64Model.Convert(c).(color.NRGBA)

	// we use this to avoid the overhead of defer()
	// since Encode is called in a tight loop
	exit := func() error {
		e.prev = px
		return e.err
	}

	if px == e.prev {
		e.run++
	}

	if e.run == 0x2020 || px != e.prev {
		e.writeRun()
	}

	if px == e.prev {
		return e.err
	}

	pos := hash(px)

	// check if we've seen this color before
	if e.seen[pos] == px {
		e.writeByte(QOI_INDEX | pos)
		return exit()
	}

	e.seen[pos] = px

	dR := int(px.R) - int(e.prev.R)
	dG := int(px.G) - int(e.prev.G)
	dB := int(px.B) - int(e.prev.B)
	dA := int(px.A) - int(e.prev.A)

	// see if we can write out a delta
	if dR > -17 && dR < 16 &&
		dG > -17 && dG < 16 &&
		dB > -17 && dB < 16 &&
		dA > -17 && dA < 16 {
		switch {
		case dA == 0 &&
			dR > -3 && dR < 2 &&
			dG > -3 && dG < 2 &&
			dB > -3 && dB < 2:
			e.writeByte(QOI_DIFF_8 | byte(((dR+2)<<4)|(dG+2)<<2|(dB+2)))
		case dA == 0 &&
			dR > -17 && dR < 16 &&
			dG > -9 && dG < 8 &&
			dB > -9 && dB < 8:
			e.writeByte(QOI_DIFF_16 | byte(dR+16))
			e.writeByte(byte(((dG + 8) << 4) | (dB + 8)))
		default:
			e.writeByte(QOI_DIFF_24 | byte((dR+16)>>1))
			e.writeByte(byte(((dR + 16) << 7) | ((dG + 16) << 2) | ((dB + 16) >> 3)))
			e.writeByte(byte(((dB + 16) << 5) | (dA + 16)))
		}
		return exit()
	}

	// color is too different; write need to write out the whole value
	mask := QOI_COLOR
	if dR != 0 {
		mask |= 1 << 3
	}
	if dG != 0 {
		mask |= 1 << 2
	}
	if dB != 0 {
		mask |= 1 << 1
	}
	if dA != 0 {
		mask |= 1 << 0
	}
	e.writeByte(mask)
	if dR != 0 {
		e.writeByte(px.R)
	}
	if dG != 0 {
		e.writeByte(px.G)
	}
	if dB != 0 {
		e.writeByte(px.B)
	}
	if dA != 0 {
		e.writeByte(px.A)
	}

	return exit()
}

func (e *Encoder) Finish() error {
	e.writeRun()
	if e.err != nil {
		return e.err
	}
	return e.w.Flush()
}

func (e *Encoder) writeRun() {
	if e.run <= 0 {
		return
	}
	if e.run < 33 {
		e.writeByte(QOI_RUN_8 | byte(e.run-1))
	} else {
		e.run -= 33
		e.writeByte(QOI_RUN_16 | byte(e.run>>8))
		e.writeByte(byte(e.run & 0xff))
	}
	e.run = 0
}

func (e *Encoder) writeByte(b byte) {
	if e.err != nil {
		return
	}
	e.err = e.w.WriteByte(b)
}

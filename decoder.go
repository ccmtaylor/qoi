package qoi

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
)

const (
	QOI_INDEX   = 0x00 // 00xxxxxx
	QOI_RUN_8   = 0x40 // 010xxxxx
	QOI_RUN_16  = 0x60 // 011xxxxx
	QOI_DIFF_8  = 0x80 // 10xxxxxx
	QOI_DIFF_16 = 0xc0 // 110xxxxx
	QOI_DIFF_24 = 0xe0 // 1110xxxx
	QOI_COLOR   = 0xf0 // 1111xxxx

	QOI_MASK_2 = 0xc0 // 11000000
	QOI_MASK_3 = 0xe0 // 11100000
	QOI_MASK_4 = 0xf0 // 11110000
)

var (
	ErrBadMagic = errors.New("bad magic value")

	transparent = color.NRGBA{0, 0, 0, 255}
)

func init() {
	image.RegisterFormat("qoi", "qoif", Decode, DecodeConfig)
}

func Decode(r io.Reader) (image.Image, error) {
	desc, err := readDesc(r)
	if err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, int(desc.Width), int(desc.Height)))
	p := NewDecoder(r)
	for p.Next() {
		c := p.Current()
		img.Pix = append(img.Pix, c.R, c.G, c.B, c.A)
	}
	if p.Err() == io.EOF {
		return img, nil
	}
	return img, p.Err()
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	desc, err := readDesc(r)
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{
		ColorModel: color.NRGBAModel,
		Width:      int(desc.Width),
		Height:     int(desc.Height),
	}, nil
}

type desc struct {
	Magic                [4]byte
	Width, Height        uint32
	Channels, Colorspace uint8
}

func readDesc(r io.Reader) (desc desc, err error) {
	err = binary.Read(r, binary.BigEndian, &desc)
	if err != nil {
		return
	}
	if string(desc.Magic[:]) != "qoif" {
		return desc, ErrBadMagic
	}
	// if desc.Width == 0 || desc.Height == 0 {}
	if desc.Channels < 3 || desc.Channels > 4 {
		return desc, fmt.Errorf("bad channels: %d", desc.Channels)
	}
	return
}

type Decoder struct {
	r   *bufio.Reader
	cur color.NRGBA

	seen [64]color.NRGBA
	run  int

	err error
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:   bufio.NewReader(r),
		cur: transparent,
	}
}

func (p *Decoder) read8() (b byte, ok bool) {
	b, p.err = p.r.ReadByte()
	return b, p.err == nil
}

func (p *Decoder) Next() bool {
	// we're in a run of consecutive identical pixels; no need to read more data
	if p.run > 0 {
		p.run--
		return true
	}

	if p.err != nil {
		return false
	}

	b1, ok := p.read8()
	if !ok {
		return false
	}

	switch {
	case (b1 & QOI_MASK_2) == QOI_INDEX:
		p.cur = p.seen[b1^QOI_INDEX]

	case (b1 & QOI_MASK_3) == QOI_RUN_8:
		p.run = int(b1 & 0x1f)

	case (b1 & QOI_MASK_3) == QOI_RUN_16:
		b2, ok := p.read8()
		if !ok {
			return false
		}
		p.run = (((int(b1) & 0x1f) << 8) | int(b2)) + 32

	case (b1 & QOI_MASK_2) == QOI_DIFF_8:
		p.cur.R += ((b1 >> 4) & 0x03) - 2
		p.cur.G += ((b1 >> 2) & 0x03) - 2
		p.cur.B += ((b1 >> 0) & 0x03) - 2

	case (b1 & QOI_MASK_3) == QOI_DIFF_16:
		b2, ok := p.read8()
		if !ok {
			return false
		}
		p.cur.R += (b1 & 0x1f) - 16
		p.cur.G += (b2 >> 4) - 8
		p.cur.B += (b2 & 0x0f) - 8

	case (b1 & QOI_MASK_4) == QOI_DIFF_24:
		b2, ok := p.read8()
		if !ok {
			return false
		}
		b3, ok := p.read8()
		if !ok {
			return false
		}
		p.cur.R += (((b1 & 0x0f) << 1) | (b2 >> 7)) - 16
		p.cur.G += ((b2 & 0x7c) >> 2) - 16
		p.cur.B += (((b2 & 0x03) << 3) | ((b3 & 0xe0) >> 5)) - 16
		p.cur.A += (b3 & 0x1f) - 16

	case (b1 & QOI_MASK_4) == QOI_COLOR:
		if b1&8 != 0 {
			if p.cur.R, ok = p.read8(); !ok {
				return false
			}
		}
		if b1&4 != 0 {
			if p.cur.G, ok = p.read8(); !ok {
				return false
			}
		}
		if b1&2 != 0 {
			if p.cur.B, ok = p.read8(); !ok {
				return false
			}
		}
		if b1&1 != 0 {
			if p.cur.A, ok = p.read8(); !ok {
				return false
			}
		}
	}

	p.seen[(p.cur.R^p.cur.G^p.cur.B^p.cur.A)%64] = p.cur
	return true
}

func (p *Decoder) Current() color.NRGBA {
	return p.cur
}

func (p *Decoder) Err() error {
	return p.err
}

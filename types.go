package qoi

import (
	"image/color"
)

const (
	QOI_INDEX   byte = 0x00 // 00xxxxxx
	QOI_RUN_8   byte = 0x40 // 010xxxxx
	QOI_RUN_16  byte = 0x60 // 011xxxxx
	QOI_DIFF_8  byte = 0x80 // 10xxxxxx
	QOI_DIFF_16 byte = 0xc0 // 110xxxxx
	QOI_DIFF_24 byte = 0xe0 // 1110xxxx
	QOI_COLOR   byte = 0xf0 // 1111xxxx

	QOI_MASK_2 byte = 0xc0 // 11000000
	QOI_MASK_3 byte = 0xe0 // 11100000
	QOI_MASK_4 byte = 0xf0 // 11110000
)

type desc struct {
	Magic                [4]byte
	Width, Height        uint32
	Channels, Colorspace uint8
}

var (
	Magic = string(magicBytes[:])

	magicBytes  = [4]byte{'q', 'o', 'i', 'f'}
	transparent = color.NRGBA{0, 0, 0, 255}
)

func hash(px color.NRGBA) uint8 {
	return (px.R ^ px.G ^ px.B ^ px.A) % 64
}

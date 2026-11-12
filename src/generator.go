package main

import (
	"image"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

func GenerateFakeText(targetBytes int) []byte {
	if targetBytes <= 0 {
		return nil
	}

	var sb strings.Builder
	sb.Grow(targetBytes + 512)

	for sb.Len() < targetBytes {
		paragraph := gofakeit.LoremIpsumParagraph(1, 4, 10, "\n\n")
		sb.WriteString(paragraph)
		sb.WriteString("\n\n")
	}

	fullStr := sb.String()
	return []byte(fullStr[:targetBytes-4])
}

func CalculatePayloadSize(bounds image.Rectangle, percent float64, saltLen, ivLen int) int {
	width := bounds.Dx()
	height := bounds.Dy()
	totalBits := width * height * 3
	totalBytes := totalBits / 8
	targetBytes := int(float64(totalBytes)*percent/100) - saltLen - ivLen
	return targetBytes
}

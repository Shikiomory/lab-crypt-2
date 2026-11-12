package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
)

func getPositions(password string, capacityBits int) []int {
	hash := sha256.Sum256([]byte(password))
	seed := int64(binary.BigEndian.Uint64(hash[:8]))

	rng := rand.New(rand.NewSource(seed))
	return rng.Perm(capacityBits)
}

func CipherRandLsb(cover []byte, byteText []byte, password string) ([]byte, error) {
	const mask byte = 0xFE

	stego := make([]byte, len(cover))
	copy(stego, cover)

	lenHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(lenHeader, uint32(len(byteText)))
	byteTextH := append(lenHeader, byteText...)

	totalBits := len(byteTextH) * 8

	capacityBits := len(stego) / 4 * 3
	if totalBits > capacityBits {
		return nil, fmt.Errorf("Сообщение превышает размер изображения: %d/%d бит", totalBits, capacityBits)
	}

	positions := getPositions(password, capacityBits)
	for i := 0; i < totalBits; i++ {
		bytePos := i / 8
		currentBitIdx := 7 - (i % 8)
		secretBit := (byteTextH[bytePos] >> currentBitIdx) & 1

		pos := positions[i]
		currentChannelIdx := pos % 3
		pixelIdx := pos/3*4 + currentChannelIdx

		stego[pixelIdx] = (stego[pixelIdx] & mask) | secretBit
	}

	return stego, nil
}

func DecryptRandLsb(imgPixels []byte, password string) ([]byte, error) {
	imgCapacity := len(imgPixels) / 4 * 3
	if imgCapacity < 32 {
		return nil, fmt.Errorf("В изображении нет места даже для длины сообщения")
	}

	headerBytes := make([]byte, 4)
	headerLen := 4
	positions := getPositions(password, imgCapacity)

	for i := 0; i < headerLen*8; i++ {
		pos := positions[i]
		currentChannelIdx := pos % 3
		pixelIdx := pos/3*4 + currentChannelIdx

		currentBit := imgPixels[pixelIdx] & 1

		byteHeaderPos := i / 8
		currentHeaderBitIdx := 7 - (i % 8)
		headerBytes[byteHeaderPos] |= currentBit << byte(currentHeaderBitIdx)
	}

	msgLen := binary.BigEndian.Uint32(headerBytes)
	msg := make([]byte, msgLen)
	totalBits := int(msgLen) * 8

	if totalBits+headerLen*8 > imgCapacity {
		return nil, fmt.Errorf("Длина сообщения превышает размер изображения")
	}

	for i := 0; i < totalBits; i++ {
		globalIdx := 32 + i
		pos := positions[globalIdx]
		currentChannelIdx := pos % 3
		pixelIdx := pos/3*4 + currentChannelIdx

		currentBit := imgPixels[pixelIdx] & 1

		bytePos := i / 8
		currentHeaderBitIdx := 7 - (i % 8)
		msg[bytePos] |= currentBit << byte(currentHeaderBitIdx)
	}
	return msg, nil
}

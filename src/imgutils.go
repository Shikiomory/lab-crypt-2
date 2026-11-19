package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
)

func readImage(filePath string) ([]byte, image.Rectangle, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, image.Rectangle{}, fmt.Errorf("Ошибка при открытии файла: %w", err)
	}

	defer file.Close()

	srcImg, err := png.Decode(file)
	if err != nil {
		return nil, image.Rectangle{}, fmt.Errorf("Ошибка при распаковке изображения: %w", err)
	}

	bounds := srcImg.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, srcImg, bounds.Min, draw.Src)

	return nrgba.Pix, bounds, nil
}

func saveImage(pixels []byte, bounds image.Rectangle, filePath string) error {
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Ошибка при сохраении файла: %w", err)
	}

	defer outFile.Close()

	img := &image.NRGBA{
		Pix:    pixels,
		Stride: bounds.Dx() * 4,
		Rect:   bounds,
	}

	err = png.Encode(outFile, img)
	if err != nil {
		return fmt.Errorf("Ошибка при кодировании файла: %w", err)
	}

	return nil
}

func SaveDiffImage(cover, stego []byte, bounds image.Rectangle, filePath string) error {

	diffImg := make([]byte, len(cover))
	copy(diffImg, cover)

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Ошибка при сохраении файла: %w", err)
	}

	defer outFile.Close()

	for i := 0; i < len(cover); i += 4 {
		// isChanged := cover[i] != stego[i] || cover[i+1] != stego[i+1] || cover[i+2] != stego[i+2]
		diffR := math.Abs(float64(cover[i]) - float64(stego[i]))
		diffG := math.Abs(float64(cover[i+1]) - float64(stego[i+1]))
		diffB := math.Abs(float64(cover[i+2]) - float64(stego[i+2]))

		diffImg[i] = byte(diffR * 255)
		diffImg[i+1] = byte(diffG * 255)
		diffImg[i+2] = byte(diffB * 255)
		// if isChanged {
		// 	diffImg[i] = byte(math.Abs(float64(cover[i] - stego[i])))
		// 	diffImg[i+1] = byte(math.Abs(float64(cover[i+1] - stego[i+1])))
		// 	diffImg[i+2] = byte(math.Abs(float64(cover[i+2] - stego[i+2])))
		// }
		// else {
		// 	diffImg[i] = cover[i] / 2
		// 	diffImg[i+1] = cover[i+1] / 2
		// 	diffImg[i+2] = cover[i+2] / 2
		// }
	}

	img := &image.NRGBA{
		Pix:    diffImg,
		Stride: bounds.Dx() * 4,
		Rect:   bounds,
	}
	err = png.Encode(outFile, img)
	if err != nil {
		return fmt.Errorf("Ошибка при кодировании файла: %w", err)
	}

	return nil
}

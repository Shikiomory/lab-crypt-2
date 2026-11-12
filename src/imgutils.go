package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
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
		isChanged := cover[i] != stego[i] || cover[i+1] != stego[i+1] || cover[i+2] != stego[i+2]
		if isChanged {
			diffImg[i] = 255
			diffImg[i+1] = 255
			diffImg[i+2] = 255
		} else {
			diffImg[i] = cover[i] / 2
			diffImg[i+1] = cover[i+1] / 2
			diffImg[i+2] = cover[i+2] / 2
		}
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

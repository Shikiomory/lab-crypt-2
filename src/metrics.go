package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"math"
	"os"
	"strconv"
)

type MetricRow struct {
	Method       string
	Payload      int
	PSNR         float64
	SSIM         float64
	ChangedCount int
	ChangedRatio float64
	ProbLSB0     float64
	ProbLSB1     float64
	ChiSquare    float64
	Pvalue       float64
}

func CalculatePSNR(cover, stego []byte) (psnr float64) {
	totalSum := 0.0
	totalComponents := 0

	for i := 0; i < len(cover); i++ {
		if i%4 == 3 {
			continue
		}

		diff := float64(cover[i]) - float64(stego[i])
		totalSum += diff * diff
		totalComponents++
	}

	mse := totalSum / float64(totalComponents)
	return 10.0 * math.Log10((255.0*255.0)/mse)
}

func CalculateSSIM(cover, stego []byte, bounds image.Rectangle) (ssim float64) {
	width := bounds.Dx()
	height := bounds.Dy()

	const blocksize = 8
	const coef = 1.0 / (blocksize * blocksize)
	const c1 = 0.01 * 0.01 * 255.0 * 255.0
	const c2 = 0.03 * 0.03 * 255.0 * 255.0
	totalSSIM := 0.0
	totalBlocks := 0
	getPixelVal := func(x, y int, pixels []byte) float64 {
		idx := (width*y + x) * 4
		r := float64(pixels[idx])
		g := float64(pixels[idx+1])
		b := float64(pixels[idx+2])

		return 0.299*r + 0.587*g + 0.114*b
	}
	for by := 0; by+blocksize <= height; by += blocksize {
		for bx := 0; bx+blocksize <= width; bx += blocksize {
			sumX := 0.0
			sumY := 0.0
			sumXY := 0.0
			sumXX := 0.0
			sumYY := 0.0
			for y := by; y < by+blocksize; y++ {
				for x := bx; x < bx+blocksize; x++ {
					lcover := getPixelVal(x, y, cover)
					sstego := getPixelVal(x, y, stego)

					sumX += lcover
					sumY += sstego
					sumXY += lcover * sstego
					sumXX += lcover * lcover
					sumYY += sstego * sstego
				}
			}
			muX := coef * sumX
			muY := coef * sumY
			muXX := muX * muX
			muYY := muY * muY
			muXY := muX * muY

			varXX := coef*sumXX - muXX
			varYY := coef*sumYY - muYY

			covXY := coef*sumXY - muXY

			num := (2*muXY + c1) * (2*covXY + c2)
			denom := (muXX + muYY + c1) * (varXX + varYY + c2)

			totalSSIM += num / denom
			totalBlocks++
		}
	}

	if totalBlocks == 0 {
		return 1.0
	}

	mMSIM := totalSSIM / float64(totalBlocks)
	return mMSIM
}

func CalculateChangedRatio(cover, stego []byte) (chCount int, chRatio float64) {
	changedCount := 0
	totalComponents := 0
	for i := 0; i < len(cover); i++ {
		if i%4 == 3 {
			continue
		}
		if cover[i] != stego[i] {
			changedCount++
		}
		totalComponents++

	}

	if totalComponents == 0 {
		return 0, 0.0
	}

	changedRatio := float64(changedCount) / float64(totalComponents)
	return changedCount, changedRatio
}

func CalculateLSBDDistribution(pixels []byte) (p0, p1 float64) {
	countZero := 0
	countOne := 0

	for i := 0; i < len(pixels); i++ {
		if i%4 == 3 {
			continue
		}

		if pixels[i]&1 == 0 {
			countZero++
		} else {
			countOne++
		}
	}
	countTotal := countOne + countZero

	if countTotal == 0 {
		return 0.0, 0.0
	}

	p0 = float64(countZero) / float64(countTotal)
	p1 = float64(countOne) / float64(countTotal)
	return p0, p1
}

func CalculateChiSquare(pixels []byte) (chi2, pValue float64) {
	var hist [256]int
	for i := 0; i < len(pixels); i++ {
		if i%4 == 3 {
			continue
		}
		hist[pixels[i]]++
	}

	activePairs := 0
	chi2 = 0.0

	for k := 0; k < 128; k++ {
		cEven := float64(hist[2*k])
		cOdd := float64(hist[2*k+1])
		sum := cEven + cOdd

		if sum > 0 {
			diff := cEven - cOdd
			chi2 += (diff * diff) / sum
			activePairs++
		}
	}

	df := activePairs - 1
	if df <= 0 {
		return 0.0, 1.0
	}

	kVal := float64(df)
	factor := 2.0 / (9.0 * kVal)
	z := (math.Pow(chi2/kVal, 1.0/3.0) - (1.0 - factor)) / math.Sqrt(factor)

	pValue = 0.5 * math.Erfc(z/math.Sqrt2)

	if pValue < 0.0 {
		pValue = 0.0
	} else if pValue > 1.0 {
		pValue = 1.0
	}

	return chi2, pValue
}

func CalcFrequency(data []byte) [256]int {
	var counts [256]int
	for i, el := range data {
		if i%4 == 3 {
			continue
		}
		counts[el]++
	}
	return counts
}

func SaveHistogram(filePath string, cover, stego [256]int) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"byte", "cover", "stego"}
	err = writer.Write(header)
	if err != nil {
		return fmt.Errorf("не удалось записать заголовок: %w", err)
	}

	for i := 0; i < 256; i++ {
		row := []string{
			strconv.Itoa(i),
			strconv.Itoa(cover[i]),
			strconv.Itoa(stego[i]),
		}
		err = writer.Write(row)
		if err != nil {
			return fmt.Errorf("не удалось записать строку: %w", err)
		}
	}
	return nil
}

func SaveMetrics(filePath string, rows []MetricRow) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Не удалось создать файл: %w", err)
	}

	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Method", "Payload", "PSNR", "SSIM", "CHANGED_RATIO", "P(LSB=1)", "CHI2", "p-value"}
	err = writer.Write(header)
	if err != nil {
		return fmt.Errorf("Не удалось записать заголовок: %w", err)
	}

	for _, row := range rows {
		record := []string{
			row.Method,
			fmt.Sprintf("%d", row.Payload),
			fmt.Sprintf("%f", row.PSNR),
			fmt.Sprintf("%f", row.SSIM),
			fmt.Sprintf("%f", row.ChangedRatio),
			fmt.Sprintf("%f", row.ProbLSB1),
			fmt.Sprintf("%f", row.ChiSquare),
			fmt.Sprintf("%f", row.Pvalue),
		}
		err = writer.Write(record)
		if err != nil {
			return fmt.Errorf("Не удалось записать строку: %w", err)
		}
	}
	return nil
}

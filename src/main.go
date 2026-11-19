package main

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
)

const ivLen = aes.BlockSize
const saltLen = aes.BlockSize
const keyLen = aes.BlockSize
const metricsPathFile string = "results/metrics.csv"
const differencePathFile string = "results/difference.png"

func createOutPath(srcImgPath string, alg string, tag ...string) (srcDecImgPath, srcEncImgPath string) {
	baseName := filepath.Base(srcImgPath)
	ext := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)
	suffix := alg
	if len(tag) > 0 && tag[0] != "" {
		suffix = alg + "_" + tag[0]
	}
	srcDecImgPath = filepath.Join("imgs", name+"-"+suffix+"-dec"+ext)
	srcEncImgPath = filepath.Join("imgs", name+"-"+suffix+"-enc"+ext)

	return srcDecImgPath, srcEncImgPath
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}

func RunPlainLSB(byteMsg []byte, plainPixels []byte, bounds image.Rectangle, srcImgPath string, tag ...string) (cipherPixels []byte, err error) {
	_, srcEncImgPath := createOutPath(srcImgPath, "plainLsb", tag...)

	cipherPixels, err = CipherPlainLsb(plainPixels, byteMsg)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при шифровании изображения: %w", err)
	}

	err = saveImage(cipherPixels, bounds, srcEncImgPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения стегоизображения %s: %w", srcEncImgPath, err)
	}

	cipherPixels, bounds, err = readImage(srcEncImgPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения стегоизображения %s: %w", srcEncImgPath, err)
	}

	decryptedText, err := DecryptPlainLsb(cipherPixels)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при дешифровании изображения: %w", err)
	}

	equalsMsg(decryptedText, byteMsg)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при проверке сообщений: %w", err)
	}
	return cipherPixels, err
}

func RunPlainCipherLSB(byteMsg []byte, password string, iv []byte, plainPixels []byte, bounds image.Rectangle, srcImgPath string, tag ...string) (cipherPixels []byte, err error) {

	cipherMsgUn, err := addHeader2Msg(byteMsg, password, iv)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при добавлении заголовка: %w", err)
	}

	_, srcEncImgPath := createOutPath(srcImgPath, "plainLsbAes", tag...)

	cipherPixels, err = CipherPlainLsb(plainPixels, cipherMsgUn)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при шифровании изображения: %w", err)
	}

	err = saveImage(cipherPixels, bounds, srcEncImgPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения стегоизображения %s: %w", srcEncImgPath, err)
	}

	cipherPixels, bounds, err = readImage(srcEncImgPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения стегоизображения %s: %w", srcEncImgPath, err)
	}

	decryptMsg, err := DecryptPlainLsb(cipherPixels)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при дешифровании изображения: %w", err)
	}

	encryptedText, key, iv, err := getHeaderAndMsg(decryptMsg, password)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при извлечении данных: %w", err)
	}
	decryptedText, err := CTRCipher(encryptedText, key, iv)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при расшифровке данных: %w", err)
	}

	equalsMsg(decryptedText, byteMsg)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при проверке сообщений: %w", err)
	}
	return cipherPixels, err
}

func RunRandLSB(byteMsg []byte, password string, iv []byte, plainPixels []byte, bounds image.Rectangle, srcImgPath string, tag ...string) (cipherPixels []byte, err error) {

	_, srcEncImgPath := createOutPath(srcImgPath, "encryptedRandomLsb", tag...)

	cipherMsgUn, err := addHeader2Msg(byteMsg, password, iv)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при добавлении заголовка: %w", err)
	}

	cipherPixels, err = CipherRandLsb(plainPixels, cipherMsgUn, password)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при шифровании изображения: %w", err)
	}

	err = saveImage(cipherPixels, bounds, srcEncImgPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения стегоизображения %s: %w", srcEncImgPath, err)
	}

	cipherPixels, bounds, err = readImage(srcEncImgPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения стегоизображения %s: %w", srcEncImgPath, err)
	}

	decryptMsg, err := DecryptRandLsb(cipherPixels, password)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при дешифровании изображения: %w", err)
	}

	encryptedText, key, iv, err := getHeaderAndMsg(decryptMsg, password)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при извлечении данных: %w", err)
	}
	decryptedText, err := CTRCipher(encryptedText, key, iv)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при расшифровке данных: %w", err)
	}

	err = equalsMsg(decryptedText, byteMsg)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при проверке сообщений: %w", err)
	}
	return cipherPixels, err
}

func addHeader2Msg(byteMsg []byte, password string, iv []byte) (msg []byte, err error) {
	salt, err := GenerateSalt(saltLen)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при генерации соли: %w", err)
	}

	key, err := DeriveKey(password, salt, keyLen)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при генерации ключа: %w", err)
	}

	cipherMsg, err := CTRCipher(byteMsg, key, iv)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при шифровании сообщения: %w", err)
	}

	cipherMsgUn := make([]byte, 0, len(salt)+len(iv)+len(cipherMsg))
	cipherMsgUn = append(cipherMsgUn, salt...)
	cipherMsgUn = append(cipherMsgUn, iv...)
	cipherMsgUn = append(cipherMsgUn, cipherMsg...)

	return cipherMsgUn, nil
}

func getHeaderAndMsg(decryptedMsg []byte, password string) (encryptedText []byte, key []byte, iv []byte, err error) {
	headerLen := saltLen + ivLen
	if len(decryptedMsg) < headerLen {
		return nil, nil, nil, fmt.Errorf("Размер сообщениея меньше длины заголовка.")
	}
	salt := decryptedMsg[0:saltLen]
	iv = decryptedMsg[saltLen : saltLen+ivLen]
	encryptedText = decryptedMsg[saltLen+ivLen:]

	key, err = DeriveKey(password, salt, keyLen)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Ошибка при извлечении ключа: %w", err)
	}
	return encryptedText, key, iv, nil
}

func equalsMsg(decryptedMsg, plainMsg []byte) error {
	if !bytes.Equal(decryptedMsg, plainMsg) {
		return fmt.Errorf("Сообщения не совпдают")
	}

	extractedText := string(decryptedMsg)
	preview := extractedText
	if len([]rune(extractedText)) > 20 {
		preview = string([]rune(extractedText)[:20])
	}
	fmt.Printf("расшифрованное сообщение: %s\n", preview)
	return nil
}

func addMetrics(cover, stego []byte, bounds image.Rectangle, method string, payload float64, metrics []MetricRow) []MetricRow {
	payloadPercent := int(payload)
	psnr := CalculatePSNR(cover, stego)
	ssim := CalculateSSIM(cover, stego, bounds)
	chCount, chRatio := CalculateChangedRatio(cover, stego)
	p0, p1 := CalculateLSBDDistribution(stego)
	chi2, pValue := CalculateChiSquare(stego)
	currMetrics := []MetricRow{
		{method, payloadPercent, psnr, ssim, chCount, chRatio, p0, p1, chi2, pValue},
	}
	return append(metrics, currMetrics...)
}

func printMetrics(metric MetricRow, method string) {
	fmt.Printf("============ %s ============\nPSNR: %f, SSIM: %f, changed_count: %d, changed_ratio: %f, P(LSB=0): %f, P(LSB=1): %f, chi2: %f, p-value: %f\n\n", method, metric.PSNR, metric.SSIM, metric.ChangedCount, metric.ChangedRatio, metric.ProbLSB0, metric.ProbLSB1, metric.ChiSquare, metric.Pvalue)
}
func main() {
	histogramImgPath := getEnv("HISTOGRAM_IMG_PATH", "results/histogram.png")
	histogramCsvPath := getEnv("HISTOGRAM_CSV_PATH", "results/histogram.csv")
	srcImgPath := getEnv("SRC_IMG_PATH", "imgs/photo.png")

	_ = os.Remove(histogramImgPath)
	_ = os.Remove(histogramCsvPath)

	if len(os.Args) > 1 {
		srcImgPath = os.Args[1]
	}

	fmt.Println("Работа с", srcImgPath)
	cover, bounds, err := readImage(srcImgPath)
	password := "Penumbra Phantasm"

	var metrics []MetricRow
	metrics = addMetrics(cover, cover, bounds, "Cover", 0.0, metrics)
	last := metrics[len(metrics)-1]
	printMetrics(last, "СOVER")

	//=============== Проверка шифрования при разных размерах текста
	payloadLevels := []float64{50.0, 100.0}

	var stego []byte
	var stego2 []byte
	var stego3 []byte

	//генерация iv
	iv, err := GenerateIV(ivLen)
	if err != nil {
		fmt.Println("Сбой работы при генерации IV: ", err)
		return
	}

	for i, percent := range payloadLevels {
		fmt.Printf("=========== Заход %d (%d%%) ===========\n\n", i+1, int(percent))

		iterationTag := fmt.Sprintf("%d-percent", int(percent))
		//генерация сообщения

		targetBytes := CalculatePayloadSize(bounds, percent, 0, 0)
		plainMsg := GenerateFakeText(targetBytes)
		if len(plainMsg) < saltLen+ivLen {
			fmt.Printf("Ошибка в %d подходе: Слишком маленькое сообщение, соль и iv не влезают", i)
		}
		plainMsg = plainMsg[:len(plainMsg)-saltLen-ivLen]
		//==== Работа PlainLSB

		stego, err = RunPlainLSB(plainMsg, cover, bounds, srcImgPath, iterationTag)
		if err != nil {
			fmt.Println("Ошибка при работе с PlainLSB: ", err)
			return
		}

		metrics = addMetrics(cover, stego, bounds, "PlainLSB", percent, metrics)
		last := metrics[len(metrics)-1]

		printMetrics(last, "PLAINLSB")

		//Генерация нового сообщения из-за добавления шапки
		// if len(plainMsg) < saltLen+ivLen {
		// 	fmt.Printf("Ошибка в %d подходе: Слишком маленькое сообщение, соль и iv не влезают", i)
		// }
		// plainMsg = plainMsg[:len(plainMsg)-saltLen-ivLen]

		//===== Работа PlainCipherLSB

		stego2, err = RunPlainCipherLSB(plainMsg, password, iv, cover, bounds, srcImgPath, iterationTag)
		if err != nil {
			fmt.Println("Ошибка при работе с PlainCipherLSB: ", err)
			return
		}

		metrics = addMetrics(cover, stego2, bounds, "PlainCipherLSB", percent, metrics)
		last = metrics[len(metrics)-1]

		printMetrics(last, "PLAINCIPHERLSB")

		// ====== Работа randLSB

		stego3, err = RunRandLSB(plainMsg, password, iv, cover, bounds, srcImgPath, iterationTag)
		if err != nil {
			fmt.Println("Ошибка при работе с RandLSB: ", err)
			return
		}

		metrics = addMetrics(cover, stego3, bounds, "RandLSB", percent, metrics)
		last = metrics[len(metrics)-1]

		printMetrics(last, "RANDLSB")
	}

	err = SaveMetrics(metricsPathFile, metrics)
	if err != nil {
		fmt.Println("Сбой при сохранении метрик: ", err)
		return
	}
	fmt.Println("Метрики сохранены в: ", metricsPathFile)

	//карта разницы
	percent := 15.0
	targetBytes := CalculatePayloadSize(bounds, percent, 0, 0)
	plainMsg := GenerateFakeText(targetBytes)
	plainMsg = plainMsg[:len(plainMsg)-ivLen-saltLen]
	stego4, err := RunRandLSB(plainMsg, password, iv, cover, bounds, srcImgPath)
	if err != nil {
		fmt.Println("Ошибка при работе с PlainCipherLSB: ", err)
		return
	}
	err = SaveDiffImage(cover, stego4, bounds, differencePathFile)
	if err != nil {
		fmt.Println("Ошибка при сохранении карты разности: ", err)
		return
	}
	fmt.Println("Изображение разницы сохранено в:", differencePathFile)

	coverFreq := CalcFrequency(cover)
	StegoFreq := CalcFrequency(stego4)
	err = SaveHistogram(histogramCsvPath, coverFreq, StegoFreq)
	if err != nil {
		fmt.Println("Ошибка при сохранении данных для гистограммы:", err)
		return
	}
	fmt.Println("Данные для гистограммы сохранены в:", histogramCsvPath)
}

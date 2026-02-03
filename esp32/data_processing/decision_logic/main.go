package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/stat"

	"go.bug.st/serial"
)

const (
	FS_MIC    = 40000.0
	FS_ACC    = 10000.0
	FRAME_SIZE = 1024
)

// --- Feature struct ---
type Features struct {
	RMS  float64
	Kurt float64
	Dom  float64
	Low  float64
	Mid  float64
	High float64
}

func (f Features) ToSlice() []float64 {
	return []float64{f.RMS, f.Kurt, f.Dom, f.Low, f.Mid, f.High}
}

// --- CSV baseline loading ---
func loadCSV(path string) ([]float64, []float64, float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer file.Close()

	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, nil, 0, err
	}

	var mic, acc, temps []float64
	for i, row := range rows {
		if i == 0 || len(row) < 5 {
			continue
		}
		a, _ := strconv.ParseFloat(row[1], 64)
		v, _ := strconv.ParseFloat(row[2], 64)
		t, _ := strconv.ParseFloat(row[4], 64)
		mic = append(mic, a)
		acc = append(acc, v)
		temps = append(temps, t)
	}

	micMean := stat.Mean(mic, nil)
	accMean := stat.Mean(acc, nil)
	for i := range mic {
		mic[i] -= micMean
		acc[i] -= accMean
	}

	return mic, acc, stat.Mean(temps, nil), nil
}

func getFeatures(signal []float64, fs float64) Features {
	n := len(signal)
	if n == 0 {
		return Features{}
	}

	// RMS
	var sumSq float64
	for _, v := range signal {
		sumSq += v * v
	}
	rms := math.Sqrt(sumSq / float64(n))

	// Kurtosis
	kurt := stat.ExKurtosis(signal, nil) + 3.0

	fft := fourier.NewFFT(n)
	coeffs := fft.Coefficients(nil, signal)

	var low, mid, high float64
	maxAmp := -1.0
	domIdx := 0

	for i := 0; i < len(coeffs); i++ {
		c := coeffs[i]
		amp := math.Hypot(real(c), imag(c))
		freq := float64(i) * fs / float64(n)

		if freq >= 10 && freq < 200 {
			low += amp
		} else if freq >= 200 && freq < 1000 {
			mid += amp
		} else if freq >= 1000 && freq < 3000 {
			high += amp
		}

		if amp > maxAmp {
			maxAmp = amp
			domIdx = i
		}
	}

	return Features{
		RMS:  rms,
		Kurt: kurt,
		Dom:  float64(domIdx) * fs / float64(n),
		Low:  low,
		Mid:  mid,
		High: high,
	}
}

func processFolder(folder string) ([]float64, []float64, float64) {
	files, err := filepath.Glob(filepath.Join(folder, "*.csv"))
	if err != nil {
		log.Fatal(err)
	}

	var micAgg, accAgg []float64
	var tSum float64
	count := 0

	for _, f := range files {
		mic, acc, t, err := loadCSV(f)
		if err != nil || len(mic) == 0 {
			continue
		}
		micFeats := getFeatures(mic, FS_MIC).ToSlice()
		accFeats := getFeatures(acc, FS_ACC).ToSlice()

		if micAgg == nil {
			micAgg = make([]float64, 6)
			accAgg = make([]float64, 6)
		}

		for i := 0; i < 6; i++ {
			micAgg[i] += micFeats[i]
			accAgg[i] += accFeats[i]
		}
		tSum += t
		count++
	}

	if count == 0 {
		return make([]float64, 6), make([]float64, 6), 0
	}

	for i := 0; i < 6; i++ {
		micAgg[i] /= float64(count)
		accAgg[i] /= float64(count)
	}
	return micAgg, accAgg, tSum / float64(count)
}

// --- Sensor fusion logic ---
func sensorFusionDiagnosis(ac, vb []float64, t float64, hAc, hVb []float64, hT float64) (string, string, float64, float64) {
	vKurtShift := ((vb[1] - hVb[1]) / hVb[1]) * 100
	vDomShift := ((vb[2] - hVb[2]) / hVb[2]) * 100
	aEnergyShift := ((ac[4] + ac[5]) / (hAc[4] + hAc[5]) - 1) * 100
	tempDiff := t - hT

	status, origin, confidence := "HEALTHY", "Normal Operation", 0.0

	if vKurtShift > 15 && vDomShift > 50 {
		status = "CRITICAL"
		origin = "Vibration kurtosis increased >15%, dom freq >50%"
		confidence = 0.85
		if tempDiff > 5 {
			confidence = 0.95
		}
	} else if aEnergyShift > 25 {
		status = "WARNING"
		origin = "Acoustic energy increased >25%"
		confidence = 0.70
		if tempDiff > 3 {
			confidence = 0.80
		}
	} else if vKurtShift > 10 {
		status = "ADVISORY"
		origin = "Vibration kurtosis increased >10%"
		confidence = 0.60
	}

	return status, origin, confidence, tempDiff
}

// --- Serial frame reading ---
func readFrames(port serial.Port, micBuf, accBuf *[]float64, tempCh chan float64) {
	frameMicSize := 7  // 1 + 4 + 2
	frameAccSize := 7  // 1 + 4 + 2
	frameTempSize := 9 // 1 + 4 + 4

	buf := make([]byte, 256)
	for {
		n, err := port.Read(buf)
		if err != nil {
			continue
		}
		i := 0
		for i < n {
			switch buf[i] {
			case 0xA5:
				if i+frameMicSize <= n {
					val := int16(binary.LittleEndian.Uint16(buf[i+5:]))
					*micBuf = append(*micBuf, float64(val))
					i += frameMicSize
				} else {
					break
				}
			case 0xA6:
				if i+frameAccSize <= n {
					val := int16(binary.LittleEndian.Uint16(buf[i+5:]))
					*accBuf = append(*accBuf, float64(val))
					i += frameAccSize
				} else {
					break
				}
			case 0xB6:
				if i+frameTempSize <= n {
					val := math.Float32frombits(binary.LittleEndian.Uint32(buf[i+5:]))
					tempCh <- float64(val)
					i += frameTempSize
				} else {
					break
				}
			default:
				i++
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <healthy_data_folder>")
	}
	folder := os.Args[1]

	// --- Healthy baseline ---
	hMic, hAcc, hT := processFolder(folder)
	fmt.Println("--- Healthy Baseline ---")
	fmt.Printf("Mic: %v\nAcc: %v\nTemp: %.2f\n", hMic, hAcc, hT)

	// --- Serial setup ---
	mode := &serial.Mode{BaudRate: 2000000}
	port, err := serial.Open("/dev/ttyACM0", mode)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	var micBuf, accBuf []float64
	tempCh := make(chan float64, 10)
	var currentTemp float64

	go readFrames(port, &micBuf, &accBuf, tempCh)

	// --- Real-time processing ---
	for {
		select {
		case t := <-tempCh:
			currentTemp = t
		default:
		}

		if len(micBuf) >= FRAME_SIZE && len(accBuf) >= FRAME_SIZE {
			micSlice := micBuf[:FRAME_SIZE]
			accSlice := accBuf[:FRAME_SIZE]
			micBuf = micBuf[FRAME_SIZE:]
			accBuf = accBuf[FRAME_SIZE:]

			micFeatures := getFeatures(micSlice, FS_MIC)
			accFeatures := getFeatures(accSlice, FS_ACC)

			status, origin, conf, tempRise := sensorFusionDiagnosis(
				micFeatures.ToSlice(), accFeatures.ToSlice(),
				currentTemp, hMic, hAcc, hT,
			)

			fmt.Printf("\n--- SENSOR FUSION ---\n")
			fmt.Printf("STATUS: %s | ORIGIN: %s | CONF: %.1f%% | ΔT: %.2f°C\n",
				status, origin, conf*100, tempRise)
		}
		time.Sleep(10 * time.Millisecond)
	}
}


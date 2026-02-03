package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/stat"
)

const FS = 10000.0 // sampling frequency Hz

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

// Load CSV sensor data
func loadSensorData(path string) ([]float64, []float64, float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, 0, err
	}

	var ac, vb, temps []float64
	for i, row := range rows {
		if i == 0 || len(row) < 5 {
			continue
		}
		a, _ := strconv.ParseFloat(row[1], 64)
		v, _ := strconv.ParseFloat(row[2], 64)
		t, _ := strconv.ParseFloat(row[4], 64)
		ac = append(ac, a)
		vb = append(vb, v)
		temps = append(temps, t)
	}

	acMean := stat.Mean(ac, nil)
	vbMean := stat.Mean(vb, nil)
	for i := range ac {
		ac[i] -= acMean
		vb[i] -= vbMean
	}

	return ac, vb, stat.Mean(temps, nil), nil
}

// Extract RMS, kurtosis, dominant frequency, and low/mid/high band energy
func getFeatures(signal []float64) Features {
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

	// FFT
	fft := fourier.NewFFT(n)
	coeffs := fft.Coefficients(nil, signal)

	var low, mid, high float64
	maxAmp := -1.0
	domIdx := 0

	for i := 0; i < len(coeffs); i++ {
		c := coeffs[i]
		amp := math.Hypot(real(c), imag(c))
		freq := float64(i) * FS / float64(n)

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
		Dom:  float64(domIdx) * FS / float64(n),
		Low:  low,
		Mid:  mid,
		High: high,
	}
}

// Aggregate all CSV files in a directory
func processDir(dir string) ([]float64, []float64, float64) {
	files, _ := filepath.Glob(filepath.Join(dir, "*.csv"))
	var acAgg, vbAgg []float64
	var tSum float64
	count := 0

	for _, f := range files {
		ac, vb, t, err := loadSensorData(f)
		if err != nil || len(ac) == 0 {
			continue
		}

		aFeats := getFeatures(ac).ToSlice()
		vFeats := getFeatures(vb).ToSlice()

		if acAgg == nil {
			acAgg, vbAgg = make([]float64, 6), make([]float64, 6)
		}

		for i := 0; i < 6; i++ {
			acAgg[i] += aFeats[i]
			vbAgg[i] += vFeats[i]
		}
		tSum += t
		count++
	}

	if count == 0 {
		return make([]float64, 6), make([]float64, 6), 0
	}

	for i := 0; i < 6; i++ {
		acAgg[i] /= float64(count)
		vbAgg[i] /= float64(count)
	}
	return acAgg, vbAgg, tSum / float64(count)
}

// Sensor fusion decision with normal thresholds and English descriptions
func sensorFusionDiagnosis(ac, vb []float64, t float64, hAc, hVb []float64, hT float64) (string, string, float64, float64) {
	vKurtShift := ((vb[1] - hVb[1]) / hVb[1]) * 100
	vDomShift := ((vb[2] - hVb[2]) / hVb[2]) * 100
	aEnergyShift := ((ac[4] + ac[5]) / (hAc[4] + hAc[5]) - 1) * 100
	tempDiff := t - hT

	status, origin, confidence := "HEALTHY", "Normal Operation", 0.0

	if vKurtShift > 15 && vDomShift > 50 {
		status = "CRITICAL"
		origin = "Vibration kurtosis increased more than 15%, dominant frequency increased more than 50%"
		confidence = 0.85
		if tempDiff > 5 {
			confidence = 0.95
		}
	} else if aEnergyShift > 25 {
		status = "WARNING"
		origin = "Acoustic energy increased more than 25%"
		confidence = 0.70
		if tempDiff > 3 {
			confidence = 0.80
		}
	} else if vKurtShift > 10 {
		status = "ADVISORY"
		origin = "Vibration kurtosis increased more than 10%"
		confidence = 0.60
	}

	return status, origin, confidence, tempDiff
}

func main() {
	hAc, hVb, hT := processDir("./healthy")
	fAc, fVb, fT := processDir("./faulty")

	status, origin, conf, tRise := sensorFusionDiagnosis(fAc, fVb, fT, hAc, hVb, hT)

	fmt.Printf("\n--- SENSOR FUSION DIAGNOSIS ---\n")
	fmt.Printf("STATUS:     %s\n", status)
	fmt.Printf("ORIGIN:     %s\n", origin)
	fmt.Printf("CONFIDENCE: %.1f%%\n", conf*100)
	fmt.Println("---------------------------------------------")
	fmt.Printf("%-20s | %-10s | %-10s\n", "Feature Source", "Healthy", "Faulty")
	fmt.Printf("%-20s | %-10.4f | %-10.4f\n", "Vibe Kurtosis", hVb[1], fVb[1])
	fmt.Printf("%-20s | %-10.1f | %-10.1f\n", "Vibe Dom Freq (Hz)", hVb[2], fVb[2])
	fmt.Printf("%-20s | %-10.0f | %-10.0f\n", "Acoustic Energy", hAc[4]+hAc[5], fAc[4]+fAc[5])
	fmt.Printf("%-20s | %-10.2f | %-10.2f (Δ %+.1f)\n", "Temperature (°C)", hT, fT, tRise)
}


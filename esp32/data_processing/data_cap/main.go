package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
	"unsafe"

	"go.bug.st/serial"
)

const (
	SerialPort      = "/dev/ttyACM0" // Change to your port
	BaudRate        = 2000000
	MicSyncByte     = 0xA5
	AccSyncByte     = 0xA6
	TempSyncByte    = 0xB6
	DiagSyncByte    = 0xC7
	CaptureDuration = 60 * time.Second // Total capture duration
	SplitSamples    = 40000            // Split CSV every 40k samples
)

// Frame values
var micVal int16
var accVal int16
var tempVal float32

// Directory to save CSVs
var outputDir string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <output_directory>")
		return
	}
	outputDir = os.Args[1]

	mode := &serial.Mode{
		BaudRate: BaudRate,
	}
	port, err := serial.Open(SerialPort, mode)
	if err != nil {
		panic(err)
	}
	defer port.Close()

	sampleCount := 0

	csvFile, csvWriter := newCSVFile()
	defer csvFile.Close()

	start := time.Now()
	buf := make([]byte, 8) // Largest frame

	for time.Since(start) < CaptureDuration {
		b := make([]byte, 1)
		_, err := port.Read(b)
		if err != nil {
			continue
		}

		switch b[0] {
		case MicSyncByte:
			_, err := port.Read(buf[:6])
			if err != nil {
				continue
			}
			micVal = int16(buf[4]) | int16(buf[5])<<8

		case AccSyncByte:
			_, err := port.Read(buf[:6])
			if err != nil {
				continue
			}
			accVal = int16(buf[4]) | int16(buf[5])<<8

		case TempSyncByte:
			_, err := port.Read(buf[:8])
			if err != nil {
				continue
			}
			tempVal = float32FromBytes(buf[4:8])

		case DiagSyncByte:
			skip := make([]byte, 24)
			port.Read(skip)
			continue

		default:
			continue
		}

		csvWriter.Write([]string{
			fmt.Sprintf("%d", sampleCount),
			fmt.Sprintf("%d", micVal),
			fmt.Sprintf("%d", accVal),
			fmt.Sprintf("%.3f", tempVal),
		})
		sampleCount++

		if sampleCount%SplitSamples == 0 {
			csvWriter.Flush()
			csvFile.Close()
			csvFile, csvWriter = newCSVFile()
			defer csvFile.Close()
		}

		csvWriter.Flush()
	}

	fmt.Println("Capture complete.")
}

// newCSVFile creates a CSV file in the specified output directory with date-time-based unique name
func newCSVFile() (*os.File, *csv.Writer) {
	timestamp := time.Now().Format("20060102_150405") // YYYYMMDD_HHMMSS
	name := fmt.Sprintf("healthy_data_%s.csv", timestamp)
	fullPath := filepath.Join(outputDir, name)
	f, err := os.Create(fullPath)
	if err != nil {
		panic(err)
	}
	writer := csv.NewWriter(f)
	writer.Write([]string{"sample", "mic", "acc", "temperature_C"})
	writer.Flush()
	return f, writer
}

// Convert 4 bytes to float32 (little endian)
func float32FromBytes(b []byte) float32 {
	if len(b) != 4 {
		return 0
	}
	bits := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return *(*float32)(unsafe.Pointer(&bits))
}


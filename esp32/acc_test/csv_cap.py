import serial
import csv
import time
import sys
import select
import termios
import tty
from datetime import datetime

PORT = "/dev/ttyACM0"   # adjust
BAUD = 1_000_000

# ---- put terminal into raw mode ----
fd = sys.stdin.fileno()
old_settings = termios.tcgetattr(fd)
tty.setcbreak(fd)

def key_pressed():
    return select.select([sys.stdin], [], [], 0)[0]

ser = serial.Serial(PORT, BAUD, timeout=0.1)
time.sleep(2)

print("s = start capture | q = quit")

capturing = False
csv_file = None
writer = None

try:
    while True:
        # ---- keyboard ----
        if key_pressed():
            key = sys.stdin.read(1)

            if key == "s" and not capturing:
                ts = datetime.now().strftime("%H-%M-%S")
                filename = f"adxl335_{ts}.csv"
                csv_file = open(filename, "w", newline="")
                writer = csv.writer(csv_file)
                writer.writerow(["adc", "g"])
                capturing = True
                print(f"\nCAPTURE STARTED → {filename}")

            elif key == "q":
                print("\nEXIT")
                break

        # ---- serial ----
        if capturing:
            line = ser.readline().decode(errors="ignore").strip()
            if not line:
                continue

            if line.startswith("BURST_SAMPLES_CAPTURED"):
                print(line)
                capturing = False
                csv_file.close()
                writer = None
                print("WAITING")
                continue

            if "," in line:
                try:
                    adc, g = line.split(",")
                    writer.writerow([int(adc), float(g)])
                except ValueError:
                    pass

finally:
    termios.tcsetattr(fd, termios.TCSADRAIN, old_settings)
    ser.close()


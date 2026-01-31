import serial
import struct
import csv
from datetime import datetime
import os

# ================= CONFIG =================
PORT = "/dev/ttyACM0"
BAUD = 1_000_000

FRAME_MARKER = b'\xAA'
SAMPLE_FMT = "<IffHf"
SAMPLE_SIZE = struct.calcsize(SAMPLE_FMT)

ROLLING_SAMPLES = 10_000
BURSTS_PER_FILE = 30  # New threshold
OUTPUT_DIR = "faulty"
os.makedirs(OUTPUT_DIR, exist_ok=True)

# ================= SERIAL =================
ser = serial.Serial(PORT, BAUD, timeout=None)

print(f"Receiving binary frames. Saving every {BURSTS_PER_FILE} bursts...")

samples = []
burst_count = 0
file_idx = 0

def save_csv(buf, idx):
    ts = datetime.now().strftime("%H-%M-%S")
    fname = os.path.join(OUTPUT_DIR, f"esp32_{ts}_batch_{idx}.csv")

    with open(fname, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["ts_us", "acous", "acc", "ct", "temp"])
        w.writerows(buf)

    print(f"Saved {len(buf)} samples ({BURSTS_PER_FILE} bursts) → {fname}")

# ================= MAIN LOOP =================
while True:
    # 1. Sync
    b = ser.read(1)
    if b != FRAME_MARKER:
        continue

    # 2. Read count
    count_bytes = ser.read(2)
    if len(count_bytes) != 2:
        continue
    count = struct.unpack("<H", count_bytes)[0]

    # 3. Read payload
    payload_len = count * SAMPLE_SIZE
    raw = ser.read(payload_len)
    if len(raw) != payload_len:
        continue

    # 4. Unpack chunk and append to buffer
    for i in range(count):
        off = i * SAMPLE_SIZE
        sample = struct.unpack(SAMPLE_FMT, raw[off : off + SAMPLE_SIZE])
        samples.append(sample)

    # 5. Increment burst counter and check threshold
    burst_count += 1
    if burst_count >= BURSTS_PER_FILE:
        save_csv(samples, file_idx)
        samples = []      # Clear buffer for next batch
        burst_count = 0   # Reset burst tracker
        file_idx += 1

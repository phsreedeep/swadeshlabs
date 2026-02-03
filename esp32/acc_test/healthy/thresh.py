import numpy as np
import pandas as pd

# ================= CONFIG =================

CSV_PATH = "esp32_21-53-55_batch_0.csv"   # path to healthy data
FS = 20000                  # sampling rate (Hz)
N = 1024                   # window size (samples)

# Band edges (Hz)
LOW_BAND  = (0, 500)
MID_BAND  = (500, 2000)
HIGH_BAND = (2000, FS / 2)

# ================= LOAD CSV =================

data = pd.read_csv(CSV_PATH, header=None).values
# select 3rd column only (zero-indexed)
accel_data = data[:, 2].astype(np.float32)

# Trim to whole windows
num_windows = len(accel_data) // N
accel_data = accel_data[:num_windows * N]
windows = accel_data.reshape(num_windows, N)

# ================= FFT SETUP =================

freqs = np.fft.rfftfreq(N, d=1.0 / FS)

def band_energy(fft_mag_sq, band):
    idx = np.where((freqs >= band[0]) & (freqs < band[1]))[0]
    return np.sum(fft_mag_sq[idx])

# ================= FEATURE EXTRACTION =================

R_low  = []
R_mid  = []
R_high = []

for w in windows:
    w = w - np.mean(w)                # remove DC
    fft = np.fft.rfft(w)
    mag_sq = np.abs(fft) ** 2

    E_low  = band_energy(mag_sq, LOW_BAND)
    E_mid  = band_energy(mag_sq, MID_BAND)
    E_high = band_energy(mag_sq, HIGH_BAND)
    E_tot  = E_low + E_mid + E_high

    if E_tot == 0:
        continue

    R_low.append(E_low / E_tot)
    R_mid.append(E_mid / E_tot)
    R_high.append(E_high / E_tot)

R_low  = np.array(R_low)
R_mid  = np.array(R_mid)
R_high = np.array(R_high)

# ================= THRESHOLDS =================
# Robust percentile-based thresholds

T_low  = np.percentile(R_low,  99.5)
T_mid  = np.percentile(R_mid,  99.5)
T_high = np.percentile(R_high, 99.5)

# ================= OUTPUT =================

print("=== HEALTHY BASELINE THRESHOLDS ===")
print(f"T_LOW  = {T_low:.4f}")
print(f"T_MID  = {T_mid:.4f}")
print(f"T_HIGH = {T_high:.4f}")

print("\n=== MEANS ===")
print(f"μ_LOW  = {R_low.mean():.4f}")
print(f"μ_MID  = {R_mid.mean():.4f}")
print(f"μ_HIGH = {R_high.mean():.4f}")

print("\n=== STDS ===")
print(f"σ_LOW  = {R_low.std():.4f}")
print(f"σ_MID  = {R_mid.std():.4f}")
print(f"σ_HIGH = {R_high.std():.4f}")


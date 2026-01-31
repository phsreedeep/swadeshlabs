import csv
import numpy as np
from scipy.fft import rfft, rfftfreq
from scipy.stats import kurtosis

FS = 20000.0  # sampling rate (Hz)

def load_g(fname):
    g = []
    with open(fname, "r") as f:
        reader = csv.reader(f)
        next(reader)  # skip header
        for row in reader:
            g.append(float(row[1]))
    return np.array(g)

def analyze(g):
    N = len(g)

    rms = np.sqrt(np.mean(g**2))
    peak = np.max(np.abs(g))
    peak_to_peak = np.max(g) - np.min(g)
    crest = peak / rms
    kurt = kurtosis(g, fisher=False)

    # FFT
    yf = np.abs(rfft(g))
    xf = rfftfreq(N, 1 / FS)
    dom_freq = xf[np.argmax(yf)]

    # Band energies
    low = np.sum(yf[(xf >= 10) & (xf < 200)])
    mid = np.sum(yf[(xf >= 200) & (xf < 1000)])
    high = np.sum(yf[(xf >= 1000) & (xf < 3000)])

    zcr = np.sum(np.diff(np.signbit(g))) / N

    return {
        "RMS (g)": rms,
        "Peak-to-Peak (g)": peak_to_peak,
        "Crest Factor": crest,
        "Kurtosis": kurt,
        "Dominant Freq (Hz)": dom_freq,
        "Low Band Energy": low,
        "Mid Band Energy": mid,
        "High Band Energy": high,
        "Zero Crossing Rate": zcr
    }

# ---- load data ----
g_healthy = load_g("./healthy/esp32_21-53-55_batch_0.csv")
g_faulty  = load_g("./faulty/esp32_21-44-56_batch_0.csv")

res_h = analyze(g_healthy)
res_f = analyze(g_faulty)

# ---- print comparison ----
print("\nFEATURE COMPARISON\n")

for k in res_h:
    print(f"{k:20s} | Healthy: {res_h[k]:10.4f} | Faulty: {res_f[k]:10.4f}")


import csv
import numpy as np
from pathlib import Path
from scipy.fft import rfft, rfftfreq
from scipy.stats import kurtosis

FS = 10000.0  # Sampling rate (Hz)

def load_sensor_data(filepath):
    """Parses Acoustic (Col 1), Acc (Col 2), and Temp (Col 4)."""
    ac_data, vb_data, temp_data = [], [], []
    try:
        with open(filepath, "r") as f:
            reader = csv.reader(f)
            next(reader) 
            for row in reader:
                if len(row) > 4:
                    ac_data.append(float(row[1]))
                    vb_data.append(float(row[2]))
                    temp_data.append(float(row[4]))
    except (IndexError, ValueError):
        return None, None, None
    
    ac, vibe = np.array(ac_data), np.array(vb_data)
    # Zero-center AC and Vibe; Keep Temp absolute
    if ac.size > 0:
        ac -= np.mean(ac)
        vibe -= np.mean(vibe)
        
    return ac, vibe, np.mean(temp_data)

def get_features(signal):
    """Calculates RMS, Kurtosis, Dom Freq, and Band Energies."""
    N = len(signal)
    if N == 0: return None
    rms = np.sqrt(np.mean(signal**2))
    kurt = kurtosis(signal, fisher=False)
    yf = np.abs(rfft(signal))
    xf = rfftfreq(N, 1/FS)
    dom_freq = xf[np.argmax(yf)]
    low = np.sum(yf[(xf >= 10) & (xf < 200)])
    mid = np.sum(yf[(xf >= 200) & (xf < 1000)])
    high = np.sum(yf[(xf >= 1000) & (xf < 3000)])
    
    return {"rms": rms, "kurt": kurt, "dom": dom_freq, "low": low, "mid": mid, "high": high}

def process_dir(directory):
    path = Path(directory)
    ac_feats, vb_feats, temps = [], [], []
    
    for f in path.glob("*.csv"):
        ac, vibe, t_avg = load_sensor_data(f)
        if ac is not None:
            a_res, v_res = get_features(ac), get_features(vibe)
            if a_res and v_res:
                ac_feats.append(list(a_res.values()))
                vb_feats.append(list(v_res.values()))
                temps.append(t_avg)
                
    if not ac_feats: return np.zeros(6), np.zeros(6), 0
    return np.mean(ac_feats, axis=0), np.mean(vb_feats, axis=0), np.mean(temps)

def sensor_fusion_diagnosis(ac_avg, vb_avg, t_avg, h_ac, h_vb, h_t):
    # 1. Calculate Shifts
    v_kurt_shift = ((vb_avg[1] - h_vb[1]) / h_vb[1]) * 100
    v_dom_shift = ((vb_avg[2] - h_vb[2]) / h_vb[2]) * 100
    a_energy_shift = ((ac_avg[4] + ac_avg[5]) / (h_ac[4] + h_ac[5]) - 1) * 100
    temp_diff = t_avg - h_t
    
    # 2. Decision Logic
    status, origin, confidence = "HEALTHY", "Normal Operation", 0.0
    
    # Critical: Vibration Impact + High Frequency + Thermal Rise
    if v_kurt_shift > 15 and v_dom_shift > 50:
        status = "CRITICAL"
        origin = "Bearing Spalling / Structural Damage"
        confidence = 0.95 if temp_diff > 5 else 0.85
    # Warning: Acoustic energy rise without vibration impacts
    elif a_energy_shift > 25:
        status = "WARNING"
        origin = "Lubrication Failure / Friction"
        confidence = 0.80 if temp_diff > 3 else 0.70
    # Advisory: Small vibration changes
    elif v_kurt_shift > 10:
        status = "ADVISORY"
        origin = "Early Mechanical Looseness"
        confidence = 0.60

    return status, origin, confidence, temp_diff

# ---- EXECUTION ----
h_ac, h_vb, h_t = process_dir("./healthy")
f_ac, f_vb, f_t = process_dir("./faulty")

status, origin, conf, t_rise = sensor_fusion_diagnosis(f_ac, f_vb, f_t, h_ac, h_vb, h_t)

print(f"\n--- SENSOR FUSION DIAGNOSIS ---")
print(f"STATUS:     {status}")
print(f"ORIGIN:     {origin}")
print(f"CONFIDENCE: {conf*100:.1f}%")
print("-" * 45)
print(f"{'Feature Source':20s} | {'Healthy':10s} | {'Faulty':10s}")
print(f"{'Vibe Kurtosis':20s} | {h_vb[1]:10.4f} | {f_vb[1]:10.4f}")
print(f"{'Vibe Dom Freq (Hz)':20s} | {h_vb[2]:10.1f} | {f_vb[2]:10.1f}")
print(f"{'Acoustic Energy':20s} | {h_ac[4]+h_ac[5]:10.0f} | {f_ac[4]+f_ac[5]:10.0f}")
print(f"{'Temperature (°C)':20s} | {h_t:10.2f} | {f_t:10.2f} (Δ {t_rise:+.1f})")

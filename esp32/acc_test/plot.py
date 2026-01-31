import csv
import matplotlib.pyplot as plt
from datetime import datetime

def load_csv(fname):
    adc = []
    g = []
    with open(fname, "r") as f:
        reader = csv.reader(f)
        next(reader)  # skip header
        for row in reader:
            adc.append(int(row[0]))
            g.append(float(row[1]))
    return adc, g

# Load files
adc_h, g_h = load_csv("healthy.csv")
adc_f, g_f = load_csv("fault.csv")

# Time axis
fs = 10000.0
t_h = [i / fs for i in range(len(g_h))]
t_f = [i / fs for i in range(len(g_f))]

# Plot
T_START = 0.0
T_END   = 0.1    # 100 ms window → stretches peaks
plt.figure(figsize=(18, 4))
plt.plot(t_h, g_h, label="Healthy", linewidth=1)
plt.plot(t_f, g_f, label="Fault", linewidth=1)

plt.xlim(T_START, T_END)

plt.xlabel("Time (s)")
plt.ylabel("Acceleration (g)")
plt.title("ADXL335 Vibration Comparison")
plt.legend()
plt.grid(True)
plt.tight_layout()

# Save with timestam8
ts = datetime.now().strftime("%H-%M-%S")
plt.savefig(f"plot_time_{ts}.png", dpi=300)

plt.show()


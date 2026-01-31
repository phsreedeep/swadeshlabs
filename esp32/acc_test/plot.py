import pandas as pd
import matplotlib.pyplot as plt

def process_and_plot(file1, file2, start=0, end=1000):
    # Load datasets
    df1 = pd.read_csv(file1)
    df2 = pd.read_csv(file2)

    # Slice data for the specified 1000 sample window
    # Vibration: third from last (-3), Temperature: last (-1)
    vib_1 = df1.iloc[start:end, -3]
    temp_1 = df1.iloc[start:end, -1]
    
    vib_2 = df2.iloc[start:end, -3]
    temp_2 = df2.iloc[start:end, -1]

    # Vibration Plot (Zoomed)
    plt.figure(figsize=(12, 6))
    plt.plot(vib_1.values, label='Healthy', alpha=0.8, color='blue')
    plt.plot(vib_2.values, label='Faulty', alpha=0.8, color='red')
    plt.title(f'Vibration Comparison (Samples {start}-{end})')
    plt.xlabel('Sample Offset')
    plt.ylabel('Acceleration (g)')
    plt.legend()
    plt.grid(True)
    plt.savefig('vibration_zoomed.png')
    plt.close()

    # Temperature Plot (Zoomed)
    plt.figure(figsize=(12, 6))
    plt.plot(temp_1.values, label='Healthy', alpha=0.8, color='green')
    plt.plot(temp_2.values, label='Faulty', alpha=0.8, color='orange')
    plt.title(f'Temperature Comparison (Samples {start}-{end})')
    plt.xlabel('Sample Offset')
    plt.ylabel('Temperature (°C)')
    plt.legend()
    plt.grid(True)
    plt.savefig('temperature_zoomed.png')
    plt.close()

if __name__ == "__main__":
    process_and_plot('./healthy/esp32_21-53-55_batch_0.csv', './faulty/esp32_21-44-56_batch_0.csv', start=0, end=1000)

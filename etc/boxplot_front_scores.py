import os
import glob
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

# --- Config ---
column_index = 1  # zero-based index (1 = second column, e.g., 0 = size, 1 = sample score, 2 = eqtl etc)
data_folder = "."  # update this path
pattern = "*.fronts"
output_plot = "boxplot_over_generations.png"

# --- Collect data ---
records = []

for filepath in sorted(glob.glob(os.path.join(data_folder, pattern)), key=lambda f: int(os.path.basename(f).split('.')[0])):
    generation = int(os.path.basename(filepath).split('.')[0])
    with open(filepath) as f:
        for line in f:
            if line.startswith("#") or line.strip() == "":
                continue
            parts = line.strip().split()
            if len(parts) <= column_index:
                continue
            try:
                value = float(parts[column_index])
                records.append({"generation": generation, "value": value})
            except ValueError:
                continue

df = pd.DataFrame(records)

# --- Plot ---
plt.figure(figsize=(14, 6))
sns.boxplot(x="generation", y="value", data=df)
plt.title(f"Distribution of column {column_index} over generations")
plt.xlabel("Generation")
plt.ylabel("Metric value")
plt.xticks(rotation=45)
plt.tight_layout()
plt.savefig(output_plot)
plt.show()
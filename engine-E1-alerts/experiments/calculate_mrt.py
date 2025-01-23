import pandas as pd
import sys

if len(sys.argv) < 3:
    print("Error, run like: $>python calculate_mrt.py <trace.csv> <metrics.csv>")
    exit(1)

trace_csv = sys.argv[1]
metrics_csv = sys.argv[2]

trace = pd.read_csv(sys.argv[1])
metrics = pd.read_csv(sys.argv[2])

mean_response_time_ns = int(trace["responseTime"].mean())

metrics["mrt"] = mean_response_time_ns

# print(metrics)

metrics.to_csv(metrics_csv, index=False)

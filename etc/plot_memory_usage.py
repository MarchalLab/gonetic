import subprocess
import matplotlib.pyplot as plt
import argparse
import re

def extract_memory_usage(profile_file, binary):
    # Run the pprof tool and extract the heap memory or allocations from the profile file
    result = subprocess.run(['go', 'tool', 'pprof', '-text', binary, profile_file],
                            capture_output=True, text=True)
    return result.stdout


def parse_pprof_memory(pprof_output):
    data = []

    # Regular expression to capture memory usage and function names
    # This regex looks for memory usage followed by a function name
    pattern = r'(\d+\.?\d*)MB\s+\d+.\d+%\s+\d+.\d+%\s+\d+\.?\d*MB\s+(\d+.\d+%)\s+([^\n]+)'

    matches = re.findall(pattern, pprof_output)

    for match in matches:
        memory_usage_mb = float(match[0])  # Capture the memory usage in MB
        function_name = match[2]  # Capture the function name
        data.append((function_name, memory_usage_mb))

    return data


def extract_name(text):
    # Regular expression to match numbers after "heap_", "goroutine_", or "_allocs"
    match = re.search(r'heap_(\d+)|goroutine_(\d+)|allocs_(\d+)', text)

    if match:
        # Return the number matched, ignoring the non-matched parts
        return match.group(), int(match.group(1) or match.group(2) or match.group(3))
    return None


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("pprofs", nargs='+',
                        help='.pprof files to parse')
    parser.add_argument("-b", '--binary', required=True,
                        help='gonetic binary')
    parser.add_argument('-o', '--out', default='memory_usage.pdf',
                        help='plot output (preferably ending with .pdf)')
    args = parser.parse_args()

    memory_data = {}
    for pprof in args.pprofs:
        name, no = extract_name(pprof)
        data = extract_memory_usage(pprof, args.binary)
        for function_name, memory_usage in parse_pprof_memory(data):
            memory_data.setdefault(function_name,[]).append((no, memory_usage))

    plt.figure(figsize=(12, 8))

    for function_name, memory_data in memory_data.items():
        if len(memory_data) > 10:
            # Extract iterations and memory usage
            sorted_memory_data = sorted(memory_data, key=lambda x: x[0])
            iterations = [point[0] for point in sorted_memory_data]
            memory_usage = [point[1] for point in sorted_memory_data]

            # Plot each function's memory usage
            plt.plot(iterations, memory_usage, marker='o', label=function_name)

    # Adding labels and title
    plt.xlabel('Iteration')
    plt.ylabel('Memory Usage (MB)')
    plt.title('Memory Usage per Function Over Iterations')

    # Display a legend to differentiate the functions
    plt.legend(loc='upper left', bbox_to_anchor=(1, 1))

    # Show the plot
    plt.tight_layout()
    plt.savefig(args.out)

if __name__ == "__main__":
    main()

import sys
import re

def read_network(file_path):
    genes = set()
    with open(file_path, 'r') as f:
        for line in f:
            if line.startswith('#'):
                continue
            parts = line.strip().split('\t')
            if len(parts) == 3:
                genes.add(parts[0])
                genes.add(parts[1])
    return genes

def filter_input_file(network_genes, input_file):
    with open(input_file, 'r') as f:
        for line in f:
            for gene in network_genes:
                pattern = re.compile(rf'[\t\n ;\.]{gene}[\t\n ;\.]')
                if pattern.search(line):
                    print(line, end='')
                    #print(f"# {gene}")
                    break

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python script.py <input_file>")
        sys.exit(1)

    network_file = 'weighted.network'
    input_file = sys.argv[1]

    network_genes = read_network(network_file)
    filter_input_file(network_genes, input_file)

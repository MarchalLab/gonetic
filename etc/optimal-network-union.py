"""

Given a set of networks with sizes ranging from 1 to x, for each size there can be multiple networks of that size. We now want to select for each size a single network, such that the union of all these selected networks has a minimal number of nodes.

This is not really a well-known problem, but it is definitely related to the Set Cover, Minimum Hitting Set, and Minimum k-Union NP-hard problems.

Plan of action:
- Baseline performance could be median of random network selection.
- Start with greedy expansion from a given start network (e.g. largest size network with best EQTL score), see how bad that gets. This corresponds with a max depth 1 BFS in each step, perhaps this depth could be increased to go slightly deeper.
- Could also have an alternative approach where we simply pick for each size the best scoring network for a particular metric (e.g. EQTL score), and combine those.
- List which steps are the worst in increasing the union size, this could be acted upon by excluding these sizes or backtracking to these size and picking a different option.

"""

import glob
import os


# Function to read and parse network files, filtering duplicates
def read_network_files(base_path="size_*"):
    network_data = []
    unique_networks = {}

    for folder in glob.glob(base_path):
        size = int(folder.split('_')[1])
        if size not in unique_networks:
            unique_networks[size] = set()

        for file in glob.glob(os.path.join(folder, "*.network")):
            with open(file, 'r') as f:
                lines = f.readlines()
                scores = eval('[' + lines[0].strip().split('[')[1].replace(' ', ', '))
                edges = [line.strip() for line in lines[2:]]

                # Use frozenset to create a hashable set of edges
                edges_set = frozenset(edges)

                # parse the edges
                edges = [parse_edge(edge) for edge in edges]

                # gather the nodes
                nodes = set(edge[0] for edge in edges) | set(edge[1] for edge in edges)

                # Check for duplicates
                if edges_set not in unique_networks[size]:
                    while len(network_data) < size + 1:
                        network_data.append([])

                    # register this edge set
                    unique_networks[size].add(edges_set)

                    network_data[size].append({
                        # network size
                        'size': size,
                        # file path
                        'file': file,
                        # edge set
                        'edges': edges,
                        # node set
                        'nodes': nodes,
                        # 1 / size
                        'size_score': scores[0],
                        # eqtl in eqtl setting, mutation in qtl setting, etc.
                        'main_score': scores[1],
                        # mutation or expression score in eqtl setting
                        'secondary_score': scores[2] if len(scores) > 2 else 0,
                        # expression score in eqtl setting
                        'tertiary_score': scores[3] if len(scores) > 3 else 0,
                    })

    return network_data


def parse_edge(edge_string):
    parts = edge_string.split(';')
    edge = {
        'source': parts[0],
        'sink': parts[1],
        'type': parts[2],
        'directed': parts[3] == 'directed',
        'weight': float(parts[4]),
        'id': int(parts[5]),
    }
    return edge['source'], edge['sink']


def greedy_select_networks(networks, tag='nodes'):
    # Initialize with the best starting network
    initial_network = initial_best_network(networks)
    initial_size = initial_network['size']
    selected_networks = [initial_network]

    # create dicts for nodes and edges in the union
    node_union = add_to_union({}, initial_network['nodes'], initial_size)
    edge_union = add_to_union({}, initial_network['edges'], initial_size)

    # pick which size to optimize
    current_union = edge_union if tag == 'edges' else node_union

    print(initial_size, initial_network['size'], len(current_union), initial_network['main_score'],
          initial_network['file'])

    for size in range(initial_network['size'], 1, -1):
        best_network, min_increase = None, float('inf')

        for network in networks[size]:
            if network not in selected_networks:
                union_size = calculate_union_size(current_union, network[tag])
                increase = union_size - len(current_union)

                if increase < min_increase or (
                        increase == min_increase and network['main_score'] > best_network['main_score']):
                    best_network, min_increase = network, increase

        if best_network is None:
            continue

        # Add the best network for the current size
        selected_networks.append(best_network)

        prev_node_union_size = len(node_union)
        node_union = add_to_union(node_union, best_network['nodes'], size)
        node_change = len(node_union) - prev_node_union_size

        prev_edge_union_size = len(edge_union)
        edge_union = add_to_union(edge_union, best_network['edges'], size)
        edge_change = len(edge_union) - prev_edge_union_size

        print(best_network['size'],
              min_increase if min_increase > 0 else '-',
              node_change if node_change > 0 else '-',
              edge_change if edge_change > 0 else '-',
              best_network['main_score'],
              best_network['file'],
              )

    return selected_networks, node_union, edge_union


"""
networks is a list of lists, where networks[x] contains networks of size x
"""


def initial_best_network(networks):
    print(len(networks))
    # Select the initial network based on criteria (e.g., largest size, best main score)
    return max(networks[len(networks) - 1], key=lambda nw: (nw['secondary_score']))
    return max(networks[len(networks) - 1], key=lambda nw: (nw['main_score']))
    # return max(networks[len(networks) - 1], key=lambda nw: (nw['size'], nw['main_score']))


def add_to_union(union, extension_list, value):
    for elt in extension_list:
        union[elt] = value
    return union


def calculate_union_size(union, extension_list):
    count = len(union)
    for elt in extension_list:
        if elt not in union:
            count += 1
    return count


def write_ranked_nodes(file_path, node_union):
    inverse_node_union = []
    for node, value in node_union.items():
        while len(inverse_node_union) < value + 1:
            inverse_node_union.append([])
        inverse_node_union[value].append(node)
    with open(file_path, 'w') as f:
        for i, nodes in enumerate(inverse_node_union):
            for node in nodes:
                f.write(f'{i} {node}\n')


def write_subnetwork_js_file(js_file_path, node_union, edge_union, max_size):
    # open the JS file for writing
    with open(js_file_path, 'w') as js_file:
        js_file.write('graph = {\n')
        # nodes are written to the JS file in the following format: {id:"NODEID", genesOfInterest:[]}
        js_file.write('nodes: [\n')
        for node, value in node_union.items():
            js_file.write('{id: "' + node + '", genesOfInterest: []},\n')
        js_file.write('],\n')
        # edges are written to the JS file in the following format: {source:"SOURCEID", target:"TARGETID"}
        # {source:"SOURCE", target:"SINK", type:"pp", direction:"directed", max_cost:cost, evidence:""},
        js_file.write('links: [\n')
        for edge, value in edge_union.items():
            source, target = edge
            js_file.write(
                '{source: "' + source + '", target: "' + target + '", type: "pp", direction: "directed", max_cost: ' + str(
                    ((value - 1) * (8 - 0.4) / (max_size - 1)) + 0.4) + ', evidence: ""},\n')
        js_file.write('],\n')
        # conditions are written to the JS file in the following format: "CONDITION"
        js_file.write('conditions: [],\n')
        js_file.write('};')


# Main function to tie everything together
def main():
    networks = read_network_files()
    selected_networks, node_union, edge_union = greedy_select_networks(networks)
    print([nw['file'] for nw in selected_networks])
    print(len(node_union), node_union)
    print(len(edge_union), edge_union)
    write_ranked_nodes('../ranked_nodes.txt', node_union)
    write_subnetwork_js_file('../subnetwork.js', node_union, edge_union, initial_best_network(networks)['size'])


# Execute main function
if __name__ == "__main__":
    main()

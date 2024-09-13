import random
import math

def generate_random(seed, config):
    # Generate Data for random example
    random.seed(seed)
    bins = {}
    pods = {}
    # Pods
    index_pods = [i for i in range(0, random.randint(config['num_pods_lower'], config['num_pods_upper']) + 1)]
    pods["index"] = index_pods
    ram_pods = [random.randint(config['ram_pods_lower'], config['ram_pods_upper']) for _ in index_pods]
    pods["ram"] = ram_pods
    cpu_pods = [random.randint(config['cpu_pods_lower'], config['cpu_pods_upper']) for _ in index_pods]
    pods["cpu"] = cpu_pods
    where_pods = [0 for _ in index_pods] # On which node the pod is allocated, 0 if not allocated
    pods["where"] = where_pods
    priority_pods = [random.randint(config['priority_pods_lower'], config['priority_pods_upper']) for _ in index_pods]
    pods["priority"] = priority_pods
    label_pods = [[] for _ in index_pods]
    pods["label"] = label_pods
    # Affinity and anti-affinity label can only reduce the space of the problem, so for benchmarking purpose we are ignoring it
    affinity_pods = [[] for _ in range(0, len(pods["index"]))]
    pods["affinity"] = affinity_pods
    anti_affinity_pods = [[] for _ in range(0, len(pods["index"]))]
    pods["anti_affinity"] = anti_affinity_pods
    # Bins
    index_bins = [i for i in range(1, random.randint(config['num_bins_lower'], config['num_bins_upper'])+1)]
    bins["index"] = index_bins
    num_bins = len(index_bins)
    total_ram_pods = sum(ram_pods)
    total_cpu_pods = sum(cpu_pods)
    ram_bins = [math.ceil((total_ram_pods / config['ram_bins_ratio']) / num_bins) for _ in index_bins]
    bins["ram"] = ram_bins
    cpu_bins = [math.ceil((total_cpu_pods / config['cpu_bins_ratio']) / num_bins) for _ in index_bins]
    bins["cpu"] = cpu_bins
    label_bins = [[] for _ in range(0, len(bins["index"]))]
    bins["label"] = label_bins
    # Allocate as much pods as you can with heuristic (First-Fit Decreasing)
    sorted_pods = sorted(range(len(pods["ram"])), key=lambda k: pods["ram"][k], reverse=True)
    current_bin = [bins["index"][0], bins["ram"][0], bins["cpu"][0]] # 0: index, 1: remaining ram, 2: remaining cpu
    for i_pod in sorted_pods:
        if pods["ram"][i_pod] > current_bin[1] or pods["cpu"][i_pod] > current_bin[2]:
            if current_bin[0] != bins["index"][-1]:
                current_bin[0] = current_bin[0]+1
                current_bin[1] = bins["ram"][current_bin[0]-1]
                current_bin[2] = bins["cpu"][current_bin[0]-1]
            else:
                break
        pods["where"][i_pod] = current_bin[0] # Position
        current_bin[1] -= pods["ram"][i_pod]  # Update ram remaining in bin
        current_bin[2] -= pods["cpu"][i_pod]  # Update cpu remaining in bin
    return bins, pods

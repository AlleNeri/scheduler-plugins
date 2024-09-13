import lib.converter as conv
import lib.generate as gen
import copy

def get_new_pods(pods: dict, bins: dict, x: dict, o, solver: str) -> dict:
    '''Print the final pods position'''
    if solver != 'z3' and solver != 'ortools':
        raise Exception('solver must be `z3` or `ortools`')
    new_pods = copy.deepcopy(pods)
    model = {}
    if solver == 'z3':
        model = o.model()
    for pod_index, i in enumerate(pods["index"]):
        which_bin = 0
        for _, j in enumerate(bins["index"]):
            if solver == 'z3':
                if model[x[(i, j)]].as_long() == 1:
                    which_bin = j
                    break
            elif solver == 'ortools':
                if x[(i, j)].solution_value() == 1:
                    which_bin = j
                    break
        new_pods['where'][pod_index] = which_bin
    return new_pods


def print_pods(pods, filter):
    for i in filter:
        n0 = 10
        n1 = 15
        n2 = 20
        print(f"{'Pod ' + str(i):<{n0}} {'ram:' + str(pods['ram'][i]):<{n0}} {'cpu:' + str(pods['cpu'][i]):<{n0}} {'priority: ' + str(pods['priority'][i]):<{n1}} {'where: ' + str(pods['where'][i]):<{n1}} {'affinity: ' + str(pods['affinity'][i]):<{n2}} {'anti-affinity: ' + str(pods['anti_affinity'][i])}")

def print_ratio(bins, pods):
    '''Print the ratio for bins and pods'''
    total_ram_pods = 0
    total_cpu_pods = 0
    total_ram_bins = 0
    total_cpu_bins = 0
    for bin_index, _ in enumerate(bins["index"]):
        total_ram_bins += bins["ram"][bin_index]
        total_cpu_bins += bins["cpu"][bin_index]
    for pod_index, _ in enumerate(pods["index"]):
        total_ram_pods += pods["ram"][pod_index]
        total_cpu_pods += pods["cpu"][pod_index]
    print(f"The ratio pods / bins for ram is {total_ram_pods / total_ram_bins}")
    print(f"The ratio pods / bins for cpu is {total_cpu_pods / total_cpu_bins}")

def print_initial_situation(bins, pods):
    '''Print the initial Bin situation'''
    padding = 30
    print("┌-- Initial Bin situation --┐")
    for bin_index, j in enumerate(bins["index"]):
        ram_bin_max = bins["ram"][bin_index]
        cpu_bin_max = bins["cpu"][bin_index]
        ram_bin_remain = ram_bin_max
        cpu_bin_remain = cpu_bin_max
        pods_here = []
        for pod_index, i in enumerate(pods["index"]):
            if pods["where"][pod_index] == j:
                ram_bin_remain -= pods["ram"][pod_index]
                cpu_bin_remain -= pods["cpu"][pod_index]
                pods_here.append(i)
        bin_print = f"Bin {j}: {pods_here}"
        remaining_print = f"label: {bins['label'][bin_index]}, remaining ram: {ram_bin_remain}/{ram_bin_max}, remaining cpu: {cpu_bin_remain}/{cpu_bin_max}"
        print(f"{bin_print : <{padding}}{remaining_print : >{padding}}")
    pods_here = []
    for pod_index, i in enumerate(pods["index"]):
        if pods["where"][pod_index] == 0:
            pods_here.append(i)
    print(f"Un-bin: {pods_here}")
    print(f"{' ' * int(padding/10)}{'-' * int(padding)}{' ' * int(padding/10)}")
    print_pods(pods, pods['index'])
    print("└-- Initial Bin situation --┘")

def generate_example(config_for_random: dict, kind: str = 'complex', seed = None, quiet = False, csv_path: str = ""):
    if not quiet:
        print(f"Generating example of kind: {kind} and seed: {seed}")
    bins = {}
    pods = {}
    if kind == 'complex':
        bins, pods = conv.csv_to_internal("benchmark/complex.csv")
    elif kind == 'simple':
        bins, pods = conv.csv_to_internal("benchmark/simple.csv")
    elif kind == 'random42':
        bins, pods = conv.csv_to_internal("benchmark/random_42.csv")
    elif kind == 'random':
        bins, pods = gen.generate_random(seed, config_for_random)
    elif kind == 'custom':
        bins, pods = conv.csv_to_internal(csv_path)
    else:
        raise Exception(f"Not a possible kind of example, recived `{kind}`")
    return bins, pods

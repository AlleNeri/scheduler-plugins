from ortools.linear_solver import pywraplp

def init_vars_constraints(bins, pods):
    '''Initialize Variable and Constraits'''
    o: pywraplp.Solver = pywraplp.Solver.CreateSolver("SAT")
    if not o:
        assert False
    # x_ij = 1 if item i packed in bin j
    x = {}
    for i in pods["index"]:
        for bin_index, j in enumerate(bins["index"]):
            # x_ij in {0, 1}
            x[(i, j)] = o.IntVar(0, 1, f"x_({i}, {j})")
            # Use Label for anti_affinity, force the variable to be 0
            for l in pods["affinity"][i]:
                if l not in bins["label"][bin_index]:
                    o.Add(x[(i, j)] <= 0)
            # Use Label for affinity, force the variable to be 0
            for l in pods["anti_affinity"][i]:
                if l in bins["label"][bin_index]:
                    o.Add(x[(i, j)] <= 0)
    # The amount packed in each bin cannot exceed its capacity.
    for j_index, j in enumerate(bins["index"]):
        o.Add(sum(x[(i, j)] * pods["ram"][i] for i in pods["index"]) <= bins["ram"][j_index])
        o.Add(sum(x[(i, j)] * pods["cpu"][i] for i in pods["index"]) <= bins["cpu"][j_index])
    # Each pod assigned at most to one bin
    for i in pods["index"]:
        o.Add(sum(x[(i, j)] for j in bins["index"]) <= 1)
    # Redundant Costraint for equivalent pods
    # If two pods are equivalent we impose and "order" so that the one with the lowest index must be placed first
    for p1 in pods["index"]:
        for p2 in pods["index"][p1+1:]:
            if pods["ram"][p1] == pods["ram"][p2] and pods["cpu"][p1] == pods["cpu"][p2] and pods["priority"][p1] == pods["priority"][p2] and pods["affinity"][p1] == pods["affinity"][p2] and pods["anti_affinity"][p1] == pods["anti_affinity"][p2]:
                for b1 in bins['index']:
                    for b2 in bins['index'][b1+1:]:
                        o.Add(x[(p1, b1)] <= x[(p2, b2)])
    # Redundant Costraint for equivalent bins
    # If two bins are equivalent both for cpu and ram we impose that the cpu and ram allocated to the one with the lowest index is less or equal than the one on the other side
    for j in bins["index"]:
        for j2 in bins["index"][j:]:
            bin_index = j-1
            bin_index2 = j2-1
            if bins["ram"][bin_index] == bins["ram"][bin_index2] and bins["cpu"][bin_index] == bins["cpu"][bin_index2] and bins["label"][bin_index] == bins["label"][bin_index2]:
                amount_packed_j = sum([ x[(i, j)] * pods["ram"][i] for i in pods["index"] ])
                amount_packed_j2 = sum([ x[(i, j2)] * pods["ram"][i] for i in pods["index"] ])
                o.Add(amount_packed_j <= amount_packed_j2)
    return o, x

def add_new_constraints(o, x, bins, pods_indexes):
    '''Add new constraints for pods which indexes are in pod_indexe'''
    for i in pods_indexes:
        # Pods must be allocated in exactly one bin
        o.Add(sum(x[(i, j)] for j in bins["index"]) == 1)

def quiet_print_bin_situation(bins, pods, x) -> tuple[int, int, int]:
    '''Placeholder for counting moved and removed pods without printing anything (apart from timeout change)'''
    moved_count = 0
    removed_count = 0
    added_count = 0
    pods_before = [pods['where'][pod_index] for pod_index, _ in enumerate(pods['index'])]
    pods_after = [0 for _, _ in enumerate(pods['index'])]
    for bin_index, j in enumerate(bins["index"]):
        ram_bin_max = bins["ram"][bin_index]
        cpu_bin_max = bins["cpu"][bin_index]
        ram_bin_remain = ram_bin_max
        cpu_bin_remain = cpu_bin_max
        pods_here = []
        for pod_index, i in enumerate(pods["index"]):
            if x[(i, j)].solution_value() == 1:
                ram_bin_remain -= pods["ram"][pod_index]
                cpu_bin_remain -= pods["cpu"][pod_index]
                pods_here.append(i)
                pods_after[pod_index] = j
                if pods["where"][pod_index] != j and pods["where"][pod_index] != 0:
                    moved_count += 1
    for i, before in enumerate(pods_before):
        after = pods_after[i]
        if before != 0 and after == 0:
            removed_count += 1
        elif before == 0 and after != 0:
            added_count += 1
    return moved_count, removed_count, added_count

def print_bin_situation(bins, pods, x) -> tuple[int, int, int]:
    '''Print the current Bin situation'''
    padding = 30
    print("┌-- Current Bin situation --┐")
    moved_count = 0
    removed_count = 0
    added_count = 0
    pods_before = [pods['where'][pod_index] for pod_index, _ in enumerate(pods['index'])]
    pods_after = [0 for _, _ in enumerate(pods['index'])]
    for bin_index, j in enumerate(bins["index"]):
        ram_bin_max = bins["ram"][bin_index]
        cpu_bin_max = bins["cpu"][bin_index]
        ram_bin_remain = ram_bin_max
        cpu_bin_remain = cpu_bin_max
        pods_here = []
        for pod_index, i in enumerate(pods["index"]):
            if x[(i, j)].solution_value() == 1:
                ram_bin_remain -= pods["ram"][pod_index]
                cpu_bin_remain -= pods["cpu"][pod_index]
                pods_here.append(i)
                pods_after[pod_index] = j
                if pods["where"][pod_index] != j and pods["where"][pod_index] != 0:
                    moved_count += 1
        bin_print = f"Bin {j}: {pods_here}"
        remaining_print = f"label: {bins['label'][bin_index]}, remaining ram: {ram_bin_remain}/{ram_bin_max}, remaining cpu: {cpu_bin_remain}/{cpu_bin_max}"
        print(f"{bin_print : <{padding}}{remaining_print : >{padding}}")
    print(f"Moved count = {moved_count}")
    for i, before in enumerate(pods_before):
        after = pods_after[i]
        if before != 0 and after == 0:
            removed_count += 1
        elif before == 0 and after != 0:
            added_count += 1
    print(f"Removed count = {removed_count}")
    print(f"Added count = {added_count}")
    print("└-- Current Bin situation --┘")
    return moved_count, removed_count, added_count

def pretty_print_check_code(code: int):
    if code == 0:
        print(f"{code = }, OPTIMAL")
    elif code == 1:
        print(f"{code = }, FEASIBLE")
    elif code == 2:
        print(f"{code = }, INFEASIBLE")
    elif code == 3:
        print(f"{code = }, UNBOUNDED")
    elif code == 4:
        print(f"{code = }, ABNORMAL")
    elif code == 5:
        print(f"{code = }, MODEL_INVALID")
    elif code == 6:
        print(f"{code = }, NOT_SOLVED")

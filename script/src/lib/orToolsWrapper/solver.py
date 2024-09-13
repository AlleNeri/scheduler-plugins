from ortools.linear_solver import pywraplp
from lib.orToolsWrapper.optimization import *
from lib.common import *
import time

def main(kind, seed, timeout, verbose, config_for_random, quiet, csv_path) -> tuple[bool, int, int, int, tuple[dict, dict], dict]:
    STARTING_TIME = time.time()
    is_optimal = True
    moved_count = 0
    removed_count = 0
    added_count = 0
    # Generate Raw Data
    bins: dict[str, list[int]] = {}
    pods: dict[str, list[int]] = {}
    bins, pods = generate_example(config_for_random, kind, seed, quiet, csv_path)
    # Print initial situation
    if not quiet:
        print_initial_situation(bins, pods)
    if verbose and not quiet:
        print_ratio(bins, pods)
    # Handle priority order
    sorted_priority_pods: list[int] = sorted(set(pods["priority"]), reverse=True)
    list_inserted_pods: list[int] = list()
    x = {}
    o, x = init_vars_constraints(bins, pods)
    o.SetTimeLimit(timeout)
    # Check if pods are all allocated, if so return
    # This is done only for benckmarks, should not be a case in real world scenario
    if all(pods['where']):
        return is_optimal, moved_count, removed_count, added_count, ({}, {}), {}
    # First fix highest priority (should be kube-system or equivalent)
    if verbose and not quiet:
        print(f"Adding constraint for highest priority (should be kube-system)")
    filtered_pods_index = [i for i in pods["index"] if pods["priority"][i] == sorted_priority_pods[0]]
    list_inserted_pods += filtered_pods_index
    if verbose and not quiet:
        print(f"Filtered pods index: {filtered_pods_index}")
        print_pods(pods, filtered_pods_index)
    add_new_constraints(o, x, bins, filtered_pods_index)
    # Minimize movement
    obj = sum(1 - x[(i, pods["where"][i])] for i in filtered_pods_index if pods["where"][i] != 0)
    o.Minimize(obj)
    # Update time remaining for check
    remaining_timeout = timeout - int((time.time() - STARTING_TIME) * 1000)
    if remaining_timeout < 0:
        raise Exception("Unable to finish")
    o.SetTimeLimit(remaining_timeout)
    last_check = o.Solve()
    if verbose and not quiet:
        print("last_check", last_check)
        pretty_print_check_code(last_check)
    if not (last_check == pywraplp.Solver.OPTIMAL or last_check == pywraplp.Solver.FEASIBLE):
        raise Exception("Unable to finish")
    # Get obj value
    value_found = sum(1 - x[(i, pods["where"][i])].solution_value() for i in filtered_pods_index if pods["where"][i] != 0)
    if verbose and not quiet:
        print_bin_situation(bins, pods, x)
    # Add constraint regarding allocation
    o.Add(value_found >= obj)
    # Update time remaining for check
    remaining_timeout = timeout - int((time.time() - STARTING_TIME) * 1000)
    if remaining_timeout < 0:
        raise Exception("Unable to finish")
    o.SetTimeLimit(remaining_timeout)
    last_check = o.Solve()
    if last_check != pywraplp.Solver.OPTIMAL:
        raise Exception("Unable to finish")
    for p in sorted_priority_pods[1:]:
        filtered_pods_index = [i for i in pods["index"] if pods["priority"][i] == p]
        if verbose and not quiet:
            print(f"Analyzing priority {p}")
            print(f"Filtered pods index: {filtered_pods_index}")
            print_pods(pods, filtered_pods_index)
        # Maximize pod allocated for this priority
        obj = sum(x[(i, j)] for i in pods['index'] if pods['priority'][i] == p for j in bins['index'])
        o.Maximize(obj)
        # Update time remaining for check
        remaining_timeout = timeout - int((time.time() - STARTING_TIME) * 1000)
        if remaining_timeout < 0:
            raise Exception("Unable to finish")
        o.SetTimeLimit(remaining_timeout)
        last_check = o.Solve()
        if verbose and not quiet:
            print("last_check", last_check)
            pretty_print_check_code(last_check)
        if not (last_check == pywraplp.Solver.OPTIMAL or last_check == pywraplp.Solver.FEASIBLE):
            raise Exception("Unable to finish")
        # Get obj value
        value_found = sum(x[(i, j)].solution_value() for i in pods['index'] if pods['priority'][i] == p for j in bins['index'])
        # Check if overall problem is optimal
        if value_found < len(filtered_pods_index):
            is_optimal = False
        # Add constraint regarding allocation
        o.Add(value_found <= obj)

        # Minimize movement
        obj = sum(1 - x[(i, pods["where"][i])] for i in filtered_pods_index if pods["where"][i] != 0)
        o.Minimize(obj)
        # Update time remaining for check
        remaining_timeout = timeout - int((time.time() - STARTING_TIME) * 1000)
        if remaining_timeout < 0:
            raise Exception("Unable to finish")
        o.SetTimeLimit(remaining_timeout)
        last_check = o.Solve()
        if verbose and not quiet:
            print("last_check", last_check)
            pretty_print_check_code(last_check)
        if not (last_check == pywraplp.Solver.OPTIMAL or last_check == pywraplp.Solver.FEASIBLE):
            raise Exception("Unable to finish")
        # Get obj value
        value_found = sum(1 - x[(i, pods["where"][i])].solution_value() for i in filtered_pods_index if pods["where"][i] != 0)
        if verbose and not quiet:
            print_bin_situation(bins, pods, x)
        # Add constraint regarding allocation
        o.Add(value_found >= obj)
    # Final Check
    # Update time remaining for check
    remaining_timeout = timeout - int((time.time() - STARTING_TIME) * 1000)
    if remaining_timeout < 0:
        raise Exception("Unable to finish")
    o.SetTimeLimit(remaining_timeout)
    last_check = o.Solve()
    if verbose and not quiet:
        print("last_check", last_check)
        pretty_print_check_code(last_check)
    if not (last_check == pywraplp.Solver.OPTIMAL or last_check == pywraplp.Solver.FEASIBLE):
        raise Exception("Unable to finish")
    if quiet:
        moved_count, removed_count, added_count = quiet_print_bin_situation(bins, pods, x)
    else:
        moved_count, removed_count, added_count = print_bin_situation(bins, pods, x)
    new_pods = get_new_pods(pods, bins, x, o, 'ortools')
    return is_optimal, moved_count, removed_count, added_count, (pods, new_pods), bins

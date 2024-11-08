import argparse
import json
# import lib.z3Wrapper.solver as z3Solver

import lib.orToolsWrapper.solver as orToolsSolver

parser = argparse.ArgumentParser()
parser.add_argument("--kind", type=str, help="Kind of example (simple|complex|random|custom), default custom", default="custom")
parser.add_argument("--seed", type=int, help="Seed for the random example, default None", default=None)
parser.add_argument("--timeout", type=int, help="Timeout of the solver in ms, default 1000", default=1000)
parser.add_argument("--solver", type=str, help="select solver (ortools), default ortools", default="ortools")
parser.add_argument("--verbose", help="increase output verbosity", action="store_true")
parser.add_argument("--config-path", help="path for config file for random generation", default="benchmark/config.json")
parser.add_argument("--csv-path", help="path for csv file of description of cluster", default="")
parser.add_argument("--quiet", help="make program quiet", action="store_true")
parser.add_argument("--json", help="output json as last line of stdout", action="store_true")

def is_better_situation(pods_before, pods_after) -> bool:
    # Get all priority
    priority_set = set()
    for p in pods_before['priority']:
        priority_set.add(p)
    priority_list = sorted(list(priority_set), reverse = True)
    # Check if it's better for each priority, starting from the highest
    for p in priority_list:
        index_to_check = []
        for i in pods_before['index']:
            if pods_before['priority'][i] == p:
                index_to_check.append(i)
        count_before = 0
        count_after = 0
        for i in index_to_check:
            if pods_before['where'][i] != 0:
                count_before += 1
            if pods_after['where'][i] != 0:
                count_after += 1
        if count_after > count_before:
            return True
        elif count_after < count_before:
            return False
    return False

if  __name__ == '__main__':
    args = parser.parse_args()
    assert args.solver in ['ortools']

    config_for_random = dict()
    if args.kind == 'random':
        with open(args.config_path, "r") as file:
            config_for_random = json.load(file)

    moved_count = 0
    removed_count = 0
    added_count = 0
    new_pods = {}

    if args.solver == 'ortools':
        try:
            is_optimal, moved_count, removed_count, added_count, (old_pods, new_pods), _ = orToolsSolver.main(args.kind, args.seed, args.timeout, args.verbose, config_for_random, args.quiet, args.csv_path)
            if args.verbose:
                print(f"is_optimal: {is_optimal}")
                print(f"moved_count: {moved_count}")
                print(f"removed_count: {removed_count}")
                print(f"{added_count = }")
        except Exception as err:
            print(f"Unexpected `{err}` exception")
            exit(1)
    else:
        raise Exception(f"Not a possible kind of solver, recived `{args.solver}`")

    if not is_better_situation(old_pods, new_pods):
        print(f"The solution found by the solver is worse or equal to the one already present on the cluster")
        exit(1)

    if args.verbose:
        print("New Pods Json:")

    if args.json:
        json_result = {}
        json_result['old_pods'] = old_pods
        json_result['new_pods'] = new_pods
        json_result['moved_count'] = moved_count
        json_result['removed_count'] = removed_count
        print(json.dumps(json_result))


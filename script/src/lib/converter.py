import csv
import ast

def csv_to_internal(pathToCsv: str) -> tuple[dict, dict]:
    bins = {}
    bins["index"] = []
    bins["ram"] = []
    bins["cpu"] = []
    bins["label"] = []
    pods = {}
    pods["index"] = []
    pods["ram"] = []
    pods["cpu"] = []
    pods["where"] = []
    pods["priority"] = []
    pods["affinity"] = []
    pods["anti_affinity"] = []
    pods["namespace"] = []
    with open(pathToCsv, newline='') as csvfile:
        spamreader = csv.reader(csvfile, delimiter=',', quotechar='|', escapechar="\\")
        headers = ["type",
                   "index", "ram", "cpu",
                   "label",
                   "where", "priority",
                   "affinity", "anti_affinity", "namespace"]
        if next(spamreader) != headers:
            raise Exception("Error in CSV file (You could be missing the title)")
        for row in spamreader:
            if row[0] == "bin":
                bins["index"].append(int(row[1]))
                bins["ram"].append(int(row[2]))
                bins["cpu"].append(int(row[3]))
                bins["label"].append(ast.literal_eval(row[4]))
            elif row[0] == "pod":
                pods["index"].append(int(row[1]))
                pods["ram"].append(int(row[2]))
                pods["cpu"].append(int(row[3]))
                pods["where"].append(int(row[5]))
                pods["priority"].append(int(row[6]))
                pods["affinity"].append(ast.literal_eval(row[7]))
                pods["anti_affinity"].append(ast.literal_eval(row[8]))
                pods["namespace"].append(row[9])
            else:
                raise Exception("Error in CSV file")
    return bins, pods

def internal_to_csv(bins, pods, pathToCsv: str):
    headers = ["type",
               "index", "ram", "cpu",
               "label",
               "where", "priority",
               "affinity", "anti_affinity", "namespace"]
    with open(pathToCsv, 'w', newline='') as csvfile:
        spamwriter = csv.writer(csvfile, delimiter=',', quotechar='|', escapechar='\\')
        spamwriter.writerow(headers)
        for b, _ in enumerate(bins["index"]):
            spamwriter.writerow(["bin",
                                 bins["index"][b], bins["ram"][b], bins["cpu"][b],
                                 bins["label"][b],
                                 "", "",
                                 "", ""])
        for p, _ in enumerate(pods["index"]):
            spamwriter.writerow(["pod",
                                 pods["index"][p], pods["ram"][p], pods["cpu"][p],
                                 "",
                                 pods["where"][p], pods["priority"][p],
                                 pods["affinity"][p], pods["anti_affinity"][p], pods["namespace"][p]])
    return

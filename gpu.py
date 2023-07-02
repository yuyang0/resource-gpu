import sys
import os
import subprocess
import json
import argparse


empty_node_usage_cmd = 'empty-node-usage'
get_node_info_cmd = 'get-node-info'
empty_node_capacity_cmd = 'empty-node-capacity'


def pretty_print_json(d):
    print(json.dumps(d, indent=2))


def empty_gpu_usage(cwd, nodename):
    cmd = [os.path.join(cwd, "gpu"), "fix-node-resource"]
    d = {
        "nodename": nodename,
    }
    in_str = json.dumps(d)
    out = subprocess.run(cmd, cwd=cwd, input=in_str, capture_output=True, text=True).stdout

    return json.loads(out)


def empty_gpu_capacity(cwd, nodename):
    cmd = [os.path.join(cwd, "gpu"), "set-node-resource-capacity"]
    d = {
        "nodename": nodename,
        "delta": False,
        "incr": False,
    }
    in_str = json.dumps(d)
    out = subprocess.run(cmd, cwd=cwd, input=in_str, capture_output=True, text=True).stdout

    return json.loads(out)


def get_node_info(cwd, nodename):
    cmd = [os.path.join(cwd, "gpu"), "get-node-resource-info"]
    d = {
        "nodename": nodename,
    }
    in_str = json.dumps(d)
    out = subprocess.run(cmd, cwd=cwd, input=in_str, capture_output=True, text=True).stdout

    return json.loads(out)


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("-cwd", "--cwd", help="dir of binary", default=os.getcwd())
    parser.add_argument("-nodename", "--nodename", help="nodename", required=True)
    subparsers = parser.add_subparsers(dest="cmd") 
    subparsers.add_parser(empty_node_usage_cmd)
    subparsers.add_parser(empty_node_capacity_cmd)
    subparsers.add_parser(get_node_info_cmd)

    args = parser.parse_args()
    return args


if __name__ == "__main__":
    args = parse_args()
    if args.cmd == empty_node_usage_cmd:
        ans = empty_gpu_usage(args.cwd, args.nodename)
        pretty_print_json(ans)
    elif args.cmd == empty_node_capacity_cmd:
        ans = empty_gpu_capacity(args.cwd, args.nodename)
        pretty_print_json(ans)
    elif args.cmd == get_node_info_cmd:
        ans = get_node_info(args.cwd, args.nodename)
        pretty_print_json(ans)
    else:
        print("unknown cmd")
        sys.exit(-1)
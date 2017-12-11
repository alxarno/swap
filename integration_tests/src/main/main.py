import argparse
import subprocess
import basic
import sys


def run(cmd):
    p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)


def createParser():
    parser = argparse.ArgumentParser()
    parser.add_argument('-ip', '--ip', default='192.168.0.1:8000', type=str)

    return parser


if __name__ == "__main__":
    parser = createParser()
    namespace = parser.parse_args(sys.argv[1:])
    if len(sys.argv) > 1:
        basic.set_prefix(namespace.ip)
        print("Start integrate test Spatium")
        cmd = "go run ..\..\src\main.go -test"
        run(cmd)
    else:
        print("Where are IP?")

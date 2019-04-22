import subprocess
import requests
import argparse


def main():
    parser = argparse.ArgumentParser(description="""
        This script forks all the relevant openshift repositories present in the payload into
        a namespace(user or organization given by the [namespace] argument so that they can
        be used by the CI operator to create images and subsequently a payload).""")
    
    parser.add_argument('--namespace', help="username/orgname where the repo needs to be forked.")
    args = parser.parse_args()
    print(args)

s
if __name__ == "__main__":
    main()
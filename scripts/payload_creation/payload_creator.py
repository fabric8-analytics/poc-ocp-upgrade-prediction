import subprocess
import argparse
from payload_creator.commit_sources import run_with_relese_info

def run_openshift_ci(args):
    # First push all sources to github.
    git_refs = commit_sources.run_with_relese_info(args)
    for remote, branch in git_refs.items():
        repo_name = remote.split('/')[-1]
        ref = "{}/{}@{}".format(args.git_namespace(), repo_name, branch)
        ci = subprocess.run([args.ci_operator, "--config", os.path.join(args.release_folder, "{}-release-4.1.yaml".format(repo_name)), "--namespace", args.ci_namespace, "--git-ref", ref])


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--ci-namespace", required=True, help="Name of namespace where these jobs should be run and payload created.")
    parser.add_argument("--release-folder", required=True, help="Path to openshift/release/ci-operator/config/openshift/ containing all the openshift CI job templates.")
    parser.add_argument("--ci-operator", required=True, help="CI operator binary.")
    parser.add_argument("--git-namespace", required=True, help="Namespace (user/org) whose forks will be used to push source code to.")
    parser.add_argument("--cluster-version", required=True, help="Version of OCP that will be used as base.")
    args = parser.parse_args()
    run_openshift_ci(args)
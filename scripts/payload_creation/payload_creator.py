import subprocess
import argparse
import os
import logging
from pathlib import Path
from payload_creation.commit_sources import run_with_release_info

logging.basicConfig(level=logging.INFO)
_logger = logging.getLogger(__name__)


def run_openshift_ci(args):
    # First push all sources to github.
    ci_config_path = Path(args.release_folder)
    git_refs = run_with_release_info(args)
    for remote, branch in git_refs.items():
        repo_name = remote.split("/")[-1]
        ref = "{}/{}@{}".format(args.git_namespace, repo_name, branch)
        _logger.info("Building image for: {}".format(ref))
        configpath = ci_config_path / "{}/openshift-{}-release-4.1.yaml".format(repo_name, repo_name)
        configpath = configpath.expanduser()
        if not os.path.exists(configpath):
            # There's no template for this!
            continue
        ci = subprocess.run(
            [
                args.ci_operator,
                "--config",
                 configpath._str,
                "--namespace",
                args.ci_namespace,
                "--git-ref",
                ref,
            ],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        _logger.debug("CI Operator output: {}\n Error: {}\n".format(ci.stdout, ci.stderr))


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--ci-namespace",
        required=True,
        help="Name of namespace where these jobs should be run and payload created.",
    )
    parser.add_argument(
        "--release-folder",
        required=True,
        help="Path to openshift/release/ci-operator/config/openshift/ containing all the openshift CI job templates.",
    )
    parser.add_argument("--ci-operator", required=True, help="CI operator binary.")
    parser.add_argument(
        "--git-namespace",
        required=True,
        help="Namespace (user/org) whose forks will be used to push source code to.",
    )
    parser.add_argument(
        "--cluster-version", required=True, help="Version of OCP that will be used as base."
    )
    parser.add_argument(
        "--destdir", required=True, help="The directory to use to create the payload."
    )
    parser.add_argument(
        "--no-verify", default=False, help="If you use git-secrets, need to set this to true"
    )
    parser.add_argument(
        "--cloned",
        default=False,
        help="If destdir already contains a clone of all the repositories.",
    )  # TODO: See if this argument is required.
    parser.add_argument(
        "--pushed", default=False, help="If all refs have already been pushed to Github.", type=bool
    )
    args = parser.parse_args()
    run_openshift_ci(args)

import requests
import logging
import argparse
import subprocess
import os
import json
from urllib.parse import urljoin, urlparse, ParseResult
from pathlib import Path
import sys

logging.basicConfig(level=logging.INFO)
_logger = logging.getLogger("github_forker")
GITHUB_API_URL = "https://api.github.com/"


def run_with_release_info(args):
    if os.environ.get("GH_TOKEN") == None:
        _logger.fatal("No GH_TOKEN available in environment.")
        import sys

        sys.exit(1)
    # First get all the repositories that are required to be forked.
    if args.get("digest", None) is None:
        process = subprocess.run(
            [
                "oc",
                "adm",
                "release",
                "info",
                "--commits=true",
                "registry.svc.ci.openshift.org/ocp/release:{}".format(args.cluster_version),
                "--output=json",
            ],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        cluster_payload = json.loads(process.stdout)
        service_arr = cluster_payload.get("references", {}).get("spec", {}).get("tags", {})
        clusterversion = cluster_payload.get("digest", "")
    else:
        clusterversion = args.digest
    git_refs = commit_sources(clusterversion, args.destdir, args.git_namespace, args.no_verify)
    return git_refs


def commit_sources(clusterversion, destdir, namespace, noverify):
    git_refs = {}
    path = Path(destdir) / clusterversion / "src/github.com/openshift/"
    for repopath in path.glob("*"):
        if os.stat(repopath / ".git") == None:  # Check for git repository.
            continue
        # First add the changes
        branch_name = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        _logger.debug("{}\n{}\n".format(branch_name.stdout, branch_name.stderr))
        # Create a branch
        branch_name_cleaned = str(branch_name.stdout.strip().decode("utf-8"))[:7]
        branch_create = subprocess.run(
            ["git", "checkout", "-b", branch_name_cleaned],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        _logger.debug("{}\n{}\n".format(branch_create.stdout, branch_create.stderr))
        # Now commit them
        add_code = subprocess.run(
            ["git", "add", "."], cwd=repopath, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        _logger.debug("{}\n{}\n".format(add_code.stdout, add_code.stderr))
        if noverify:
            pushcommand = [
                "git",
                "-c",
                "user.name=pocbot",
                "-c",
                "user.email=pocbot@localhost",
                "commit",
                "--no-verify",
                "-m",
                "tracer",
            ]
        else:
            [
                "git",
                "-c",
                "user.name=pocbot",
                "-c",
                "user.email=pocbot@localhost",
                "commit",
                "-m",
                "tracer",
            ]
        commit = subprocess.run(
            pushcommand, stdout=subprocess.PIPE, stderr=subprocess.PIPE, cwd=repopath
        )
        _logger.debug("{}\n{}\n".format(commit.stdout, commit.stderr))
        origin_url = subprocess.run(
            ["git", "config", "--get", "remote.origin.url"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        _logger.debug("{}\n{}\n".format(origin_url.stdout, origin_url.stderr))
        remote_url = urlparse(str(origin_url.stdout.strip().decode("utf-8")))
        # Set the name to our namespace
        path_parts = remote_url.path.split("/")
        path_parts[1] = namespace
        _logger.debug(path_parts)
        fork_url = ParseResult(
            remote_url.scheme,
            remote_url.netloc,
            "/".join(path_parts),
            remote_url.params,
            remote_url.query,
            remote_url.fragment,
        )
        remote_add = subprocess.run(
            ["git", "remote", "add", "ocp-poc-fork", fork_url.geturl()],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        _logger.debug("{}\n{}\n".format(remote_add.stdout, remote_add.stderr))
        push = subprocess.run(
            ["git", "push", "ocp-poc-fork", branch_name_cleaned],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        _logger.debug("{}\n{}\n".format(push.stdout, push.stderr))
        git_refs[remote_url.geturl()] = [branch_name_cleaned]


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="""
        This script commits all the patches to [namespace]'s forks of the relevant openshift
        repositories present in the payload into be used by the CI operator to create images and
        subsequently a payload)."""
    )
    parser.add_argument(
        "--git-namespace", help="username/orgname where the repo needs to be forked.", required=True
    )
    parser.add_argument(
        "--no-verify",
        required=True,
        help="Whether the no-verify option needs to be used (for people using git-secrets).",
    )
    parser.add_argument(
        "--org",
        type=bool,
        nargs="?",
        const=True,
        default=False,
        help="Whether target namespace is an org",
    )
    parser.add_argument(
        "--cluster-version",
        help="""The cluster version for which the services
        need to be validated.""",
        required=True,
    )

    parser.add_argument(
        "--destdir",
        default="/tmp/",
        help="Path, if other than /tmp where the payload's repositories have been cloned.",
    )
    args = parser.parse_args()
    run_with_release_info(args)

import requests
import logging
import argparse
import subprocess
import os
import json
import sys
from urllib.parse import urljoin, urlparse, ParseResult
from pathlib import Path
import sys

logging.basicConfig(level=logging.INFO)
_logger = logging.getLogger("github_forker")
GITHUB_API_URL = "https://api.github.com/"


def run_with_release_info(args):
    """Get the release payload information and return github refs for CI operator to use."""
    if os.environ.get("GH_TOKEN") == None:
        _logger.fatal("No GH_TOKEN available in environment.")
        sys.exit(1)
    # First get all the repositories that are required to be forked.
    if "digest" not in args:
        process = subprocess.run(
            [
                "oc",
                "adm",
                "release",
                "info",
                "--commits=true",
                "quay.io/openshift-release-dev/ocp-release:{}".format(args.cluster_version),
                "--output=json",
            ],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        cluster_payload = json.loads(process.stdout)
        service_arr = cluster_payload.get("references", {}).get("spec", {}).get("tags", {})
        service_name_origin_map = {
            service["name"]: service["annotations"]["io.openshift.build.source-location"].replace(
                "https://", ""
            )
            for service in service_arr
            if service["annotations"]["io.openshift.build.source-location"] != ""
        }
        clusterversion = cluster_payload.get("digest", "")
    else:
        clusterversion = args.digest
    if args.pushed:  # No need to commit sources
        return get_pushed_branches(clusterversion.split(':')[1][:7], args.destdir)
    else:
        return commit_sources(clusterversion.split(':')[1][:7], args.destdir, args.git_namespace, args.no_verify)


def commit_sources(cluster_service_dir, destdir, namespace, noverify):
    """Commit all patches to Github to our fork, for use by the CI opereator."""
    git_refs = {}
    path = Path(destdir) / cluster_service_dir / "src/github.com/openshift/"
    for repopath in path.expanduser().glob("*"):
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
        branch_name_cleaned = str(branch_name.stdout.strip().decode("utf-8"))[:7] + '-patch'
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
            ["git", "push", "-f", "ocp-poc-fork", branch_name_cleaned],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        _logger.debug("{}\n{}\n".format(push.stdout, push.stderr))
        git_refs[remote_url.geturl()] = branch_name_cleaned
    return git_refs


def get_pushed_branches(clusterdir, destdir):
    """If branches are already pushed, get them mapped to remotes."""
    git_refs = {}
    path = Path(destdir) / clusterdir / "src/github.com/openshift/"
    for repopath in path.expanduser().glob("*"):
        if os.stat((repopath / ".git")) == None:
            continue
        cmd = subprocess.run(
            ["git", "branch"], stderr=subprocess.PIPE, stdout=subprocess.PIPE, cwd=repopath
        )
        branch_name = cmd.stdout.strip().decode("utf-8").split("\n")[0].replace("* ", "")
        origin = subprocess.run(
            ["git", "remote", "get-url", "origin"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        fork = subprocess.run(
            ["git", "remote", "get-url", "ocp-poc-fork"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=repopath,
        )
        origin_remote = str(origin.stdout.strip().decode("utf-8"))
        fork_remote = str(fork.stdout.strip().decode("utf-8"))
        origin_remote = origin_remote.split("@")[1]
        git_refs["https://" + origin_remote] = branch_name
    return git_refs


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

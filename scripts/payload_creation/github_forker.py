import subprocess
import requests
import argparse
import os
import json
import subprocess
import logging
from urllib.parse import urljoin

logging.basicConfig(level=logging.DEBUG)
_logger = logging.getLogger("github_forker")
GITHUB_API_URL = "https://api.github.com/"


def main():
    parser = argparse.ArgumentParser(
        description="""
        This script forks all the relevant openshift repositories present in the payload into
        a namespace(user or organization given by the [namespace] argument so that they can
        be used by the CI operator to create images and subsequently a payload)."""
    )
    parser.add_argument(
        "--namespace",
        help="username/orgname where the repo needs to be forked.",
        required=True,
    )
    parser.add_argument(
        "--org",
        type=str2bool,
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
    if os.environ.get("GH_TOKEN") == None:
        os.Exit("No GH_TOKEN available in environment.")
    args = parser.parse_args()
    # First get all the repositories that are required to be forked.
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
    fork_repos(service_arr, args)


def fork_repos(service_arr, args):
    service_repo_map = {}
    for service_obj in service_arr:
        service_name = service_obj["name"]
        service_repo = service_obj["annotations"]["io.openshift.build.source-location"]
        service_repo_map[service_name] = service_repo
    # Now call the github API to get all repositories of our namespace
    response = requests.get(
        urljoin(GITHUB_API_URL, "/users/{}/repos".format(args.namespace)),
        headers={
            "Authorization": "token {}".format(os.environ.get("GH_TOKEN"))
        },
    )
    if response.status_code > 400:
        _logger.error(
            "Got an error from the Github API, status code: {}".format(
                response.status_code
            )
        )
    forked = set()
    for repo in response.json():
        forked.add(repo["name"])
    # Now fork all the repositories that are not in this namespace
    for service, repo in service_repo_map.items():
        reponame = repo.split("/")[-1]
        if reponame == "":
            _logger.warning("Skipping blank repo name for service {}".format(service))
        if reponame not in forked:
            # Send a request to the Github API to fork
            request = requests.Request(
                "POST",
                url=urljoin(
                    GITHUB_API_URL, "/repos/{}/forks".format('/'.join(repo.split('/')[-2:]))
                ),
                headers={
                    "Authorization": "token " + os.environ.get("GH_TOKEN"),
                    "Content-Type": "application/json",
                    "user-agent": "ocp-poc",
                },
                data=json.dumps({"organization": args.namespace}),
            ).prepare()
            response = requests.Session().send(request)
            if response.status_code > 400:
                _logger.error(
                    "Got an error from Github: {}, {}".format(
                        response.status_code, response.text
                    )
                )
        else:
            _logger.info("{} is already forked, not forking again.".format(repo))


def str2bool(v):
    if v.lower() in ("yes", "true", "t", "y", "1"):
        return True
    elif v.lower() in ("no", "false", "f", "n", "0"):
        return False
    else:
        raise argparse.ArgumentTypeError("Boolean value expected.")


def pretty_print_POST(req):
    """
    At this point it is completely built and ready
    to be fired; it is "prepared".

    However pay attention at the formatting used in 
    this function because it is programmed to be pretty 
    printed and may differ from the actual request.
    """
    print(
        "{}\n{}\n{}\n\n{}".format(
            "-----------START-----------",
            req.method + " " + req.url,
            "\n".join("{}: {}".format(k, v) for k, v in req.headers.items()),
            req.body,
        )
    )


if __name__ == "__main__":
    main()

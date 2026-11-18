#!/usr/bin/env python3
"""Record verifiable metadata for the pinned Jokruger module, not its source.

Run after ``go mod download -json github.com/jokruger/dec128@v1.4.0``
and ``go mod verify`` on a machine with access to the official Go proxy.
The report is evidence, *not* an authorship audit or release approval.
"""

import argparse
import hashlib
import json
from pathlib import Path
import sys

MODULE = "github.com/jokruger/dec128"
VERSION = "v1.4.0"
ROOT = Path(__file__).resolve().parents[2]


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def inspect(download, sums_path, snapshot):
    if download.get("Path") != MODULE or download.get("Version") != VERSION:
        raise ValueError("unexpected dependency path/version")
    if download.get("Error"):
        raise ValueError("Go module download failed: " + repr(download["Error"]))
    module_dir = Path(download.get("Dir") or "").resolve()
    if not module_dir.is_dir() or module_dir == ROOT or ROOT in module_dir.parents:
        raise ValueError("dependency is missing or resolved inside the repository")
    module_hash = download.get("Sum", "")
    mod_hash = download.get("GoModSum", "")
    if not module_hash.startswith("h1:") or not mod_hash.startswith("h1:"):
        raise ValueError("missing official module/download hashes")
    expected_lines = {f"{MODULE} {VERSION} {module_hash}",
                      f"{MODULE} {VERSION}/go.mod {mod_hash}"}
    actual_lines = set(sums_path.read_text().splitlines())
    if not expected_lines <= actual_lines:
        raise ValueError("download hashes do not match the committed go.sum contents")
    files = []
    for name in ("LICENSE", "NOTICE"):
        official = module_dir / name
        if not official.is_file():
            raise ValueError(f"{MODULE}@{VERSION} does not contain expected {name}")
        item = {"name": name, "sha256": sha256(official)}
        reference = snapshot / name
        if reference.is_file():
            item["snapshot_sha256"] = sha256(reference)
            item["snapshot_matches"] = item["sha256"] == item["snapshot_sha256"]
        else:
            item["snapshot_matches"] = None
        files.append(item)
    return {"module": MODULE, "version": VERSION, "module_sum": module_hash,
            "mod_sum": mod_hash, "licenses": files}


def markdown(result):
    lines = ["# Official Jokruger v1.4.0 metadata audit", "",
             "Recorded from Go Modules download metadata and the local Go module cache.",
             "Run `go mod verify` independently; this tool does not verify downloaded bytes.", "",
             f"- Module: `{result['module']}@{result['version']}`",
             f"- Module checksum from Go: `{result['module_sum']}`",
             f"- go.mod checksum from Go: `{result['mod_sum']}`", "",
             "| Official file | Official SHA-256 | Snapshot comparison |",
             "|---|---|---|"]
    for item in result["licenses"]:
        matches = item["snapshot_matches"]
        state = ("identical" if matches else "DIFFERENT") if matches is not None else "not supplied"
        lines.append(f"| `{item['name']}` | `{item['sha256']}` | {state} |")
    lines.extend(["", "These checks do not establish historical Badassitron authorship,",
                  "the adequacy of legal notices, or permission to publish. Review those separately.", ""])
    return "\n".join(lines)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--download-json", type=Path, required=True)
    p.add_argument("--go-sum", type=Path, default=ROOT / "go.sum")
    p.add_argument("--snapshot", type=Path, default=ROOT / "third_party_licenses/jokruger-dec128-snapshot")
    p.add_argument("--output", type=Path, required=True)
    args = p.parse_args()
    try:
        result = inspect(json.loads(args.download_json.read_text()), args.go_sum, args.snapshot)
        if not all(item["snapshot_matches"] is True for item in result["licenses"]):
            raise ValueError("official LICENSE/NOTICE does not match retained license snapshot")
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(markdown(result))
    except (ValueError, OSError, json.JSONDecodeError) as error:
        print("official module audit failed: " + str(error), file=sys.stderr)
        return 2
    print(f"PASS: {MODULE}@{VERSION} has matching download/go.sum checksums; review the license comparison")
    return 0


if __name__ == "__main__":
    sys.exit(main())

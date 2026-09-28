#!/usr/bin/env python3
"""Check the Registry contract for a local or downloaded release directory."""

import hashlib
import json
from pathlib import Path
import sys
import shutil
import zipfile


def verify(directory: Path) -> None:
    manifests = list(directory.glob("terraform-provider-polylane_*_manifest.json"))
    assert len(manifests) == 1, "Expected exactly one Registry manifest"
    manifest = manifests[0]
    prefix = manifest.name.removesuffix("_manifest.json")
    version = prefix.removeprefix("terraform-provider-polylane_")
    assert json.loads(manifest.read_text()) == {
        "version": 1,
        "metadata": {"protocol_versions": ["6.0"]},
    }, "Unexpected Registry protocol manifest"

    sums = {}
    for line in (directory / f"{prefix}_SHA256SUMS").read_text().splitlines():
        digest, filename = line.split(maxsplit=1)
        filename = filename.removeprefix("*")
        assert Path(filename).name == filename, "Checksum path must be a filename"
        assert filename not in sums, f"Duplicate checksum: {filename}"
        sums[filename] = digest
        assert hashlib.sha256((directory / filename).read_bytes()).hexdigest() == digest, filename

    archives = list(directory.glob(f"{prefix}_*.zip"))
    required = {"linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64", "windows_amd64"}
    platforms = set()
    for archive in archives:
        platform = archive.name.removeprefix(prefix + "_").removesuffix(".zip")
        platforms.add(platform)
        binary = f"terraform-provider-polylane_v{version}"
        if platform.startswith("windows_"):
            binary += ".exe"
        with zipfile.ZipFile(archive) as bundle:
            assert binary in bundle.namelist(), f"Missing provider binary in {archive.name}"
            assert bundle.getinfo(binary).file_size > 0, f"Empty binary in {archive.name}"
    assert required <= platforms, f"Missing release platforms: {required - platforms}"
    assert {manifest.name, *(a.name for a in archives)} == set(sums), "Checksums must cover every archive and manifest"
    print(f"Verified {len(archives)} platform archives, manifest, and checksums for {version}")


if __name__ == "__main__":
    directory = Path(sys.argv[1])
    if len(sys.argv) == 3 and sys.argv[2] == "--snapshot":
        # GoReleaser includes extra_files in checksums but only copies them
        # when uploading. Stage that same file for the offline snapshot check.
        version = json.loads((directory / "metadata.json").read_text())["version"]
        shutil.copyfile(
            "terraform-registry-manifest.json",
            directory / f"terraform-provider-polylane_{version}_manifest.json",
        )
    verify(directory)

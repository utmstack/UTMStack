"""Replay fabricated raw cases through separately built EventProcessor binaries.

Requires PyYAML. This does not build or deploy anything and uses only local file
writers. See filters/audits/deceptive-bytes.md for the tested version boundary.
"""
import argparse
import errno
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

import yaml


def records(path):
    # The playground writers can append adjacent JSON objects before newlines.
    content = path.read_text() if path.exists() else ""
    decoder = json.JSONDecoder()
    result, offset = [], 0
    while offset < len(content):
        while offset < len(content) and content[offset].isspace():
            offset += 1
        if offset < len(content):
            record, offset = decoder.raw_decode(content, offset)
            result.append(record)
    return result


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--playground", required=True, type=Path)
    parser.add_argument("--plugins", required=True, type=Path)
    args = parser.parse_args()
    fixture_dir = Path(__file__).resolve().parent
    root = fixture_dir.parents[3]
    os.umask(0o077)
    work = Path(tempfile.mkdtemp(prefix="db-pg-", dir="/tmp"))
    print(f"Local evidence directory: {work}", flush=True)
    for part in ("input", "output", "pipeline/filters", "rules", "plugins", "sockets", "geolocation"):
        (work / part).mkdir(parents=True, exist_ok=True)
    binaries = {"playground": args.playground.resolve()}
    for name in ("add", "grok", "delete", "sew", "kv", "trim", "cel", "saw"):
        source = (args.plugins / f"{name}.plugin").resolve()
        require(source.is_file(), f"Missing binary: {source}")
        try:
            os.link(source, work / "plugins" / source.name)
        except OSError as error:
            if error.errno != errno.EXDEV:
                raise
            shutil.copy2(source, work / "plugins" / source.name)
        binaries[name] = source
    config = {
        "tenants": [{"id": "00000000-0000-4000-8000-000000000001", "name": "fixture"}],
        "plugins": {
            "analysis": {"order": ["sew", "cel"]},
            "correlation": {"order": ["saw"]},
            "notification": {"order": []},
            # CEL initializes a client; this selected rule has no history.
            "org.opensearch": {"opensearch": "http://127.0.0.1:19200"},
        },
    }
    (work / "pipeline/config.yaml").write_text(yaml.safe_dump(config))
    shutil.copy2(fixture_dir / "patterns.yaml", work / "pipeline/patterns.yaml")
    filter_path = root / "filters/antivirus/deceptive-bytes.yml"
    shutil.copy2(filter_path, work / "pipeline/filters/deceptive-bytes.yaml")
    rule_path = root / "rules/antivirus/deceptive-bytes/living_off_the_land_detection.yml"
    rule = yaml.safe_load(rule_path.read_text())
    require(not rule.get("afterEvents") and not rule.get("correlation"), "This fixture requires a rule without history")
    rule["id"] = 9001
    (work / "rules/living-off-the-land.yaml").write_text(yaml.safe_dump([rule]))
    cases = json.loads((fixture_dir / "raw.json").read_text())
    for name, raw in cases.items():
        event = {
            "id": f"deceptive-bytes-{name}", "dataType": "deceptive-bytes",
            "dataSource": "synthetic-device", "@timestamp": "2026-09-23T12:00:00Z",
            "tenantId": config["tenants"][0]["id"], "raw": raw,
        }
        (work / "input" / f"{name}.json").write_text(json.dumps(event))
    manifest = {
        "provenance": "fabricated raw inputs; no customer data",
        "sourceHashes": {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
                         for p in (filter_path, rule_path, fixture_dir / "patterns.yaml", fixture_dir / "raw.json")},
        "binaries": {name: {
            "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            "buildInfo": subprocess.check_output(["go", "version", "-m", str(path)], text=True),
        } for name, path in binaries.items()},
    }
    (work / "manifest.json").write_text(json.dumps(manifest, indent=2))
    env = dict(os.environ, WORK_DIR=str(work), MODE="playground")
    with (work / "execution.log").open("w") as log:
        subprocess.run([str(binaries["playground"])], env=env, stdout=log,
                       stderr=subprocess.STDOUT, check=True, timeout=360)
    parsed = records(work / "output/resulting_log.json")
    events = {r.get("id"): r for r in parsed}
    require(len(parsed) == len(cases) == len(events), "Missing or duplicate output events")
    for name, raw in cases.items():
        event = events[f"deceptive-bytes-{name}"]
        require(event.get("raw") == raw, f"Raw input mismatch: {name}")
        require(not event.get("errors"), f"Parser errors: {name}")
    command = events["deceptive-bytes-command"]
    require(command.get("origin", {}).get("command") == "synthetic --flag", "Command not preserved")
    require(command.get("origin", {}).get("path") == "/tmp/fake.bin", "Path changed")
    require(events["deceptive-bytes-rest-data"].get("log", {}).get("sampleKey") == "sampleValue", "Present KV changed")
    blocked = events["deceptive-bytes-action-blocked"]
    require(blocked.get("actionResult") == "denied", "Blocked outcome not normalized")
    require(blocked.get("log", {}).get("action") == "blocked", "Vendor action lost")
    for name, event_type in (("rule-positive", "lolbin_trap"), ("rule-negative", "ordinary")):
        vendor = events[f"deceptive-bytes-{name}"].get("log", {})
        require(vendor.get("eventtype") == event_type, "KV event_type name mismatch")
        require(vendor.get("processname") == "cmd.exe", "KV process_name mismatch")
        require(vendor.get("deceptivetarget") == "decoy", "KV deceptive_target mismatch")
    alerts = records(work / "output/resulting_alert.json")
    require(len(alerts) == 1, f"Expected one local alert, got {len(alerts)}")
    alert = alerts[0]
    require(alert.get("name") == rule["name"] and not alert.get("errors"), "Unexpected alert or evaluation error")
    require([e.get("id") for e in alert.get("events", [])] == ["deceptive-bytes-rule-positive"], "Wrong alert event IDs")
    (work / "assertions.json").write_text(json.dumps({"passed": True, "events": len(parsed), "alerts": len(alerts)}))
    print("PASS: six raw events, zero parser errors, exactly one intended local alert")


if __name__ == "__main__":
    main()

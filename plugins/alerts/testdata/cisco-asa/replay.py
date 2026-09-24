"""Replay fabricated Cisco ASA raw lines through separately built EventProcessor binaries.

Requires PyYAML. This does not build or deploy anything and uses only local file writers.
It stages the current filter, the shared grok definitions (patterns.yaml), fabricated
geolocation data and the three Cisco ASA rules, runs the playground once, and checks every
event against expected.json and every local alert against its expected rule. Two of the
rules have history searches. Their OpenSearch address is a closed local port: no fixture
may reach a history search, and an attempt would fail and be reported. Every input is
fabricated. See filters/audits/cisco-asa.md.
"""
import argparse
import errno
import hashlib
import json
import os
from pathlib import Path
import shutil
import socket
import subprocess
import tempfile

import yaml

PLUGINS = ("add", "cast", "cel", "delete", "grok", "rename", "saw", "sew", "trim")
GEOLOCATION = "com.utmstack.geolocation.plugin"
TENANT = "00000000-0000-4000-8000-000000000001"
ENVELOPE_TIMESTAMP = "2026-09-23T14:00:00Z"
ENVELOPE = ("id", "timestamp", "deviceTime", "dataType", "dataSource", "tenantId", "tenantName", "raw", "errors")
LOG_FAILURES = ("failed to unmarshal", "failed to evaluate rule", "failed to execute correlation search",
                "plugin not found", "failed to compile", "failed to start plugin",
                "failed to convert log to event", "all retries failed", "panic")


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


def flatten(value, prefix=""):
    """Dotted paths of every leaf; an empty object is kept as a leaf."""
    out = {}
    if isinstance(value, dict) and (value or not prefix):
        for key, item in value.items():
            out.update(flatten(item, f"{prefix}.{key}" if prefix else key))
    else:
        out[prefix] = value
    return out


def fields(event):
    return flatten({k: v for k, v in event.items() if k not in ENVELOPE})


def closed_port():
    # A port nothing listens on: bind an ephemeral port, then release it.
    with socket.socket() as probe:
        probe.bind(("127.0.0.1", 0))
        return probe.getsockname()[1]


def place(source, target):
    target.parent.mkdir(parents=True, exist_ok=True)
    try:
        os.link(source, target)
    except OSError as error:
        if error.errno != errno.EXDEV:
            raise
        shutil.copy2(source, target)


def run(playground, plugins, geolocation_plugin):
    """Stage everything in a fresh private directory, run the playground, return its results."""
    fixture_dir = Path(__file__).resolve().parent
    root = fixture_dir.parents[3]
    os.umask(0o077)
    work = Path(tempfile.mkdtemp(prefix="asa-pg-", dir="/tmp"))
    print(f"Local evidence directory: {work}", flush=True)
    for part in ("input", "output", "pipeline/filters", "rules", "plugins", "sockets", "geolocation"):
        (work / part).mkdir(parents=True, exist_ok=True)
    binaries = {"playground": playground.resolve()}
    for name in PLUGINS:
        source = (plugins / f"{name}.plugin").resolve()
        require(source.is_file(), f"Missing binary: {source}")
        place(source, work / "plugins" / source.name)
        binaries[name] = source
    source = geolocation_plugin.resolve()
    require(source.is_file(), f"Missing binary: {source}")
    place(source, work / "plugins" / "utmstack" / GEOLOCATION)
    binaries["geolocation"] = source
    port = closed_port()
    config = {
        "tenants": [{"id": TENANT, "name": "fixture"}],
        "plugins": {
            "analysis": {"order": ["sew", "cel"]},
            "correlation": {"order": ["saw"]},
            "notification": {"order": []},
            # CEL builds a client at start-up; nothing listens here, so a history search would fail.
            "org.opensearch": {"opensearch": f"http://127.0.0.1:{port}"},
        },
    }
    (work / "pipeline/config.yaml").write_text(yaml.safe_dump(config))
    shutil.copy2(fixture_dir / "patterns.yaml", work / "pipeline/patterns.yaml")
    filter_path = root / "filters/cisco/asa.yml"
    shutil.copy2(filter_path, work / "pipeline/filters/asa.yaml")
    # Not named geolocation/: the repository ignores directories with that name.
    geo_paths = sorted((fixture_dir / "geolocation-data").glob("*.csv"))
    require(len(geo_paths) == 5, f"Expected 5 geolocation files, found {len(geo_paths)}")
    for path in geo_paths:
        shutil.copy2(path, work / "geolocation" / path.name)
    rule_paths = sorted((root / "rules/cisco/asa").glob("*.y*ml"))
    require(len(rule_paths) == 3, f"Expected 3 Cisco ASA rules, found {len(rule_paths)}")
    stems = {}
    for offset, path in enumerate(rule_paths):
        rule = yaml.safe_load(path.read_text())
        require(isinstance(rule, dict) and "id" not in rule, f"Unexpected rule shape: {path.name}")
        rule["id"] = 9001 + offset  # the playground loader needs unique non-zero ids
        stems[rule["name"]] = path.stem
        (work / "rules" / f"{rule['id']}-{path.stem}.yaml").write_text(yaml.safe_dump([rule]))
    cases = json.loads((fixture_dir / "raw.json").read_text())["cases"]
    inputs = {}
    for number, name in enumerate(sorted(cases)):
        event_id = f"cisco-asa-{name}"
        inputs[event_id] = (name, cases[name])
        event = {"id": event_id, "dataType": "firewall-cisco-asa", "dataSource": "fixture-asa",
                 "@timestamp": ENVELOPE_TIMESTAMP, "tenantId": TENANT, "raw": cases[name]}
        (work / "input" / f"{number:03d}.json").write_text(json.dumps(event))
    hashed = [filter_path, *rule_paths, *geo_paths] + [fixture_dir / n for n in ("patterns.yaml", "raw.json")]
    manifest = {
        "provenance": "fabricated raw inputs; no customer data",
        "historyBackend": f"none: http://127.0.0.1:{port} is closed",
        "sourceHashes": {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest() for p in hashed},
        "binaries": {name: {
            "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            "buildInfo": subprocess.check_output(["go", "version", "-m", str(path)], text=True),
        } for name, path in binaries.items()},
    }
    (work / "manifest.json").write_text(json.dumps(manifest, indent=2))
    env = dict(os.environ, WORK_DIR=str(work), MODE="playground")
    with (work / "execution.log").open("w") as log:
        subprocess.run([str(binaries["playground"])], env=env, stdout=log,
                       stderr=subprocess.STDOUT, check=True, timeout=900)
    log_text = (work / "execution.log").read_text()
    events = records(work / "output/resulting_log.json")
    alerts = records(work / "output/resulting_alert.json")
    return work, inputs, stems, events, alerts, log_text


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--playground", required=True, type=Path)
    parser.add_argument("--plugins", required=True, type=Path)
    parser.add_argument("--geolocation-plugin", required=True, type=Path,
                        help="com.utmstack.geolocation.plugin built from this checkout's plugins/geolocation")
    args = parser.parse_args()
    fixture_dir = Path(__file__).resolve().parent
    expected = json.loads((fixture_dir / "expected.json").read_text())["cases"]
    work, inputs, stems, events, alerts, log_text = run(args.playground, args.plugins, args.geolocation_plugin)
    require(set(expected) == {name for name, _ in inputs.values()}, "raw.json and expected.json disagree")
    for marker in LOG_FAILURES:
        require(marker not in log_text, f"Execution log reports: {marker}")

    by_id = {e.get("id"): e for e in events}
    require(len(events) == len(inputs) == len(by_id), f"Events {len(events)} for {len(inputs)} inputs")
    for event_id, (name, raw) in inputs.items():
        event, want = by_id[event_id], expected[name]
        require(event.get("raw") == raw, f"Raw input changed: {event_id}")
        require(not event.get("errors"), f"Parser errors: {event_id}: {len(event.get('errors') or [])}")
        require(("log" in event) == want["logObject"], f"{event_id}: log object present={'log' in event}")
        got = fields(event)
        differ = sorted(k for k in set(got) | set(want["fields"]) if got.get(k, "<absent>") != want["fields"].get(k, "<absent>"))
        require(not differ, f"{event_id}: fields differ: " +
                "; ".join(f"{k}={got.get(k, '<absent>')!r}, want {want['fields'].get(k, '<absent>')!r}" for k in differ[:5]))

    fired = {}
    for alert in alerts:
        require(not alert.get("errors") and not alert.get("name", "").startswith("Circuit Breaker"),
                f"Rule evaluation failure: {alert.get('name')}")
        require(alert.get("name") in stems, f"Unknown alert: {alert.get('name')}")
        ids = [e.get("id") for e in alert.get("events", [])]
        require(ids and ids[-1] in inputs, f"Unexpected alert events: {alert.get('name')} {ids}")
        fired.setdefault(ids[-1], []).append(stems[alert["name"]])
    for event_id, (name, _) in inputs.items():
        got = sorted(fired.get(event_id, []))
        require(got == sorted(expected[name]["alerts"]), f"{event_id}: alerts {got}, want {sorted(expected[name]['alerts'])}")
    result = {"passed": True, "events": len(events), "alerts": len(alerts)}
    (work / "assertions.json").write_text(json.dumps(result))
    print(f"PASS: {len(events)} raw events, zero parser errors, {len(alerts)} local alerts, each from its "
          f"intended rule, no Circuit Breaker and no history search attempted")


if __name__ == "__main__":
    main()

"""Replay fabricated SentinelOne raw lines through separately built EventProcessor binaries.

Requires PyYAML. This does not build or deploy anything and uses only local file
writers. It stages the current filter and all 19 SentinelOne rules, runs the playground
and checks every event against expected.json and every local alert against the expected
rule. With --endpoint-harness it also stages a test-only step that adds target.host to
marked copies of the endpoint-gated lines; that step stands in for the endpoint mapping
this filter does not have yet and is never shipped. See filters/audits/sentinel-one.md.
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

PLUGINS = ("add", "cel", "delete", "grok", "kv", "reformat", "rename", "saw", "sew", "trim")
TENANT = "00000000-0000-4000-8000-000000000001"
HARNESS_SOURCE = "endpoint-harness"
HARNESS_HOST = "endpoint-01.example.com"
HARNESS_STAGE = f"""# Test-only stage, never shipped. It stands in for the deferred endpoint mapping.
pipeline:
  - dataTypes:
      - antivirus-sentinel-one
    steps:
      - add:
          function: string
          params:
            key: target.host
            value: {HARNESS_HOST}
          where: 'equals("dataSource", "{HARNESS_SOURCE}")'
"""
LOG_FAILURES = ("failed to unmarshal", "failed to evaluate rule", "plugin not found",
                "failed to compile regexp", "failed to start plugin", "failed to parse time")


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


def get(value, path):
    for part in path.split("."):
        if not isinstance(value, dict) or part not in value:
            return None, False
        value = value[part]
    return value, True


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--playground", required=True, type=Path)
    parser.add_argument("--plugins", required=True, type=Path)
    parser.add_argument("--endpoint-harness", action="store_true")
    args = parser.parse_args()
    fixture_dir = Path(__file__).resolve().parent
    root = fixture_dir.parents[3]
    os.umask(0o077)
    work = Path(tempfile.mkdtemp(prefix="s1-pg-", dir="/tmp"))
    print(f"Local evidence directory: {work}", flush=True)
    for part in ("input", "output", "pipeline/filters", "rules", "plugins", "sockets", "geolocation"):
        (work / part).mkdir(parents=True, exist_ok=True)
    binaries = {"playground": args.playground.resolve()}
    for name in PLUGINS:
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
        "tenants": [{"id": TENANT, "name": "fixture"}],
        "plugins": {
            "analysis": {"order": ["sew", "cel"]},
            "correlation": {"order": ["saw"]},
            "notification": {"order": []},
            # CEL initializes a client; no SentinelOne rule has a history query.
            "org.opensearch": {"opensearch": "http://127.0.0.1:19200"},
        },
    }
    (work / "pipeline/config.yaml").write_text(yaml.safe_dump(config))
    shutil.copy2(fixture_dir / "patterns.yaml", work / "pipeline/patterns.yaml")
    filter_path = root / "filters/antivirus/sentinel-one.yml"
    shutil.copy2(filter_path, work / "pipeline/filters/sentinel-one.yaml")
    if args.endpoint_harness:
        (work / "pipeline/filters/zz-endpoint-harness.yaml").write_text(HARNESS_STAGE)
    rule_paths = sorted((root / "rules/antivirus/sentinel-one").glob("*.y*ml"))
    require(len(rule_paths) == 19, f"Expected 19 SentinelOne rules, found {len(rule_paths)}")
    stems = {}
    for offset, path in enumerate(rule_paths):
        rule = yaml.safe_load(path.read_text())
        require(isinstance(rule, dict) and "id" not in rule, f"Unexpected rule shape: {path.name}")
        require(not rule.get("afterEvents") and not rule.get("correlation"), f"{path.name} needs history")
        rule["id"] = 9001 + offset  # the playground loader needs unique non-zero ids
        stems[rule["name"]] = path.stem
        (work / "rules" / f"{rule['id']}-{path.stem}.yaml").write_text(yaml.safe_dump([rule]))
    cases = json.loads((fixture_dir / "raw.json").read_text())
    spec = json.loads((fixture_dir / "expected.json").read_text())
    expected = spec["cases"]
    require(set(cases) == set(expected), "raw.json and expected.json disagree")
    inputs = {}
    for name, raw in cases.items():
        inputs[f"sentinel-one-{name}"] = (name, "fixture-console", raw)
        if args.endpoint_harness and "endpointHarnessAlerts" in expected[name]:
            inputs[f"sentinel-one-{name}-endpoint"] = (name, HARNESS_SOURCE, raw)
    for number, (event_id, (name, source, raw)) in enumerate(sorted(inputs.items())):
        event = {"id": event_id, "dataType": "antivirus-sentinel-one", "dataSource": source,
                 "@timestamp": spec["envelopeTimestamp"], "tenantId": TENANT, "raw": raw}
        (work / "input" / f"{number:03d}.json").write_text(json.dumps(event))
    hashed = [filter_path, *rule_paths] + [fixture_dir / n for n in ("patterns.yaml", "raw.json", "expected.json")]
    manifest = {
        "provenance": "fabricated raw inputs; no customer data",
        "endpointHarness": args.endpoint_harness,
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
    for marker in LOG_FAILURES:
        require(marker not in log_text, f"Execution log reports: {marker}")

    parsed = records(work / "output/resulting_log.json")
    events = {r.get("id"): r for r in parsed}
    require(len(parsed) == len(inputs) == len(events), f"Events {len(parsed)} for {len(inputs)} inputs")
    for event_id, (name, source, raw) in inputs.items():
        event, want = events[event_id], expected[name]
        require(event.get("raw") == raw, f"Raw input changed: {event_id}")
        require(not event.get("errors"), f"Parser errors: {event_id}: {event.get('errors')}")
        log = event.get("log") or {}
        require(sorted(log) == want["logKeys"],
                f"{event_id}: log keys {sorted(set(log) ^ set(want['logKeys']))} differ")
        for key, value in want["log"].items():
            require(log.get(key) == value, f"{event_id}: log.{key}={log.get(key)!r}, want {value!r}")
        for path, value in want["fields"].items():
            require(get(event, path)[0] == value, f"{event_id}: {path}={get(event, path)[0]!r}, want {value!r}")
        for path in want["absent"]:
            require(not get(event, path)[1], f"{event_id}: {path} should be absent")
        host = get(event, "target.host")[0]
        require(host == (HARNESS_HOST if source == HARNESS_SOURCE else None), f"{event_id}: target.host={host!r}")

    alerts = records(work / "output/resulting_alert.json")
    fired = {}
    for alert in alerts:
        require(not alert.get("errors") and not alert.get("name", "").startswith("Circuit Breaker"),
                f"Rule evaluation failure: {alert.get('name')}")
        ids = [e.get("id") for e in alert.get("events", [])]
        require(len(ids) == 1 and ids[0] in inputs, f"Unexpected alert events: {alert.get('name')} {ids}")
        require(alert.get("name") in stems, f"Unknown alert: {alert.get('name')}")
        fired.setdefault(ids[0], []).append(stems[alert["name"]])
        if inputs[ids[0]][1] == HARNESS_SOURCE:
            require((alert.get("target") or {}).get("host") == HARNESS_HOST, f"{ids[0]}: alert target")
    for event_id, (name, source, _) in inputs.items():
        want = expected[name]["endpointHarnessAlerts"] if source == HARNESS_SOURCE else expected[name]["alerts"]
        got = sorted(fired.get(event_id, []))
        require(got == sorted(want), f"{event_id}: alerts {got}, want {sorted(want)}")
    result = {"passed": True, "events": len(parsed), "alerts": len(alerts), "endpointHarness": args.endpoint_harness}
    (work / "assertions.json").write_text(json.dumps(result))
    print(f"PASS: {len(parsed)} raw events, zero parser errors, {len(alerts)} local alerts, each from its intended rule")


if __name__ == "__main__":
    main()

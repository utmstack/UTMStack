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

# Rules without history searches, staged together; each must alert on exactly one case.
RULES = {
    "living_off_the_land_detection": "rule-positive",
    "nation_state_tactic_detection": "nation-state-positive",
    "privilege_escalation_bait_detection": "privilege-positive",
}
# Vendor keys that KV stores under log. Since go-sdk v1.1.35 the field-name sanitizer
# keeps underscores, so each key keeps its underscore; the six rules read these names.
KV_FIELDS = {
    "rule-positive": {"event_type": "lolbin_trap", "process_name": "cmd.exe", "deceptive_target": "decoy"},
    "rule-negative": {"event_type": "ordinary", "process_name": "cmd.exe", "deceptive_target": "decoy"},
    "theft-names": {"event_type": "decoy_accessed", "action": "file_copy", "decoy_sensitivity": "high",
                    "decoy_file": "finance-decoy.xlsx"},
    "lateral-names": {"event_type": "trap_triggered", "trap_type": "lateral_movement"},
    "nation-state-positive": {"event_type": "decoy_interaction", "threat_level": "critical",
                              "attack_sophistication": "advanced", "threat_score": "90", "apt_indicators": "true",
                              "custom_malware": "false", "advanced_ttps": "false", "targeted_decoys": "1",
                              "persistence_attempt": "false"},
    "privilege-positive": {"event_type": "bait_accessed", "bait_type": "privileged_account", "target_privilege": "admin"},
    "ransomware-names": {"event_type": "ransomware_behavior", "behavior_pattern": "mass_encryption",
                         "process": "example.exe", "source_ip": "192.0.2.10"},
}


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
    rule_paths, names = [], {}
    for offset, stem in enumerate(RULES):
        rule_path = root / f"rules/antivirus/deceptive-bytes/{stem}.yml"
        rule = yaml.safe_load(rule_path.read_text())
        require(not rule.get("afterEvents") and not rule.get("correlation"), f"{stem} needs history")
        rule["id"] = 9001 + offset  # the playground loader needs unique non-zero ids
        (work / f"rules/{stem}.yaml").write_text(yaml.safe_dump([rule]))
        rule_paths.append(rule_path)
        names[rule["name"]] = stem
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
                         for p in (filter_path, *rule_paths, fixture_dir / "patterns.yaml", fixture_dir / "raw.json")},
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
    for name, fields in KV_FIELDS.items():
        vendor = events[f"deceptive-bytes-{name}"].get("log", {})
        for key, value in fields.items():
            require(vendor.get(key) == value, f"{name}: log.{key}={vendor.get(key)!r}, want {value!r}")
            require(key.replace("_", "") == key or key.replace("_", "") not in vendor,
                    f"{name}: log.{key} was stored without its underscore")
    alerts = records(work / "output/resulting_alert.json")
    require(len(alerts) == len(RULES), f"Expected {len(RULES)} local alerts, got {len(alerts)}")
    fired = {}
    for alert in alerts:
        require(alert.get("name") in names and not alert.get("errors"), "Unexpected alert or evaluation error")
        fired[names[alert["name"]]] = [e.get("id") for e in alert.get("events", [])]
    for stem, case in RULES.items():
        require(fired.get(stem) == [f"deceptive-bytes-{case}"], f"{stem}: alert events {fired.get(stem)}")
    (work / "assertions.json").write_text(json.dumps({"passed": True, "events": len(parsed), "alerts": len(alerts)}))
    print(f"PASS: {len(parsed)} raw events, zero parser errors, every vendor key stored with its underscore, "
          f"{len(alerts)} local alerts, each from its intended rule")


if __name__ == "__main__":
    main()

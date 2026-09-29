#!/usr/bin/env python3
"""Inventory grouping paths from a fixed git revision, not the working tree.

Requires PyYAML. Writes JSON to stdout; does not connect to instances or modify git.
"""
import argparse
from collections import Counter
import io
import json
from pathlib import Path
import subprocess
import tarfile

import yaml


def git(*args):
    root = Path(__file__).resolve().parents[3]
    return subprocess.check_output(['git', '-C', str(root), *args])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('revision', help='Reviewed base commit/ref (recorded as a full SHA)')
    args = parser.parse_args()
    revision = git('rev-parse', '--verify', args.revision + '^{commit}').decode().strip()
    records = []
    rule_files = 0
    with tarfile.open(fileobj=io.BytesIO(git('archive', revision, 'rules'))) as archive:
        for member in sorted(archive.getmembers(), key=lambda item: item.name):
            if not member.isfile() or not member.name.endswith(('.yml', '.yaml')):
                continue
            rule_files += 1
            for document, rule in enumerate(yaml.safe_load_all(archive.extractfile(member)), 1):
                if not isinstance(rule, dict):
                    raise ValueError(f'{member.name}: expected one rule mapping per YAML document')
                for kind in ('groupBy', 'deduplicateBy'):
                    for field in rule.get(kind) or []:
                        if not isinstance(field, str):
                            raise ValueError(f'{member.name}: non-string {kind} path')
                        records.append({
                            'file': member.name,
                            'document': document,
                            'dataTypes': rule.get('dataTypes', []),
                            'name': rule.get('name', ''),
                            'kind': kind,
                            'field': field,
                            'lastEventAlias': field.startswith('lastEvent.'),
                        })
    aliases = [record for record in records if record['lastEventAlias']]
    affected_files = sorted({record['file'] for record in aliases})
    result = {
        'revision': revision,
        'scope': 'YAML groupBy and deduplicateBy only; references in descriptions, where and history queries excluded',
        'ruleFilesScanned': rule_files,
        'configuredRuleFiles': len({record['file'] for record in records}),
        'configuredFieldOccurrences': len(records),
        'lastEventRuleFiles': len(affected_files),
        'lastEventFieldOccurrences': len(aliases),
        'lastEventDistinctPaths': len({record['field'] for record in aliases}),
        'lastEventRuleFilesByKind': {
            kind: len({r['file'] for r in aliases if r['kind'] == kind})
            for kind in ('groupBy', 'deduplicateBy')
        },
        'lastEventRuleFilesByDirectory': dict(sorted(Counter(path.split('/')[1] for path in affected_files).items())),
        'runtimeStatus': 'Potential query changes only. Scalar value presence, deployed rules and alert-volume impact are not established by static inventory.',
        'fields': records,
    }
    print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == '__main__':
    main()

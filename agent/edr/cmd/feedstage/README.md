# ThreatWinds hash-feed staging

`feedstage` creates a private directory with fixed input snapshots for `feedpack`. It
discovers advertised cumulative MD5, SHA1 and SHA256 lists from the
[ThreatWinds catalog](https://docs.threatwinds.com/feeds/list), downloads only
those combinations at requested levels, checks each archive against the
publisher's `bases.sum` before and after acquisition, validates the archive and
digest lines, and copies the supplied official signature containers. It writes
`feedpack-spec.json` and `receipt.json` only after the complete staging pass.
Empty advertised hash lists are recorded in the receipt and omitted from the
pack specification. No daily delta is applied to a cumulative snapshot.

Example input:

```json
{
  "version": "2026-10-26-test",
  "levels": ["level1", "level2", "level3"],
  "require_detached_signatures": true,
  "files": [
    {"name": "main.cvd", "path": "/verified/db/main.cvd", "official": true, "detached_signature": "/verified/db/main.cvd.sign"},
    {"name": "daily.cvd", "path": "/verified/db/daily.cvd", "official": true, "detached_signature": "/verified/db/daily.cvd.sign"},
    {"name": "bytecode.cvd", "path": "/verified/db/bytecode.cvd", "official": true, "detached_signature": "/verified/db/bytecode.cvd.sign"}
  ]
}
```

From `agent/`:

```sh
go run ./edr/cmd/feedstage -spec /private/inputs.json -out /private/staged-release
go run ./edr/cmd/feedpack -spec /private/staged-release/feedpack-spec.json \
  -key /private/signing-seed.b64 -out /private/signed-release \
  -sequence 1 -expires 2026-11-26T00:00:00Z \
  -require-detached-signatures
```

The output path must be new. `feedstage` uses the public endpoint by default.
Set `THREATWINDS_TOKEN` or both `THREATWINDS_API_KEY` and
`THREATWINDS_API_SECRET` in its environment to use a subscription. Credentials
are never written to the receipt. The production CLI uses HTTPS and rejects
redirects; each response, decoded archive, and the overall generation have
size limits. Its 20-minute deadline also bounds stalled requests.

`bases.sum` uses MD5 as a transfer-integrity check; it is **not** a publisher
signature. [ThreatWinds documents that endpoint and the archive
format](https://docs.threatwinds.com/feeds/download). Verify the origin and
redistribution rights of each supplied ClamAV container separately. `feedpack`
validates embedded container signatures and the combined native database,
removes only safe whole-file hash overlaps, and signs the resulting generation
with the operator's Ed25519 key. Keep that signing key outside the staging and
release directories. A staged or locally signed generation is not a production
release until it passes endpoint acceptance and the normal promotion process.

For official CVDs, `detached_signature` points to the corresponding upstream
`.cvd.sign` file. Staging verifies it against the pinned ClamAV release root,
copies the exact verified pair into the private snapshot, records the signature
size and SHA-256 in `receipt.json`, and passes its staged path to `feedpack`.
Set `require_detached_signatures` in the staging input and
`-require-detached-signatures` for `feedpack` to reject any official CVD without
the detached proof. This does not authenticate `.cld` files or replace the
native manifest signature.

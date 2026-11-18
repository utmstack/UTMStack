# Local signed-feed promotion

`feedpromote` installs an already signed `feedpack` output directory into an
existing local directory served by the native HTTPS mirror. It reads only the
public Ed25519 key. It verifies the signed manifest, expiry, sequence and each
content-addressed artifact before changing the mirror. It copies artifacts
without overwriting existing content, flushes them, and replaces
`manifest.json` atomically last. An unchanged sequence/payload can repair a
missing artifact; rollback or reuse of a sequence for different content is
rejected. A malformed existing mirror manifest also blocks promotion.

From `agent/`:

```sh
go run ./edr/cmd/feedpromote \
  -release /secure/signed-release \
  -mirror /srv/edr/native \
  -public-key /secure/native-feed-public.b64
```

The mirror directory must already exist and be served at the configured HTTPS
native-feed URL. Run this on the publisher's local POSIX filesystem; a remote
or unusual filesystem must be qualified separately for hard-link, rename and
directory-sync behavior. The public-key file contains the same base64 32-byte Ed25519
key pinned by endpoints. The publisher process needs exclusive write access to
the mirror. A `.feedpromote.lock` file prevents concurrent invocations; after a
publisher crash, inspect the mirror and remove that stale lock before retrying.
Content-addressed artifacts from prior releases are retained, so a separately
signed, higher-sequence rollback release can reuse them. This command does not
generate that rollback, rotate keys, schedule acquisition, serve HTTPS, or test
that endpoints can reach the mirror. Keep the signing key and staged upstream
feeds outside the served mirror.

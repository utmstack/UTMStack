# UTMStack Plugins

Documentation on how to create and maintain custom plugins can be found in: https://github.com/utmstack/UTMStack/wiki

## Tenant configuration SDK contract

V12 writes rule identities as strings in `pipeline/tenants.yaml`. All plugins
must use the same compatible `github.com/threatwinds/go-sdk` revision as the
backend and the `events` plugin. Older SDKs deserialize `disabledRules` as
`uint64` and reject named rules; they also omit tenant pipeline preferences.
Do not clear selections or convert rule names to arbitrary numeric values.

The tenant SDK compatibility workflow tests the real protobuf reader under each
plugin's own module graph. Run locally, for example:

```sh
cd plugins/playground
go test -mod=readonly ../../tools/runtimecompat/tenant_config_test.go
go test -mod=readonly ./...
```

The backend producer contract and SDK version alignment are covered by:

```sh
cd backend
go test ./modules/eventprocessing/domain
```

Rebuild and deploy the event processor image to distribute the updated plugin
binaries. Rebuilding the backend alone cannot replace SDK code compiled into
plugins. The event processor base image and its bundled plugins are maintained
separately; verify their tenant schema compatibility as part of release testing.

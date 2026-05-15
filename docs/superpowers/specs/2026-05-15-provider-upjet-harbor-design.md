# provider-upjet-harbor — Design

**Date:** 2026-05-15
**Status:** Approved (pending user review of this spec)
**Authors:** jonasz-lasut + Claude

## Purpose

Build a Crossplane v2 Upjet-based provider for [Harbor](https://goharbor.io/),
generated on top of the upstream
[`goharbor/terraform-provider-harbor`](https://github.com/goharbor/terraform-provider-harbor)
at tag `v3.11.6`. The provider reconciles Harbor resources (projects, robot
accounts, registries, replications, etc.) declaratively via Kubernetes CRDs.

## Non-goals

- Forking the upstream Terraform provider (its `provider.Provider()` is already
  exported — no `xpprovider` shim is needed).
- CI/CD pipelines and image publishing in v0. The Makefile/build machinery
  inherited from the upjet template will be present but not exercised by GitHub
  Actions; the user will run `make` locally.
- Hot-reload of credentials, per-resource integration tests, or covering
  cluster-scoped MRs in v0 (cluster scaffolding is in place for future use).

## Decisions

| Decision | Value | Rationale |
|---|---|---|
| Go module path | `github.com/jonasz-lasut/provider-upjet-harbor` | Personal scope; transferrable later. |
| Upstream TF provider | `github.com/goharbor/terraform-provider-harbor` at `v3.11.6` | Pinned in `go.mod`; imported as a Go module (no `vendor/` directory). |
| Upjet integration mode | **SDK-mode (no-fork)** | `provider.Provider()` is publicly importable; avoids spawning child TF binary. |
| MR scope | Namespaced only (Crossplane v2) | Dual-scope directory layout retained so cluster MRs can be added later. |
| CRD API group (cluster) | `harbor.crossplane.io` | Scaffolded but unused initially. |
| CRD API group (namespaced) | `harbor.m.crossplane.io` | Crossplane v2 `.m.` (managed) infix convention. |
| Resource set | All ~20 upstream resources | `WithIncludeList(".*$")`; refine external-name configs incrementally. |
| Auth modes | `username`+`password`, `bearer_token` | Skip `session_id` (rarely used outside Harbor's web UI). |
| Bootstrap approach | Scaffold from `crossplane/upjet-provider-template` | Clean foundation; no azuread baggage. |
| Schema source | Committed `config/schema.json`, regenerated on TF bumps | Same pattern as `provider-upjet-azuread`. |

## Architecture

### Target directory layout

`/root/code/xp/provider-upjet-harbor/`:

```
provider-upjet-harbor/
├── apis/
│   ├── cluster/                       # cluster-scoped ProviderConfig types only; no MR groups in v0
│   │   ├── v1beta1/                   # ProviderConfig + ProviderConfigUsage types
│   │   └── zz_register.go             # scheme registration (no resource groups added)
│   └── namespaced/
│       ├── v1beta1/                   # ProviderConfig + ProviderConfigUsage types (namespaced)
│       ├── <group>/<version>/         # generated CRD types per resource group
│       └── zz_register.go             # scheme registration for all generated groups
├── cmd/
│   ├── generator/main.go              # SDK-mode generation entrypoint
│   └── provider/main.go               # SDK-mode controller entrypoint
├── config/
│   ├── cluster/                       # empty; structure preserved
│   ├── namespaced/
│   │   ├── project/config.go          # project, members, webhook, label, immutable_tag_rule, preheat_instance
│   │   ├── robot/config.go            # robot_account
│   │   ├── identity/config.go         # user, group
│   │   ├── registry/config.go         # registry, replication, retention_policy
│   │   ├── maintenance/config.go      # garbage_collection, purge_audit_log, tasks
│   │   ├── harborconfig/config.go     # config_auth, config_system, config_security
│   │   └── vuln/config.go             # interrogation_services
│   ├── external_name.go               # central ExternalNameConfigs map
│   ├── provider.go                    # GetProvider + GetProviderNamespaced (SDK-mode signatures)
│   ├── schema.json                    # generated once, committed
│   └── provider-metadata.yaml         # docs scraped from .work/terraform-provider-harbor/
├── examples/, examples-generated/{cluster,namespaced}/
├── hack/
│   ├── prepare.sh                     # used once during scaffold, then deleted
│   ├── boilerplate.go.txt
│   ├── boilerplate.yaml.txt
│   └── generate-schema.go             # NEW: dumps config/schema.json from harborprovider.Provider()
├── internal/
│   ├── apis/                          # scheme registration helpers (from template)
│   ├── clients/harbor.go              # TerraformSetupBuilder (SDK-mode, basic + bearer)
│   ├── controller/{cluster,namespaced}/  # generated controllers
│   ├── features/                      # feature flags (from template)
│   └── version/                       # version embedding (from template)
├── package/crds/                      # generated CRD manifests
├── cluster/images/provider-upjet-harbor/  # Dockerfile (from template)
├── .work/terraform-provider-harbor/   # cloned at v3.11.6 for docs scraping (gitignored)
├── build/                             # git submodule → crossplane/build
├── go.mod                             # module github.com/jonasz-lasut/provider-upjet-harbor
├── Makefile
├── LICENSE                            # Apache-2.0 (from template)
└── README.md
```

### Data flow — code generation

```
goharbor/.../provider.Provider()  ──┐
                                    ├──►  cmd/generator/main.go
config.GetProvider*(ctx, sdkP, true) ┘                │
                                                       ▼
                                          upjet pipeline.Run(pc, pns, root)
                                                       │
                       ┌───────────────────────────────┼───────────────────────────────┐
                       ▼                               ▼                               ▼
              apis/namespaced/<g>/<v>/  internal/controller/namespaced/  package/crds/, examples-generated/

.work/terraform-provider-harbor/website/docs  ──►  upjet scraper  ──►  config/provider-metadata.yaml
```

### Data flow — runtime

```
harbor.m.crossplane.io/v1beta1.ProviderConfig + Secret (in user namespace)
       │
       ▼
internal/clients/harbor.go::TerraformSetupBuilder
       │
       ▼
terraform.Setup{Configuration: { url, (username+password | bearer_token),
                                  insecure, api_version, robot_prefix }}
       │
       ▼
upjet controller (SDK-mode, in-process)  ──►  goharbor/.../provider.Provider()  ──►  Harbor API
```

## Components

### `config/provider.go`

Signatures adapted from `provider-upjet-azuread` (SDK-mode, no-fork) but
simplified. Two functions, one per scope:

```go
func GetProvider(ctx context.Context, sdkProvider *schema.Provider) (*ujconfig.Provider, error)

func GetProviderNamespaced(ctx context.Context, sdkProvider *schema.Provider) (*ujconfig.Provider, error)
```

The azuread `generationProvider bool` parameter and the dual-source schema swap
(JSON during generation, live SDK at runtime) are omitted: harbor's schema has
no float/int coercion problems requiring the workaround.

Both build the upjet `Provider` via:

```go
pc := ujconfig.NewProvider(
    []byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
    ujconfig.WithRootGroup("harbor.crossplane.io"),     // or "harbor.m.crossplane.io"
    ujconfig.WithIncludeList([]string{".*$"}),
    ujconfig.WithFeaturesPackage("internal/features"),
    ujconfig.WithTerraformProvider(sdkProvider),
    ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
    // Namespaced only:
    ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{
        ManagedResourceNamespace: "crossplane-system",
    }),
)
```

The `bumpVersionsWithEmbeddedLists` singleton-list machinery from azuread is
**omitted** — Harbor's TF schema has no singleton-list APIs we need to convert.

### `config/namespaced/<group>/config.go`

Per-resource external-name configurators. Seeded with a stub per group; v0
ships with all resources using upjet's default `NameAsIdentifier` unless the
upstream resource's ID is composite (refined incrementally as needed).

### `config/external_name.go`

Central `var ExternalNameConfigs = map[string]config.ExternalName{...}`. Same
shape as the template; populated from each `config/namespaced/<group>` package.

### `cmd/generator/main.go`

```go
import harborprovider "github.com/goharbor/terraform-provider-harbor/provider"

func main() {
    ...
    sdkProvider := harborprovider.Provider()
    pc, err  := config.GetProvider(ctx, sdkProvider)          // cluster (empty resources)
    pns, err := config.GetProviderNamespaced(ctx, sdkProvider) // namespaced
    pipeline.Run(pc, pns, absRootDir)
}
```

No `xpprovider.GetProviderSchema` indirection — calls upstream `Provider()`
directly.

### `cmd/provider/main.go`

Same pattern: imports `harborprovider`, builds the SDK provider once, passes
to both `GetProvider*` calls, wires `clusterController` + `namespacedController`
(though cluster has no resources in v0). Keeps from the template:

- leader election (`crossplane-leader-election-provider-harbor`)
- webhook server + cert handling
- SafeStart CRD-gate
- MR metrics + state metrics

Removes from azuread reference:

- `bootcheck` (Azure-environment validation, irrelevant)
- singleton-list conversion register (no singleton lists to convert)
- OIDC / change-logs (out of scope for v0)

### `internal/clients/harbor.go`

```go
func TerraformSetupBuilder(tfProvider *schema.Provider) terraform.SetupFn {
    return func(ctx context.Context, c client.Client, mg xpresource.Managed) (terraform.Setup, error) {
        ps := terraform.Setup{}
        pcSpec, err := resolveProviderConfig(ctx, c, mg)
        if err != nil { return ps, err }

        creds, err := extractCredentials(ctx, c, pcSpec)
        if err != nil { return ps, err }

        ps.Configuration = map[string]any{
            "url":      pcSpec.URL,
            "insecure": pcSpec.Insecure,
        }
        if pcSpec.APIVersion != nil { ps.Configuration["api_version"] = *pcSpec.APIVersion }
        if pcSpec.RobotPrefix != "" { ps.Configuration["robot_prefix"] = pcSpec.RobotPrefix }

        switch {
        case creds.BearerToken != "":
            ps.Configuration["bearer_token"] = creds.BearerToken
        case creds.Username != "" && creds.Password != "":
            ps.Configuration["username"] = creds.Username
            ps.Configuration["password"] = creds.Password
        default:
            return ps, errors.New(errNoCredentials)
        }
        return ps, nil
    }
}
```

ProviderConfig spec fields (added to template-generated `apis/{cluster,namespaced}/v1beta1/types.go`):

| Field | Type | Required | Notes |
|---|---|---|---|
| `url` | string | yes | Harbor base URL, e.g. `https://harbor.example.com` |
| `insecure` | bool | no | Skip TLS verify; default `false` |
| `apiVersion` | *int | no | Harbor API version; default 2 |
| `robotPrefix` | string | no | Override `robot$` prefix |
| `credentials` | ProviderCredentials | yes | Standard Crossplane credentials ref |

Credentials secret payload (JSON), one of:

```json
{ "username": "admin", "password": "Harbor12345" }
```
```json
{ "bearer_token": "eyJhbGciOi..." }
```

If both are present, `bearer_token` wins.

### `config/schema.json`

Committed JSON dump of `harborprovider.Provider()`'s schema (via
`terraform-json`). Regenerated by a small `hack/generate-schema.go` only when
bumping the upstream TF version.

### `.work/terraform-provider-harbor/`

Cloned at tag `v3.11.6` for the docs scraper. Provisioned by a Makefile target
`fetch-tf-provider-source` that runs `git clone --branch v3.11.6 --depth 1`
into `.work/terraform-provider-harbor/` if the directory is absent.
Gitignored. Provides `docs/resources/` (Harbor's docs layout) to populate
`config/provider-metadata.yaml`. The exact docs subpath is verified during
implementation by inspecting the upstream repo at v3.11.6.

## Error handling

| Boundary | Strategy |
|---|---|
| Generation (`cmd/generator/main.go`) | `kingpin.FatalIfError(err, "...")` — loud dev-time crashes. |
| Provider boot (`cmd/provider/main.go`) | `kingpin.FatalIfError(...)` — pod crash-loops, Crossplane restarts. |
| ProviderConfig resolution (`clients/harbor.go`) | Wrapped errors: `errNoProviderConfig`, `errGetProviderConfig`, `errTrackUsage`, `errExtractCredentials`, `errUnmarshalCredentials`, `errNoCredentials`. Surfaces as MR `Synced=False`. |
| Terraform execution (in-process) | Harbor API errors bubble up as TF diagnostics → upjet maps to MR conditions automatically. |

Out of scope in v0: credential hot-reload, custom retry/backoff (upjet's
reconciler handles via `MaxConcurrentReconciles` + poll).

## Testing

- **Unit:** `internal/clients/harbor_test.go` — table-driven `TerraformSetupBuilder` test covering basic-auth, bearer-token, both-set (bearer wins), missing creds (error), malformed credentials JSON (error). URL is required at the CRD schema level (not re-validated in Go), so no "missing URL" test case. Fake `client.Client`; no real Harbor.
- **Build verification (manual):** `make generate` (idempotent, no diff), `make build` (binaries compile), `make lint` (golangci-lint).
- **Smoke (manual, definition of done for v0):**
  1. `make generate` clean and idempotent.
  2. `make build` produces working binaries.
  3. Provider package installs into a kind cluster (Crossplane v2) and reports `Healthy=True` against a real Harbor.
  4. A `Project` CR in `crossplane-system` reaches `Ready=True` and a project actually exists in Harbor.

**Not in v0:** GitHub Actions CI, image publishing, e2e tests for every resource.

## Open questions

None at design time. Implementation may surface external-name nuances per
Harbor resource — these will be resolved during plan execution by inspecting
the upstream resource's `Read`/`Create` Go code or schema.

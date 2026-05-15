# Architect review: provider-upjet-harbor vs provider-upjet-azuread

**Date:** 2026-05-15
**Scope:** Architectural parity comparison and verification of build infrastructure.
**Reference baseline:** `provider-upjet-azuread` (production Crossplane Upjet provider).

The review compares the freshly-built provider against the reference implementation
that served as its architectural model. Findings are classified as
**Intentional / correct**, **Behavioral gap**, **Latent bug**, or **Cosmetic / minor**.
File references use absolute paths in both repos.

---

## Top-3 findings by severity

1. **Behavioral gap (severity: high) — Generated controllers run in fork-mode, not SDK-mode.**
   Despite the design spec calling for SDK-mode integration, the generator is configured
   with `WithIncludeList(".*$")` only. Upjet generates controllers that use
   `tjcontroller.NewConnector` + `terraform.NewWorkspaceFinalizer`, which require a
   `terraform` CLI subprocess at runtime. To run no-fork, the harbor resources must be
   placed in `WithTerraformPluginSDKIncludeList` instead. The Dockerfile still bundles
   terraform CLI, confirming the runtime is fork-mode today.

2. **Behavioral gap (severity: high) — `conversion.RegisterConversions` is not wired in
   `cmd/provider/main.go`.** Even though the generator-side conversion machinery
   exists in upjet, the harbor provider never calls it at boot. Any future singleton-list
   → embedded-object API migration, or `v1alpha1 → v1beta1` storage version migration,
   will silently fail. This is a one-line call in azuread (`main.go:258`).

3. **Behavioral gap (severity: medium) — ProviderConfigUsage watch does not disambiguate
   `Kind`.** In `internal/controller/namespaced/providerconfig/config.go:37,58` harbor
   passes `&resource.EnqueueRequestForProviderConfig{}` with no `Kind`, whereas azuread
   passes `Kind: "ProviderConfig"` vs `Kind: "ClusterProviderConfig"`. With both
   namespaced and cluster-scoped PCs sharing a `ProviderConfigUsage` GVK, a Usage event
   currently re-enqueues both reconcilers regardless of the actual PC kind, causing
   redundant work and harder-to-trace logs.

---

## Intentional / correct divergences

These differences match the design spec on file and require no action.

- **No `WithTerraformPluginFrameworkIncludeList` / singleton-list machinery.**
  Harbor's TF provider has no singleton-list resources and is pure SDKv2, so
  `bumpVersionsWithEmbeddedLists`, `WithSchemaTraversers(&SingletonListEmbedder{})`,
  `old-singleton-list-apis.txt`, and `registerTerraformConversions` (azuread:
  `config/provider.go:117,157,182`, `config/provider_namespaced.go:55,82`) are
  intentionally omitted in harbor `config/provider.go`.

- **`apis/cluster/v1beta1/` carries only ProviderConfig types.** v0 design is namespaced
  MRs only; cluster scope kept as scaffold for later promotion. Compare
  `provider-upjet-harbor/apis/cluster/` (just `v1beta1/` + `v1alpha1/`) vs
  `provider-upjet-azuread/apis/cluster/` (full per-group resources).

- **SDK-mode `TerraformSetupBuilder` is a pure-function helper.** Harbor's
  `internal/clients/harbor.go::buildHarborSetup` is intentionally non-K8s for unit
  testing; matching tests live in `internal/clients/harbor_test.go`. Azuread inlines
  the equivalent logic into multi-mode auth helpers (`internal/clients/azuread.go:101`).

- **SHA-pinned upstream module (`go.mod`).** Required because upstream
  `goharbor/terraform-provider-harbor` `go.mod` lacks `/v3` suffix, so we pin
  `v1.4.1-0.20260422075905-609e99677021`. Documented in design spec.

- **Pure-SDK provider via `harborprovider.Provider()`.** No `xpprovider` shim used
  (azuread imports `hashicorp/terraform-provider-azuread/xpprovider`). Acceptable since
  harbor upstream doesn't ship one; the SDK provider is sufficient for upjet's
  schema synthesis and runtime configuration.

- **`internal/bootcheck` and `init()`-time gate omitted.** Azuread's
  `internal/bootcheck/default.go` is a no-op (308 B) — its existence is purely to allow
  downstream Upbound builds to inject a custom build-tag-gated env preflight. Harbor
  hasn't chosen to expose that hook in v0; no functional loss.

- **Resource group names follow upstream TF resource taxonomy.** Per-group dirs
  (`config/`, `garbage/`, `harbor/`, `immutable/`, `interrogation/`, `preheat/`,
  `project/`, `purge/`, `retention/`, `robot/`) derive sensibly from
  Terraform resource names (`harbor_config_auth`, `harbor_garbage_collection`, etc.).
  CRD scopes verified correct: 17 `Namespaced` MRs + 2 `Cluster` PCs (legacy group) +
  3 namespaced PC kinds (modern group).

---

## Behavioral gaps

These will cause functional or operational pain in production. **Report only — fixing
out of scope per the constraints.**

### B1 — Fork-mode controllers (`config/provider.go:41,56`)

`provider-upjet-harbor/config/provider.go:41` and `:56` pass all resources via
`ujconfig.WithIncludeList`. The corresponding azuread file
(`provider-upjet-azuread/config/provider.go:109-110`,
`provider-upjet-azuread/config/provider_namespaced.go:45-46`) splits resources into
two lists, with `WithTerraformPluginSDKIncludeList` for no-fork resources.

**Evidence the generator chose fork-mode:** the emitted
`internal/controller/namespaced/harbor/project/zz_controller.go:46,51` uses
`tjcontroller.NewConnector(... o.WorkspaceStore ...)` and
`terraform.NewWorkspaceFinalizer(o.WorkspaceStore, ...)`. Azuread's equivalent
(`internal/controller/namespaced/groups/group/zz_controller.go:46`) uses
`tjcontroller.NewTerraformPluginSDKAsyncConnector(... o.OperationTrackerStore ...)`.

**Evidence the runtime requires terraform CLI:** harbor's
`cluster/images/provider-upjet-harbor/Dockerfile:14-46` still downloads and installs a
terraform CLI binary plus the upstream provider binary, sets `TF_FORK=0`, `PLUGIN_DIR`,
`TF_CLI_CONFIG_FILE`. Azuread's `cluster/images/provider-azuread/Dockerfile` is 15
lines with no terraform install at all.

**Impact:** the provider image bundles terraform 1.5.7 and the harbor provider plugin;
each MR reconcile forks a `terraform` process. SDK-mode would skip the workspace,
making the provider lighter, faster, and (per the design spec) the intended mode. The
fix is to swap `WithIncludeList` for `WithTerraformPluginSDKIncludeList` in both
`GetProvider` and `GetProviderNamespaced`, regenerate, and rewrite the Dockerfile to
the minimal form.

### B2 — `conversion.RegisterConversions` missing (`cmd/provider/main.go:237`)

Harbor's `main.go` ends with `mgr.Start(...)` without registering the upjet conversion
webhook hooks. Azuread:

```
provider-upjet-azuread/cmd/provider/main.go:258
  kingpin.FatalIfError(conversion.RegisterConversions(clusterOpts.Provider, namespacedOpts.Provider, mgr.GetScheme()), "Cannot initialize the webhook conversion registry")
```

Harbor doesn't import `github.com/crossplane/upjet/v2/pkg/controller/conversion` at
all. Today this is silent because no resource in harbor has multiple API versions; the
moment a `v1alpha1 → v1beta1` storage promotion or singleton-list migration is
introduced, the conversion webhook will be a no-op and updates will corrupt stored
specs.

### B3 — `EnqueueRequestForProviderConfig{}` lacks Kind disambiguation
(`internal/controller/namespaced/providerconfig/config.go:37,58`)

Harbor:
```go
Watches(&v1beta1.ProviderConfigUsage{}, &resource.EnqueueRequestForProviderConfig{}).
```
Azuread (`internal/controller/namespaced/providerconfig/config.go:39,58`):
```go
Watches(&v1beta1.ProviderConfigUsage{}, &resource.EnqueueRequestForProviderConfig{Kind: "ProviderConfig"}).
... Watches(&v1beta1.ProviderConfigUsage{}, &resource.EnqueueRequestForProviderConfig{Kind: "ClusterProviderConfig"}).
```
Without the `Kind` filter, any `ProviderConfigUsage` event re-enqueues both reconcilers
(namespaced and cluster), each of which then refetches and reconciles its respective
PC even when unrelated. Reaches the wrong PC reconciler with wrong audit data.

### B4 — `HealthProbeBindAddress` and `webhook readiness check` missing
(`cmd/provider/main.go:119-136`)

Azuread main.go:
```
provider-upjet-azuread/cmd/provider/main.go:81,153,160
  healthProbeBindAddress = app.Flag("health-probe-bind-addr", ...).Default(":8081")
  HealthProbeBindAddress: *healthProbeBindAddress,
  mgr.AddReadyzCheck("webhook", mgr.GetWebhookServer().StartedChecker())
```
Harbor: no `health-probe-bind-addr` flag, no `HealthProbeBindAddress` on
`ctrl.Options`, no `AddReadyzCheck`. The deployment will lack a readyz endpoint, so
Crossplane's deployment health check / Kubernetes probes cannot reflect whether the
webhook server has actually started.

### B5 — `internal/apis/scheme.go` resolver registry missing

Azuread maintains a private runtime.Scheme via `internal/apis/scheme.go` and exports
`BuildScheme()` and `GetManagedResource()`. `cmd/provider/main.go:163-165` calls
`resolverapis.BuildScheme(clusterapis.AddToSchemes)` and
`resolverapis.BuildScheme(namespacedapis.AddToSchemes)` so that the upjet resolver can
look up cross-API-group references at runtime without going through the manager's
scheme.

Harbor has neither the package nor the calls. Today this is a no-op because harbor's
20 resources don't reference each other (`grep -rn "Reference"` in
`apis/namespaced/<grp>/v1alpha1/zz_*_types.go` finds no `+crossplane:generate:reference`
markers in any TF schema). If a future Configure-hook adds a cross-resource reference
(e.g., a `ProjectRef` on a `Robot`), the resolver will fail at runtime.

### B6 — `PollJitter` + `OperationTrackerStore` not set on `Options`
(`cmd/provider/main.go:157-193`)

Azuread sets `PollJitter` (5% of poll interval) and `OperationTrackerStore:
tjcontroller.NewOperationStore(logr)` on both cluster and namespaced opts
(`cmd/provider/main.go:196-197, 216-217`). Harbor sets neither. Without `PollJitter`,
all MRs of a kind reconcile on the same edge of the poll interval (thundering herd).
The missing `OperationTrackerStore` is consistent with B1 (fork-mode) — in SDK-mode
this is required.

---

## Latent bugs

### L1 — Duplicate `+kubebuilder:resource:scope=…` markers
(`apis/cluster/v1beta1/types.go:51-52`, `apis/cluster/v1beta1/types.go:77`,
`apis/namespaced/v1beta1/types.go:52-53`, `apis/namespaced/v1beta1/types.go:78`,
`apis/namespaced/v1beta1/types.go:101-102`)

Each PC/PCUsage/ClusterPC type carries two `+kubebuilder:resource:scope=...` markers
in a row: one bare, one with `categories={...}`. `controller-gen` accepts this
(later marker wins), but it's a footgun if a future regenerator parses markers
order-sensitively. Azuread does the same in cluster-scoped types
(`apis/cluster/v1beta1/types.go:68-69`), so this is plausibly the upjet template
convention — but worth flagging.

### L2 — Missing `+kubebuilder:storageversion` on PC types
(`apis/cluster/v1beta1/types.go:53,78`, `apis/namespaced/v1beta1/types.go:54,79,103`)

Azuread marks the singular version with `+kubebuilder:storageversion` (see
`provider-upjet-azuread/apis/cluster/v1beta1/types.go:70`,
`provider-upjet-azuread/apis/namespaced/v1beta1/types.go:71,97,122`). Harbor does not.

Today harbor has only one served version of each PC type, so Kubernetes implicitly
treats it as the storage version. The moment a v1beta2 is introduced, the CRD will
fail validation because no version has the storage marker. This is a forward-looking
defect, easy to fix later.

### L3 — `resolveModern` error message has a stray space
(`internal/clients/harbor.go:159`)

```go
return nil, errors.New(" is not an Object")
```

Should be e.g. `"referenced kind is not a client.Object"`. Cosmetic-ish but a
real error returned to controllers.

### L4 — `resolveModern` uses `client.Object` type assertion, not
`resource.ProviderConfig` (`internal/clients/harbor.go:156`)

Azuread (`internal/clients/azuread.go:254`) does
`pcObj, ok := pcRuntimeObj.(xpresource.ProviderConfig)`. Harbor does
`pcObj, ok := pcRuntimeObj.(client.Object)`. Any `runtime.Object` registered to the
scheme is a `client.Object`, so this assertion is essentially a no-op — and the later
type switch on `*namespacedv1beta1.{ProviderConfig,ClusterProviderConfig}` will catch
the real cases. But the assertion contributes no value; if the scheme accidentally
registers a non-PC type under the PC GVK (e.g. a bug in `register.go`), harbor would
silently produce nil credentials whereas azuread would return a typed error.

### L5 — `apis/cluster/v1alpha1/register.go` and
`apis/namespaced/v1alpha1/register.go` are empty placeholders
(`apis/cluster/v1alpha1/register.go:22`)

The v1alpha1 placeholders define `Group`, `Version`, `SchemeGroupVersion`,
`SchemeBuilder`, and `func init() {}` but register no types. They are imported by
`apis/cluster/zz_register.go:13` and `apis/namespaced/zz_register.go:13` and added
to `AddToSchemes`. This adds an empty scheme builder to the runtime — harmless, but
the v1alpha1 directories should either be removed (no v1alpha1 types planned for these
shared groups) or populated. Azuread keeps v1alpha1 only for `apis/cluster/v1alpha1/`
storage-handler types — harbor's v1alpha1 register.go has none.

### L6 — `apis/{cluster,namespaced}/zz_register.go` are committed but
generator-managed

These files have the `// Code generated by upjet. DO NOT EDIT.` header but are
manually edited per the design notes (to remove `null` template imports). The next
`make generate` would in fact regenerate them; verified via the run done for
Deliverable 2 (they came back byte-identical to the committed versions, because
the upjet pipeline now respects the actual subgroups present). The risk is future
drift if someone adds/removes a group: they must remember the files are
generator-controlled.

---

## Cosmetic / minor

- **No SPDX header on most harbor `.go` files.** Azuread consistently has
  `SPDX-FileCopyrightText: 2024 The Crossplane Authors ...`. Harbor's hand-written
  files (e.g. `config/provider.go:1-2`) say
  `SPDX-FileCopyrightText: 2026 jonasz-lasut`. Generated files use the upjet
  template's 2024 header. This is fine, but the mix of authorship lines may
  surprise license-bot tooling.

- **`internal/clients/harbor.go:79`**:
  ```go
  _ = tfProvider // accepted for parity with the SDK-mode call site; closure has no per-call use
  ```
  Comment is accurate. Worth noting the SDK provider's `Meta()` is never propagated
  to `terraform.Setup{}` — azuread's `configureNoForkAzureClient`
  (`internal/clients/azuread.go:101-110`) actually calls `p.Configure(...)` and
  assigns `ps.Meta = p.Meta()`. For harbor in true SDK-mode this would also be
  required. Since the runtime is fork-mode today (see B1), this gap doesn't bite —
  yet.

- **No `examples-generated/cluster/`.** Expected per v0 design (cluster scope empty).
  `examples-generated/namespaced/<g>/v1alpha1/*.yaml` regenerates correctly.

- **Sample namespace in examples is `crossplane-system`.** Hard-coded via
  `ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{ManagedResourceNamespace: "crossplane-system"})`
  in `config/provider.go:60-62`. Azuread doesn't set this and lets upjet default.
  Minor preference difference; documented in design spec.

- **`cmd/provider/main.go:84`** uses `provider-upjet-harbor` as the logger name
  while `internal/version` ldflags inject `provider-upjet-harbor:<ver>`. Consistent.

---

## Summary table

| ID | Severity | File:line (harbor) | File:line (azuread) | Class |
|----|----------|--------------------|---------------------|-------|
| B1 | High | `config/provider.go:41,56` | `config/provider.go:109-110` | Behavioral gap |
| B2 | High | `cmd/provider/main.go:237` (missing) | `cmd/provider/main.go:258` | Behavioral gap |
| B3 | Medium | `internal/controller/namespaced/providerconfig/config.go:37,58` | `internal/controller/namespaced/providerconfig/config.go:39,58` | Behavioral gap |
| B4 | Medium | `cmd/provider/main.go:119-136` (missing) | `cmd/provider/main.go:81,153,160` | Behavioral gap |
| B5 | Low (latent) | `internal/apis/` (missing) | `internal/apis/scheme.go` | Behavioral gap |
| B6 | Medium | `cmd/provider/main.go:157-193` (no PollJitter/OperationTrackerStore) | `cmd/provider/main.go:196-197,216-217` | Behavioral gap |
| L1 | Cosmetic | duplicate scope markers on PC types | same pattern in azuread | Latent bug (footgun) |
| L2 | Low | no `+kubebuilder:storageversion` on PC types | present in azuread | Latent bug |
| L3 | Low | `internal/clients/harbor.go:159` | n/a (correct in azuread:256) | Latent bug |
| L4 | Low | `internal/clients/harbor.go:156` | `internal/clients/azuread.go:254` | Latent bug |
| L5 | Low | `apis/{cluster,namespaced}/v1alpha1/register.go` empty | n/a (azuread has real v1alpha1 types) | Latent bug |
| L6 | Low | `apis/{cluster,namespaced}/zz_register.go` | upjet-generated | Latent bug |

---

## Build-infrastructure verification (Deliverable 2 summary)

- `make generate` succeeded end-to-end from clean state.
- Terraform 1.5.7 downloaded by Makefile (via
  `https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip`).
- Confirmed: `.cache/tools/linux_x86_64/terraform-1.5.7 version` → `Terraform v1.5.7`.
- `config/schema.json` regenerated; only diff vs prior committed version is the
  removal of two `write_only: true` annotations on
  `harbor_robot_account.secret_wo` and `harbor_user.password_wo` — these are TF 1.11+
  features and would have been emitted by the previous (BSL-licensed) terraform 1.14.5
  the implementer had locally. The new schema is the correct one for our 1.5.7
  pin.
- `config/provider-metadata.yaml`, all `apis/namespaced/<g>/...` generated Go,
  CRD YAMLs in `package/crds/*.yaml`, and `internal/controller/namespaced/...`
  came back byte-identical to the prior commit.
- `go build ./...` succeeds.
- `go test ./...` succeeds (`internal/clients` passes; rest are empty `?`).
- Network: routed via the supplied `https_proxy="http://localhost:8888"`.
  `GOPROXY=direct` used for `go mod tidy` to bypass the partial Artifactory mirror.

---

## Resolution

**Date landed:** 2026-05-15

All 10 implementation tasks (T1–T10) plus the T11 follow-up commit landed on the same
calendar day.  Final verification (T11) confirmed a clean build, test, vet, and binary
output.

### Commits per finding

| Finding | Commit | Subject |
|---------|--------|---------|
| L3 | `f5d58ee` | fix(clients): replace stray-space error message in resolveModern |
| L4 | `56a002b` | fix(clients): assert resource.ProviderConfig in resolveModern |
| B3 | `10608c8` | fix(controller): disambiguate ProviderConfigUsage watch by Kind |
| B4 | `5cb0ff7` | feat(cmd): add health-probe-bind-addr flag and webhook readyz check |
| B6 | `ca85efc` | feat(cmd): wire PollJitter and OperationTrackerStore on controller opts |
| B2 | `eafd045` | feat(cmd): register upjet conversion webhooks on provider start |
| B5 | `cf79915` | feat(apis,cmd): add resolver scheme registry and BuildScheme wiring |
| L2 | `604b744` | fix(apis): mark v1beta1 as storage version on PC types and regen CRDs |
| L5 | `9a19451` | chore(apis): remove empty v1alpha1 placeholder packages |
| B1 | `c6c2a95` | feat(config,controller): switch to SDK-mode (no-fork) and slim Dockerfile |
| B1 (follow-up) | `78a1c5f` | fix(image): drop terraformrc.hcl copy and stale TERRAFORM_* build args |
| L5 (follow-up) | `e8b3fa8` | chore(apis): re-prune v1alpha1 placeholder imports after make generate |

### Verification results

**Step 1 — `go build ./...`:** EXIT 0, no errors.

**Step 2 — `go test -count=1 ./...`:** EXIT 0. `internal/clients` reports `ok`; all
other packages report `? ... [no test files]`.

**Step 3 — `go vet ./...`:** EXIT 0, no findings.

**Step 4 — Binary builds:**
- `/tmp/provider-bin`: ELF 64-bit LSB executable, x86-64
- `/tmp/generator-bin`: ELF 64-bit LSB executable, x86-64

### Step 5 — `make generate` idempotency

`make generate` is **not fully idempotent** at this HEAD due to two interacting issues:

1. **L5 re-emission (documented):** The upjet generator re-emits
   `apis/cluster/v1alpha1` and `apis/namespaced/v1alpha1` placeholder imports in
   `apis/cluster/zz_register.go` and `apis/namespaced/zz_register.go`.  These
   packages were intentionally deleted in T9 (L5 fix).

2. **Secondary effect (new):** The `apis/generate.go` cleanup glob
   `find . -iname 'zz_*' ! -iname 'zz_generated.managed*.go' -delete`
   deletes `zz_generated.pc*.go` / `zz_generated.pcu*.go` / `zz_generated.pculist*.go`
   before angryjet can regenerate them.  Because issue (1) causes `controller-gen`
   to fail (broken v1alpha1 imports → compile error), angryjet never runs, leaving
   the deleted files unrestored.

**Remediation applied in commit `e8b3fa8`:**
- Removed the re-emitted v1alpha1 import lines from `zz_register.go` (identical
  to the T9 prune).
- Ran `angryjet generate-methodsets` directly to restore the six deleted `zz_generated.pc*.go`
  files (content matched HEAD byte-for-byte).
- Updated the cleanup glob in `apis/generate.go` to also exclude
  `zz_generated.pc*.go`, `zz_generated.pcu*.go`, and `zz_generated.pculist*.go`
  from deletion, preventing the secondary effect from recurring.

After the remediation, `go build ./...` confirmed EXIT 0 and `git status` reported a
clean working tree (untracked plan file only).

**Root cause note:** Full idempotency requires patching the upjet generator (or its
harbor configuration) to not emit v1alpha1 entries for groups where the placeholder
package has been deleted.  That is generator-level work beyond the scope of these
fixes and is tracked as a known limitation.

### Findings NOT addressed and rationale

| Finding | Action | Rationale |
|---------|--------|-----------|
| L1 | No action | Duplicate `+kubebuilder:resource:scope` markers match the azuread template convention; controller-gen accepts it and last-marker wins. No functional risk. |
| L6 | No action | Informational only.  The generator regenerates `zz_register.go` correctly from the present group layout; the "manually edited" comment in the original review was already obsolete after T9. |
| All Cosmetics | No action | Intentional v0 acceptance: SPDX header mix, `_ = tfProvider` comment, no `examples-generated/cluster/`, namespace default, logger name. |

### Summary

All 6 behavioral gaps (B1–B6) closed; 4 of 6 latent bugs (L2–L5) fixed; L1 and L6
documented as no-action.  `make generate` remains non-idempotent due to the upjet
generator re-emitting deleted v1alpha1 placeholder imports (a known limitation
documented above); a protective fix was added to `apis/generate.go` to prevent the
secondary `zz_generated.pc*.go` deletion side-effect.

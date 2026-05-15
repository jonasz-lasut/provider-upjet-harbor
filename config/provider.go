// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	// Note: imported to embed the provider schema document.
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/harborconfig"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/identity"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/maintenance"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/project"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/registry"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/robot"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/vuln"
)

const (
	resourcePrefix = "harbor"
	modulePath     = "github.com/jonasz-lasut/provider-upjet-harbor"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// harborBasePackages overrides upjet's DefaultBasePackages to drop the
// top-level "v1alpha1" entry. We don't have an apis/{cluster,namespaced}/v1alpha1
// placeholder package (it was removed in the L5 cleanup); leaving it in the
// list causes upjet's pipeline to re-emit a broken import on every
// `make generate` run.
var harborBasePackages = ujconfig.BasePackages{
	APIVersion: []string{
		"v1beta1",
	},
	Controller: []string{
		"providerconfig",
	},
	ControllerMap: map[string]string{
		"providerconfig": ujconfig.PackageNameConfig,
	},
}

// GetProvider returns the cluster-scoped provider configuration. In v0 the
// cluster scope has no managed resources; the structure is preserved so
// cluster MRs can be added later.
func GetProvider(_ context.Context, sdkProvider *schema.Provider) (*ujconfig.Provider, error) {
	pc := ujconfig.NewProvider(
		[]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("harbor.crossplane.io"),
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithTerraformPluginSDKIncludeList([]string{}),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithTerraformProvider(sdkProvider),
		ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
		ujconfig.WithBasePackages(harborBasePackages),
	)
	pc.ConfigureResources()
	return pc, nil
}

// GetProviderNamespaced returns the namespaced provider configuration with
// all upstream Harbor resources included.
func GetProviderNamespaced(_ context.Context, sdkProvider *schema.Provider) (*ujconfig.Provider, error) {
	pc := ujconfig.NewProvider(
		[]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("harbor.m.crossplane.io"),
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithTerraformPluginSDKIncludeList([]string{".*$"}),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithTerraformProvider(sdkProvider),
		ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
		ujconfig.WithBasePackages(harborBasePackages),
		ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{
			ManagedResourceNamespace: "crossplane-system",
		}),
	)

	for _, configure := range []func(p *ujconfig.Provider){
		project.Configure,
		robot.Configure,
		identity.Configure,
		registry.Configure,
		maintenance.Configure,
		harborconfig.Configure,
		vuln.Configure,
	} {
		configure(pc)
	}

	// Resources whose auto-derived ShortGroup duplicates the RootGroup's
	// leading segment (harbor_project, harbor_user, harbor_group, etc.) are
	// flattened so their CRDs land at harbor.m.crossplane.io rather than the
	// duplicate-prefixed harbor.harbor.m.crossplane.io.
	for _, r := range pc.Resources {
		if r.ShortGroup == "harbor" {
			r.ShortGroup = ""
		}
	}

	pc.ConfigureResources()
	return pc, nil
}

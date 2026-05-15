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

// GetProvider returns the cluster-scoped provider configuration. In v0 the
// cluster scope has no managed resources; the structure is preserved so
// cluster MRs can be added later.
func GetProvider(_ context.Context, sdkProvider *schema.Provider) (*ujconfig.Provider, error) {
	pc := ujconfig.NewProvider(
		[]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("harbor.crossplane.io"),
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithTerraformProvider(sdkProvider),
		ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
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
		ujconfig.WithIncludeList([]string{".*$"}),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithTerraformProvider(sdkProvider),
		ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
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

	pc.ConfigureResources()
	return pc, nil
}

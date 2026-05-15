// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/jonasz-lasut/provider-upjet-harbor/config/converters"
)

// Configure adds Harbor registry-related resource configurations
// (registry, replication, retention_policy).
func Configure(p *ujconfig.Provider) {
	// harbor_registry: override registry_id from int to string so the CRD
	// field is *string and ExtractParamPath returns a usable value.
	p.AddResourceConfigurator("harbor_registry", func(r *ujconfig.Resource) {
		converters.OverrideIntFieldAsString(r, "registry_id")
		r.TerraformConversions = append(r.TerraformConversions, converters.IntFieldAsString("registry_id"))
	})

	p.AddResourceConfigurator("harbor_replication", func(r *ujconfig.Resource) {
		converters.OverrideIntFieldAsString(r, "registry_id")
		r.TerraformConversions = append(r.TerraformConversions, converters.IntFieldAsString("registry_id"))
		r.References["registry_id"] = ujconfig.Reference{
			TerraformName: "harbor_registry",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("registry_id",true)`,
		}
	})

	p.AddResourceConfigurator("harbor_retention_policy", func(r *ujconfig.Resource) {
		r.References["scope"] = ujconfig.Reference{
			TerraformName: "harbor_project",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("id",true)`,
		}
	})
}

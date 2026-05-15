// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor registry-related resource configurations
// (registry, replication, retention_policy).
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("harbor_replication", func(r *ujconfig.Resource) {
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

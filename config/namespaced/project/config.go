// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package project

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor project-related resource configurations (project,
// members, webhook, label, immutable_tag_rule, preheat_instance).
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("harbor_immutable_tag_rule", func(r *ujconfig.Resource) {
		r.References["project_id"] = ujconfig.Reference{
			TerraformName: "harbor_project",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("id",true)`,
		}
	})

	p.AddResourceConfigurator("harbor_project_member_group", func(r *ujconfig.Resource) {
		r.References["project_id"] = ujconfig.Reference{
			TerraformName: "harbor_project",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("id",true)`,
		}
	})

	p.AddResourceConfigurator("harbor_project_member_user", func(r *ujconfig.Resource) {
		r.References["project_id"] = ujconfig.Reference{
			TerraformName: "harbor_project",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("id",true)`,
		}
	})

	p.AddResourceConfigurator("harbor_project_webhook", func(r *ujconfig.Resource) {
		r.References["project_id"] = ujconfig.Reference{
			TerraformName: "harbor_project",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("id",true)`,
		}
	})
}

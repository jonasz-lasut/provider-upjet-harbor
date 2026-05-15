// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package project

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor project-related resource configurations (project,
// members, webhook, label, immutable_tag_rule, preheat_instance).
func Configure(p *ujconfig.Provider) {
	// v0: rely on the default external-name behavior set in
	// config/external_name.go. Per-resource customizations
	// (reference injectors, composite IDs) go here incrementally.
	_ = p
}

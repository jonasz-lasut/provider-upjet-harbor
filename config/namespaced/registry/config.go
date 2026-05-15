// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor registry-related resource configurations
// (registry, replication, retention_policy).
func Configure(p *ujconfig.Provider) {
	_ = p
}

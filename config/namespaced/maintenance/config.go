// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package maintenance

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor maintenance-related resource configurations
// (garbage_collection, purge_audit_log, tasks).
func Configure(p *ujconfig.Provider) {
	_ = p
}

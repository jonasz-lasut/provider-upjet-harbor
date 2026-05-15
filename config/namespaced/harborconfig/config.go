// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package harborconfig

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor system-config resource configurations
// (config_auth, config_system, config_security).
func Configure(p *ujconfig.Provider) {
	_ = p
}

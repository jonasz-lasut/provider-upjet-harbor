// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package config

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor system-config resource configurations
// (config_auth, config_security, config_system).
func Configure(p *ujconfig.Provider) {
	_ = p
}

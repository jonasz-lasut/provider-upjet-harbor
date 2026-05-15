// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package vuln

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor vulnerability-related resource configurations
// (interrogation_services).
func Configure(p *ujconfig.Provider) {
	_ = p
}

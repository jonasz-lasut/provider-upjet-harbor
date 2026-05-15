// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds Harbor identity-related resource configurations (user, group).
func Configure(p *ujconfig.Provider) {
	_ = p
}

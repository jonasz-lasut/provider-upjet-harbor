// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/config"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/garbage"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/harbor"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/immutable"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/interrogation"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/preheat"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/project"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/purge"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/retention"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/cluster/robot"
)

func init() {
	ProviderConfiguration.AddConfig(config.Configure)
	ProviderConfiguration.AddConfig(garbage.Configure)
	ProviderConfiguration.AddConfig(harbor.Configure)
	ProviderConfiguration.AddConfig(immutable.Configure)
	ProviderConfiguration.AddConfig(interrogation.Configure)
	ProviderConfiguration.AddConfig(preheat.Configure)
	ProviderConfiguration.AddConfig(project.Configure)
	ProviderConfiguration.AddConfig(purge.Configure)
	ProviderConfiguration.AddConfig(retention.Configure)
	ProviderConfiguration.AddConfig(robot.Configure)
}

// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package namespaced

import (
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/config"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/garbage"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/harbor"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/immutable"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/interrogation"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/preheat"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/project"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/purge"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/retention"
	"github.com/jonasz-lasut/provider-upjet-harbor/config/namespaced/robot"
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

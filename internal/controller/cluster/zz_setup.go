// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	auth "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/config/auth"
	security "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/config/security"
	system "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/config/system"
	collection "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/garbage/collection"
	group "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/harbor/group"
	label "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/harbor/label"
	project "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/harbor/project"
	registry "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/harbor/registry"
	replication "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/harbor/replication"
	tasks "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/harbor/tasks"
	user "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/harbor/user"
	tagrule "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/immutable/tagrule"
	services "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/interrogation/services"
	instance "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/preheat/instance"
	membergroup "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/project/membergroup"
	memberuser "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/project/memberuser"
	webhook "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/project/webhook"
	providerconfig "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/providerconfig"
	auditlog "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/purge/auditlog"
	policy "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/retention/policy"
	account "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/cluster/robot/account"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		auth.Setup,
		security.Setup,
		system.Setup,
		collection.Setup,
		group.Setup,
		label.Setup,
		project.Setup,
		registry.Setup,
		replication.Setup,
		tasks.Setup,
		user.Setup,
		tagrule.Setup,
		services.Setup,
		instance.Setup,
		membergroup.Setup,
		memberuser.Setup,
		webhook.Setup,
		providerconfig.Setup,
		auditlog.Setup,
		policy.Setup,
		account.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		auth.SetupGated,
		security.SetupGated,
		system.SetupGated,
		collection.SetupGated,
		group.SetupGated,
		label.SetupGated,
		project.SetupGated,
		registry.SetupGated,
		replication.SetupGated,
		tasks.SetupGated,
		user.SetupGated,
		tagrule.SetupGated,
		services.SetupGated,
		instance.SetupGated,
		membergroup.SetupGated,
		memberuser.SetupGated,
		webhook.SetupGated,
		providerconfig.SetupGated,
		auditlog.SetupGated,
		policy.SetupGated,
		account.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

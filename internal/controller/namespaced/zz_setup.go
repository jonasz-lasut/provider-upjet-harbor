// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	auth "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/config/auth"
	security "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/config/security"
	system "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/config/system"
	collection "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/garbage/collection"
	group "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/harbor/group"
	label "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/harbor/label"
	project "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/harbor/project"
	registry "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/harbor/registry"
	replication "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/harbor/replication"
	tasks "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/harbor/tasks"
	user "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/harbor/user"
	tagrule "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/immutable/tagrule"
	services "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/interrogation/services"
	instance "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/preheat/instance"
	membergroup "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/project/membergroup"
	memberuser "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/project/memberuser"
	webhook "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/project/webhook"
	providerconfig "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/providerconfig"
	auditlog "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/purge/auditlog"
	policy "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/retention/policy"
	account "github.com/jonasz-lasut/provider-upjet-harbor/internal/controller/namespaced/robot/account"
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

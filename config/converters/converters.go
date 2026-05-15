// SPDX-FileCopyrightText: 2026 jonasz-lasut
// SPDX-License-Identifier: Apache-2.0

package converters

import (
	"fmt"
	"strconv"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// intFieldAsString is a TerraformConversion that represents a Terraform
// integer attribute as a string in the CRD spec. At reconcile time it
// converts string→int64 when sending parameters to Terraform (ToTerraform)
// and int64→string when reading state back from Terraform (FromTerraform).
type intFieldAsString struct {
	field string
}

// IntFieldAsString returns a TerraformConversion that coerces the named
// Terraform int attribute to/from a string in the Crossplane layer.
// Use together with OverrideIntFieldAsString so that the generated CRD type
// is *string rather than *int64.
func IntFieldAsString(field string) ujconfig.TerraformConversion {
	return intFieldAsString{field: field}
}

// OverrideIntFieldAsString mutates the Terraform SDK schema for the named
// attribute from TypeInt to TypeString so that upjet's code generator emits
// *string for the CRD field. This must be called inside an
// AddResourceConfigurator callback, before ConfigureResources is called.
func OverrideIntFieldAsString(r *ujconfig.Resource, field string) {
	s, ok := r.TerraformResource.Schema[field]
	if !ok {
		return
	}
	// Only override genuine integer/float fields.
	if s.Type != schema.TypeInt && s.Type != schema.TypeFloat {
		return
	}
	s.Type = schema.TypeString
}

// Convert implements ujconfig.TerraformConversion.
//
//   - ToTerraform  (K8s → TF): the CRD field is a string; parse it to int64
//     so that Terraform receives the numeric value it expects.
//   - FromTerraform (TF → K8s): TF state carries a float64 (JSON number);
//     format it as a decimal string so that the CRD field stays *string.
func (c intFieldAsString) Convert(params map[string]any, _ *ujconfig.Resource, mode ujconfig.Mode) (map[string]any, error) {
	v, ok := params[c.field]
	if !ok || v == nil {
		return params, nil
	}

	switch mode {
	case ujconfig.ToTerraform:
		// CRD layer keeps registry_id as a string; parse it back to a number
		// for the Terraform provider.
		switch val := v.(type) {
		case string:
			if val == "" {
				return params, nil
			}
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot parse field %q value %q as int64", c.field, val)
			}
			params[c.field] = n
		case float64:
			// Already numeric – pass through unchanged.
		case int64:
			// Already numeric – pass through unchanged.
		default:
			return nil, errors.Errorf("unexpected type %T for field %q in ToTerraform conversion", v, c.field)
		}

	case ujconfig.FromTerraform:
		// Terraform state returns numbers as float64 (JSON unmarshalling).
		// Convert to string so the CRD *string field can hold it.
		switch val := v.(type) {
		case float64:
			params[c.field] = fmt.Sprintf("%d", int64(val))
		case int64:
			params[c.field] = strconv.FormatInt(val, 10)
		case string:
			// Already a string – leave it.
		default:
			return nil, errors.Errorf("unexpected type %T for field %q in FromTerraform conversion", v, c.field)
		}
	}

	return params, nil
}

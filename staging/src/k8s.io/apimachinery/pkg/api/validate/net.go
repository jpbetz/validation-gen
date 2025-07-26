/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validate

import (
	"context"
	"fmt"
	"net/netip"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	netutils "k8s.io/utils/net"
)

// IP verifies that the specified value is a valid IP address.  This does not
// allow leading zeros on octet values.
func IP[T ~string](ctx context.Context, op operation.Operation, fldPath *field.Path, value, _ *T) field.ErrorList {
	vt := *value
	vs := string(vt)
	return ipstr(fldPath, vs)
}

func ipstr(fldPath *field.Path, value string) field.ErrorList {
	var allErrs field.ErrorList

	if addr, err := netip.ParseAddr(value); err != nil {
		if netutils.ParseIPSloppy(value) != nil {
			// If netutils.ParseIPSloppy parses it, but netip.ParseAddr
			// doesn't, then it must have illegal leading 0s.
			allErrs = append(allErrs, field.Invalid(fldPath, value, "must not have leading 0s"))
		} else {
			// Neither strict nor sloppy could parse it.
			allErrs = append(allErrs, field.Invalid(fldPath, value, "must be a valid IP address (e.g. 10.9.8.7 or 2001:db8::ffff)"))
		}
	} else {
		// It parsed as an IP, check for bad forms.
		if addr.String() != value {
			allErrs = append(allErrs, field.Invalid(fldPath, value, fmt.Sprintf("must be in canonical form (%q)", addr.String())))
		}
		if addr.Is4In6() {
			allErrs = append(allErrs, field.Invalid(fldPath, value, "must not be an IPv4-mapped IPv6 address"))
		}
		if addr.Zone() != "" {
			allErrs = append(allErrs, field.Invalid(fldPath, value, "must not include an IPv6 zone"))
		}
	}

	return allErrs.WithOrigin("format=k8s-ip")
}

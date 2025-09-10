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
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestIP(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		err  string
	}{
		// Good values
		{
			name: "ipv4",
			in:   "1.2.3.4",
		}, {
			name: "ipv4, all zeros",
			in:   "0.0.0.0",
		}, {
			name: "ipv4, max",
			in:   "255.255.255.255",
		}, {
			name: "ipv6",
			in:   "1234::abcd",
		}, {
			name: "ipv6, all zeros, collapsed",
			in:   "::",
		}, {
			name: "ipv6, max",
			in:   "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		},

		// Non-canonical values
		{
			name: "ipv6, all zeros, expanded (non-canonical)",
			in:   "0:0:0:0:0:0:0:0",
			err:  "must be in canonical form",
		}, {
			name: "ipv6, leading 0s (non-canonical)",
			in:   "0001:002:03:4::",
			err:  "must be in canonical form",
		}, {
			name: "ipv6, capital letters (non-canonical)",
			in:   "1234::ABCD",
			err:  "must be in canonical form",
		},

		// Questionable values that we choose not to accept
		{
			name: "ipv4 with leading 0s",
			in:   "1.1.1.01",
			err:  "must not have leading 0s",
		}, {
			name: "ipv4-in-ipv6 value",
			in:   "::ffff:1.1.1.1",
			err:  "must not be an IPv4-mapped IPv6 address",
		},

		// Bad values
		{
			name: "empty string",
			in:   "",
			err:  "must be a valid IP address",
		}, {
			name: "junk",
			in:   "aaaaaaa",
			err:  "must be a valid IP address",
		}, {
			name: "domain name",
			in:   "myhost.mydomain",
			err:  "must be a valid IP address",
		}, {
			name: "cidr",
			in:   "1.2.3.0/24",
			err:  "must be a valid IP address",
		}, {
			name: "ipv4 with out-of-range octets",
			in:   "1.2.3.400",
			err:  "must be a valid IP address",
		}, {
			name: "ipv4 with negative octets",
			in:   "-1.0.0.0",
			err:  "must be a valid IP address",
		}, {
			name: "ipv6 with out-of-range segment",
			in:   "2001:db8::10005",
			err:  "must be a valid IP address",
		}, {
			name: "ipv4:port",
			in:   "1.2.3.4:80",
			err:  "must be a valid IP address",
		}, {
			name: "ipv6 with brackets",
			in:   "[2001:db8::1]",
			err:  "must be a valid IP address",
		}, {
			name: "[ipv6]:port",
			in:   "[2001:db8::1]:80",
			err:  "must be a valid IP address",
		}, {
			name: "host:port",
			in:   "example.com:80",
			err:  "must be a valid IP address",
		}, {
			name: "ipv6 with zone",
			in:   "1234::abcd%eth0",
			err:  "must not include an IPv6 zone",
		}, {
			name: "ipv4 with zone",
			in:   "169.254.0.0%eth0",
			err:  "must be a valid IP address",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			errs := ipstr(field.NewPath("ip"), in)
			if tc.err == "" {
				if len(errs) != 0 {
					t.Errorf("%q: expected valid, got: %v", tc.in, errs)
				}
			} else if len(errs) == 0 {
				t.Errorf("%q expected error to contain %q but got no error", tc.in, tc.err)
			} else {
				if len(errs) > 1 {
					t.Errorf("%q: got multiple errors: %v", tc.in, errs)
				} else if !strings.Contains(errs[0].Detail, tc.err) {
					t.Errorf("%q expected error to contain %q but got: %q", tc.in, tc.err, errs[0].Detail)
				}
			}
		})
	}
}

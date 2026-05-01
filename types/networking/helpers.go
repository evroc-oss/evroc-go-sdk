// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import "fmt"

// PublicIPRef is a type-safe reference to a public IP resource.
type PublicIPRef string

// String returns the string representation of the PublicIPRef.
func (r PublicIPRef) String() string {
	return string(r)
}

// SecurityGroupRef is a type-safe reference to a security group resource.
type SecurityGroupRef string

// String returns the string representation of the SecurityGroupRef.
func (r SecurityGroupRef) String() string {
	return string(r)
}

// Ref returns the fully qualified reference as a type-safe PublicIPRef for use with builders.
func (pip *PublicIP) Ref() PublicIPRef {
	if pip.Metadata.Project == nil || pip.Metadata.Region == nil || pip.Metadata.Id == "" {
		return ""
	}
	fqid := fmt.Sprintf("/networking/projects/%s/regions/%s/publicIPs/%s",
		*pip.Metadata.Project, *pip.Metadata.Region, pip.Metadata.Id)
	return PublicIPRef(fqid)
}

// Ref returns the fully qualified reference as a type-safe SecurityGroupRef for use with builders.
func (sg *SecurityGroup) Ref() SecurityGroupRef {
	if sg.Metadata.Project == nil || sg.Metadata.Region == nil || sg.Metadata.Id == "" {
		return ""
	}
	fqid := fmt.Sprintf("/networking/projects/%s/regions/%s/securityGroups/%s",
		*sg.Metadata.Project, *sg.Metadata.Region, sg.Metadata.Id)
	return SecurityGroupRef(fqid)
}

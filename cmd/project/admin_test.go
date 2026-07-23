// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package project

import (
	"reflect"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// TestCollectAdminIDs covers the merge of the flat admin_ids/admin_group_ids
// lists with the expanded admins object (whose identities/groups are objects,
// not strings — SDK #61). Guards the object-field flattening that codecov
// flagged as untested.
func TestCollectAdminIDs(t *testing.T) {
	tests := []struct {
		name       string
		project    *auth.Project
		wantIDs    []string
		wantGroups []string
	}{
		{
			name: "flat lists only",
			project: &auth.Project{
				AdminIDs:      []string{"id-1", "id-2"},
				AdminGroupIDs: []string{"grp-1"},
			},
			wantIDs:    []string{"id-1", "id-2"},
			wantGroups: []string{"grp-1"},
		},
		{
			name: "expanded admins object merged and deduped",
			project: &auth.Project{
				AdminIDs:      []string{"id-1"},
				AdminGroupIDs: []string{"grp-1"},
				Admins: &auth.ProjectAdmins{
					Identities: []auth.ProjectAdminIdentity{
						{ID: "id-1"}, // duplicate of flat list
						{ID: "id-3"},
					},
					Groups: []auth.ProjectAdminGroup{
						{ID: "grp-2", Name: "Admins"},
					},
				},
			},
			wantIDs:    []string{"id-1", "id-3"},
			wantGroups: []string{"grp-1", "grp-2"},
		},
		{
			name: "empty object entries dropped",
			project: &auth.Project{
				AdminIDs: []string{"id-1"},
				Admins: &auth.ProjectAdmins{
					Identities: []auth.ProjectAdminIdentity{{ID: ""}},
					Groups:     []auth.ProjectAdminGroup{{ID: ""}},
				},
			},
			wantIDs:    []string{"id-1"},
			wantGroups: []string{},
		},
		{
			name:       "no admins at all",
			project:    &auth.Project{},
			wantIDs:    []string{},
			wantGroups: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIDs, gotGroups := collectAdminIDs(tt.project)
			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Errorf("identities = %v, want %v", gotIDs, tt.wantIDs)
			}
			if !reflect.DeepEqual(gotGroups, tt.wantGroups) {
				t.Errorf("groups = %v, want %v", gotGroups, tt.wantGroups)
			}
		})
	}
}

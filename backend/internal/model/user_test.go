package model

import "testing"

func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		role  UserRole
		valid bool
	}{
		{RoleAdmin, true},
		{RoleTeamLeader, true},
		{RoleUser, true},
		{"superadmin", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.role.IsValid(); got != tt.valid {
			t.Errorf("UserRole(%q).IsValid() = %v, want %v", tt.role, got, tt.valid)
		}
	}
}

func TestUser_IsAdmin(t *testing.T) {
	admin := &User{GlobalRole: RoleAdmin}
	user := &User{GlobalRole: RoleUser}

	if !admin.IsAdmin() {
		t.Error("expected admin.IsAdmin() == true")
	}
	if user.IsAdmin() {
		t.Error("expected user.IsAdmin() == false")
	}
}

func TestUser_IsTeamLeader(t *testing.T) {
	leader := &User{GlobalRole: RoleTeamLeader}
	user := &User{GlobalRole: RoleUser}

	if !leader.IsTeamLeader() {
		t.Error("expected leader.IsTeamLeader() == true")
	}
	if user.IsTeamLeader() {
		t.Error("expected user.IsTeamLeader() == false")
	}
}

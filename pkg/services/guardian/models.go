package guardian

import (
	"fmt"

	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
)

type DashboardGuardian struct {
	user      *m.SignedInUser
	dashboard *m.Dashboard
	acl       []*m.DashboardAclInfoDTO
	groups    []*m.UserGroup
}

func NewDashboardGuardian(dash *m.Dashboard, user *m.SignedInUser) *DashboardGuardian {
	return &DashboardGuardian{
		user:      user,
		dashboard: dash,
	}
}

func (g *DashboardGuardian) CanSave() (bool, error) {
	fmt.Printf("user %v, %v", g.user.OrgRole, g.user.HasRole(m.ROLE_EDITOR))
	if !g.dashboard.HasAcl {
		return g.user.HasRole(m.ROLE_EDITOR), nil
	}

	return g.HasPermission(m.PERMISSION_EDIT)
}

func (g *DashboardGuardian) CanEdit() (bool, error) {
	if !g.dashboard.HasAcl {
		return g.user.HasRole(m.ROLE_READ_ONLY_EDITOR), nil
	}

	return g.HasPermission(m.PERMISSION_READ_ONLY_EDIT)
}

func (g *DashboardGuardian) CanView() (bool, error) {
	if !g.dashboard.HasAcl {
		return g.user.HasRole(m.ROLE_VIEWER), nil
	}

	return g.HasPermission(m.PERMISSION_VIEW)
}

func (g *DashboardGuardian) HasPermission(permission m.PermissionType) (bool, error) {
	userGroups, err := g.getUserGroups()
	if err != nil {
		return false, err
	}

	acl, err := g.getAcl()
	if err != nil {
		return false, err
	}

	for _, p := range acl {
		if p.UserId == g.user.UserId && p.Permissions >= permission {
			return true, nil
		}

		for _, ug := range userGroups {
			if ug.Id == p.UserGroupId && p.Permissions >= permission {
				return true, nil
			}
		}
	}

	return false, nil
}

// Returns dashboard acl
func (g *DashboardGuardian) getAcl() ([]*m.DashboardAclInfoDTO, error) {
	if g.acl != nil {
		return g.acl, nil
	}

	dashId := g.dashboard.Id
	if g.dashboard.ParentId != 0 {
		dashId = g.dashboard.ParentId
	}

	query := m.GetDashboardPermissionsQuery{DashboardId: dashId}
	if err := bus.Dispatch(&query); err != nil {
		return nil, err
	}

	g.acl = query.Result
	return g.acl, nil
}

func (g *DashboardGuardian) getUserGroups() ([]*m.UserGroup, error) {
	if g.groups == nil {
		return g.groups, nil
	}

	query := m.GetUserGroupsByUserQuery{UserId: g.user.UserId}
	err := bus.Dispatch(&query)

	g.groups = query.Result
	return query.Result, err
}

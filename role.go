package disgord

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/andersfylling/disgord/internal/endpoint"
	"github.com/andersfylling/disgord/internal/httd"
)

type roles []*Role

var _ sort.Interface = (roles)(nil)

func (r roles) Len() int {
	return len(r)
}

// Less is reversed due to the visual ordering in Discord.
func (r roles) Less(i, j int) bool {
	a := r[i]
	b := r[j]

	if a.Position == b.Position {
		return a.ID < b.ID
	}

	return a.Position > b.Position
}

func (r roles) Swap(i, j int) {
	tmp := r[i]
	r[i] = r[j]
	r[j] = tmp
}

// SortRoles sorts a slice of roles such that the first element is the top one in the Discord Guild Settings UI.
func SortRoles(rs []*Role) {
	sort.Sort(roles(rs))
}

// NewRole ...
func NewRole() *Role {
	return &Role{}
}

// Role https://discord.com/developers/docs/topics/permissions#role-object
type Role struct {
	ID          Snowflake `json:"id"`
	Name        string    `json:"name"`
	Color       uint      `json:"color"`
	Hoist       bool      `json:"hoist"`
	Position    int       `json:"position"` // can be -1
	Permissions uint64    `json:"permissions"`
	Managed     bool      `json:"managed"`
	Mentionable bool      `json:"mentionable"`

	guildID Snowflake
}

var _ Mentioner = (*Role)(nil)
var _ Reseter = (*Role)(nil)
var _ DeepCopier = (*Role)(nil)
var _ Copier = (*Role)(nil)
var _ discordDeleter = (*Role)(nil)
var _ fmt.Stringer = (*Role)(nil)

func (r *Role) String() string {
	return r.Name
}

// Mention gives a formatted version of the role such that it can be parsed by Discord clients
func (r *Role) Mention() string {
	return "<@&" + r.ID.String() + ">"
}

// SetGuildID link role to a guild before running session.SaveToDiscord(*Role)
func (r *Role) SetGuildID(id Snowflake) {
	r.guildID = id
}

// DeepCopy see interface at struct.go#DeepCopier
func (r *Role) DeepCopy() (copy interface{}) {
	copy = NewRole()
	r.CopyOverTo(copy)

	return
}

// CopyOverTo see interface at struct.go#Copier
func (r *Role) CopyOverTo(other interface{}) (err error) {
	var ok bool
	var role *Role
	if role, ok = other.(*Role); !ok {
		return newErrorUnsupportedType("given interface{} was not a *Role")
	}

	role.ID = r.ID
	role.Name = r.Name
	role.Color = r.Color
	role.Hoist = r.Hoist
	role.Position = r.Position
	role.Permissions = r.Permissions
	role.Managed = r.Managed
	role.Mentionable = r.Mentionable
	role.guildID = r.guildID
	return
}

func (r *Role) deleteFromDiscord(ctx context.Context, s Session, flags ...Flag) (err error) {
	guildID := r.guildID
	id := r.ID

	if id.IsZero() {
		return newErrorMissingSnowflake("role has no ID")
	}
	if guildID.IsZero() {
		return newErrorMissingSnowflake("role has no guildID")
	}

	err = s.DeleteGuildRole(ctx, guildID, id, flags...)
	return err
}

//////////////////////////////////////////////////////
//
// REST Methods
//
//////////////////////////////////////////////////////

// CreateGuildRoleParams ...
// https://discord.com/developers/docs/resources/guild#create-guild-role-json-params
type CreateGuildRoleParams struct {
	Name        string `json:"name,omitempty"`
	Permissions uint64 `json:"permissions,omitempty"`
	Color       uint   `json:"color,omitempty"`
	Hoist       bool   `json:"hoist,omitempty"`
	Mentionable bool   `json:"mentionable,omitempty"`

	// Reason is a X-Audit-Log-Reason header field that will show up on the audit log for this action.
	Reason string `json:"-"`
}

// CreateGuildRole [REST] Create a new role for the guild. Requires the 'MANAGE_ROLES' permission.
// Returns the new role object on success. Fires a Guild Role Create Gateway event.
//  Method                  POST
//  Endpoint                /guilds/{guild.id}/roles
//  Discord documentation   https://discord.com/developers/docs/resources/guild#create-guild-role
//  Reviewed                2018-08-18
//  Comment                 All JSON params are optional.
func (c *Client) CreateGuildRole(ctx context.Context, id Snowflake, params *CreateGuildRoleParams, flags ...Flag) (ret *Role, err error) {
	r := c.newRESTRequest(&httd.Request{
		Method:      httd.MethodPost,
		Ctx:         ctx,
		Endpoint:    endpoint.GuildRoles(id),
		Body:        params,
		ContentType: httd.ContentTypeJSON,
		Reason:      params.Reason,
	}, flags)
	r.CacheRegistry = GuildRoleCache
	r.factory = func() interface{} {
		return &Role{}
	}
	r.preUpdateCache = func(x interface{}) {
		r := x.(*Role)
		r.guildID = id
	}

	return getRole(r.Execute)
}

// ModifyGuildRole [REST] Modify a guild role. Requires the 'MANAGE_ROLES' permission.
// Returns the updated role on success. Fires a Guild Role Update Gateway event.
//  Method                  PATCH
//  Endpoint                /guilds/{guild.id}/roles/{role.id}
//  Discord documentation   https://discord.com/developers/docs/resources/guild#modify-guild-role
//  Reviewed                2018-08-18
//  Comment                 -
func (c *Client) UpdateGuildRole(ctx context.Context, guildID, roleID Snowflake, flags ...Flag) (builder *updateGuildRoleBuilder) {
	builder = &updateGuildRoleBuilder{}
	builder.r.itemFactory = func() interface{} {
		return &Role{}
	}
	builder.r.flags = flags
	builder.r.IgnoreCache().setup(c.cache, c.req, &httd.Request{
		Method:      httd.MethodPatch,
		Ctx:         ctx,
		Endpoint:    endpoint.GuildRole(guildID, roleID),
		ContentType: httd.ContentTypeJSON,
	}, nil)

	builder.r.cacheMiddleware = func(resp *http.Response, v interface{}, err error) error {
		role := v.(*Role)
		role.guildID = guildID
		return nil
	}

	return builder
}

// DeleteGuildRole [REST] Delete a guild role. Requires the 'MANAGE_ROLES' permission.
// Returns a 204 empty response on success. Fires a Guild Role Delete Gateway event.
//  Method                  DELETE
//  Endpoint                /guilds/{guild.id}/roles/{role.id}
//  Discord documentation   https://discord.com/developers/docs/resources/guild#delete-guild-role
//  Reviewed                2018-08-18
//  Comment                 -
func (c *Client) DeleteGuildRole(ctx context.Context, guildID, roleID Snowflake, flags ...Flag) (err error) {
	r := c.newRESTRequest(&httd.Request{
		Method:   httd.MethodDelete,
		Endpoint: endpoint.GuildRole(guildID, roleID),
		Ctx:      ctx,
	}, flags)
	r.expectsStatusCode = http.StatusNoContent

	_, err = r.Execute()
	return err
}

// GetGuildRoles [REST] Returns a list of role objects for the guild.
//  Method                  GET
//  Endpoint                /guilds/{guild.id}/roles
//  Discord documentation   https://discord.com/developers/docs/resources/guild#get-guild-roles
//  Reviewed                2018-08-18
//  Comment                 -
func (c *Client) GetGuildRoles(ctx context.Context, guildID Snowflake, flags ...Flag) (ret []*Role, err error) {
	r := c.newRESTRequest(&httd.Request{
		Endpoint: "/guilds/" + guildID.String() + "/roles",
		Ctx:      ctx,
	}, flags)
	r.CacheRegistry = GuildRolesCache
	r.factory = func() interface{} {
		tmp := make([]*Role, 0)
		return &tmp
	}
	r.preUpdateCache = func(x interface{}) {
		roles := *x.(*[]*Role)
		for i := range roles {
			roles[i].guildID = guildID
		}
	}

	return getRoles(r.Execute)
}

// GetMemberPermissions populates a uint64 with all the permission flags
func (c *Client) GetMemberPermissions(ctx context.Context, guildID, userID Snowflake, flags ...Flag) (permissions PermissionBits, err error) {
	roles, err := c.GetGuildRoles(ctx, guildID, flags...)
	if err != nil {
		return 0, err
	}

	member, err := c.GetMember(ctx, guildID, userID, flags...)
	if err != nil {
		return 0, err
	}

	roleIDs := member.Roles
	for i := range roles {
		for j := range roleIDs {
			if roles[i].ID == roleIDs[j] {
				permissions |= roles[i].Permissions
				roleIDs = roleIDs[:j+copy(roleIDs[j:], roleIDs[j+1:])]
				break
			}
		}

		if len(roleIDs) == 0 {
			break
		}
	}

	return permissions, nil
}

//////////////////////////////////////////////////////
//
// REST Builders
//
//////////////////////////////////////////////////////

// updateGuildRoleBuilder ...
//generate-rest-basic-execute: role:*Role,
//generate-rest-params: name:string, permissions:PermissionBits, color:uint, hoist:bool, mentionable:bool,
type updateGuildRoleBuilder struct {
	r RESTBuilder
}

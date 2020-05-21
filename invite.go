package disgord

import (
	"context"

	"github.com/andersfylling/disgord/internal/endpoint"
	"github.com/andersfylling/disgord/internal/httd"
)

// PartialInvite ...
// {
//    "code": "abc"
// }
type PartialInvite = Invite

// Invite Represents a code that when used, adds a user to a guild.
// https://discord.com/developers/docs/resources/invite#invite-object
// Reviewed: 2018-06-10
type Invite struct {
	// Code the invite code (unique Snowflake)
	Code string `json:"code"`

	// Guild the guild this invite is for
	Guild *PartialGuild `json:"guild"`

	// Channel the channel this invite is for
	Channel *PartialChannel `json:"channel"`

	// Inviter the user that created the invite
	Inviter *User `json:"inviter"`

	// CreatedAt the time at which the invite was created
	CreatedAt Time `json:"created_at"`

	// MaxAge how long the invite is valid for (in seconds)
	MaxAge int `json:"max_age"`

	// MaxUses the maximum number of times the invite can be used
	MaxUses int `json:"max_uses"`

	// Temporary whether or not the invite is temporary (invited users will be kicked on disconnect unless they're assigned a role)
	Temporary bool `json:"temporary"`

	// Uses how many times the invite has been used (always will be 0)
	Uses int `json:"uses"`

	Revoked bool `json:"revoked"`
	Unique  bool `json:"unique"`

	// ApproximatePresenceCount approximate count of online members
	ApproximatePresenceCount int `json:"approximate_presence_count,omitempty"`

	// ApproximatePresenceCount approximate count of total members
	ApproximateMemberCount int `json:"approximate_member_count,omitempty"`
}

var _ Copier = (*Invite)(nil)
var _ DeepCopier = (*Invite)(nil)
var _ discordDeleter = (*Invite)(nil)

func (i *Invite) deleteFromDiscord(ctx context.Context, s Session, flags ...Flag) error {
	if i.Code == "" {
		return &ErrorEmptyValue{info: "can not delete invite without the code field populate"}
	}

	_, err := s.DeleteInvite(ctx, i.Code, flags...)
	return err
}

// DeepCopy see interface at struct.go#DeepCopier
func (i *Invite) DeepCopy() (copy interface{}) {
	copy = &Invite{}
	i.CopyOverTo(copy)

	return
}

// CopyOverTo see interface at struct.go#Copier
func (i *Invite) CopyOverTo(other interface{}) (err error) {
	var ok bool
	var invite *Invite
	if invite, ok = other.(*Invite); !ok {
		err = newErrorUnsupportedType("given interface{} was not of type *Invite")
		return
	}

	invite.Code = i.Code
	invite.ApproximatePresenceCount = i.ApproximatePresenceCount
	invite.ApproximateMemberCount = i.ApproximateMemberCount

	if i.Guild != nil {
		invite.Guild = NewPartialGuild(i.Guild.ID)
	}
	if i.Channel != nil {
		c := i.Channel
		invite.Channel = &PartialChannel{
			ID:   c.ID,
			Name: c.Name,
			Type: c.Type,
		}
	}

	return nil
}

// InviteMetadata Object
// https://discord.com/developers/docs/resources/invite#invite-metadata-object
// Reviewed: 2018-06-10
type InviteMetadata struct {
	// Inviter user who created the invite
	Inviter *User `json:"inviter"`

	// Uses number of times this invite has been used
	Uses int `json:"uses"`

	// MaxUses max number of times this invite can be used
	MaxUses int `json:"max_uses"`

	// MaxAge duration (in seconds) after which the invite expires
	MaxAge int `json:"max_age"`

	// Temporary whether this invite only grants temporary membership
	Temporary bool `json:"temporary"`

	// CreatedAt when this invite was created
	CreatedAt Time `json:"created_at"`

	// Revoked whether this invite is revoked
	Revoked bool `json:"revoked"`
}

var _ Copier = (*InviteMetadata)(nil)
var _ DeepCopier = (*InviteMetadata)(nil)

// DeepCopy see interface at struct.go#DeepCopier
func (i *InviteMetadata) DeepCopy() (copy interface{}) {
	copy = &InviteMetadata{}
	i.CopyOverTo(copy)

	return
}

// CopyOverTo see interface at struct.go#Copier
func (i *InviteMetadata) CopyOverTo(other interface{}) (err error) {
	var ok bool
	var invite *InviteMetadata
	if invite, ok = other.(*InviteMetadata); !ok {
		err = newErrorUnsupportedType("given interface{} was not of type *InviteMetadata")
		return
	}

	invite.Uses = i.Uses
	invite.MaxUses = i.MaxUses
	invite.MaxAge = i.MaxAge
	invite.Temporary = i.Temporary
	invite.CreatedAt = i.CreatedAt
	invite.Revoked = i.Revoked

	if i.Inviter != nil {
		invite.Inviter = i.Inviter.DeepCopy().(*User)
	}
	return nil
}

// voiceRegionsFactory temporary until flyweight is implemented
func inviteFactory() interface{} {
	return &Invite{}
}

type GetInviteParams struct {
	WithMemberCount bool `urlparam:"with_count,omitempty"`
}

var _ URLQueryStringer = (*GetInviteParams)(nil)

// GetInvite [REST] Returns an invite object for the given code.
//  Method                  GET
//  Endpoint                /invites/{invite.code}
//  Discord documentation   https://discord.com/developers/docs/resources/invite#get-invite
//  Reviewed                2018-06-10
//  Comment                 -
//  withMemberCount: whether or not the invite should contain the approximate number of members
func (c *Client) GetInvite(ctx context.Context, inviteCode string, params URLQueryStringer, flags ...Flag) (invite *Invite, err error) {
	if params == nil {
		params = &GetInviteParams{}
	}

	r := c.newRESTRequest(&httd.Request{
		Endpoint: endpoint.Invite(inviteCode) + params.URLQueryString(),
		Ctx:      ctx,
	}, flags)
	r.factory = inviteFactory

	return getInvite(r.Execute)
}

// DeleteInvite [REST] Delete an invite. Requires the MANAGE_CHANNELS permission. Returns an invite object on success.
//  Method                  DELETE
//  Endpoint                /invites/{invite.code}
//  Discord documentation   https://discord.com/developers/docs/resources/invite#delete-invite
//  Reviewed                2018-06-10
//  Comment                 -
func (c *Client) DeleteInvite(ctx context.Context, inviteCode string, flags ...Flag) (deleted *Invite, err error) {
	r := c.newRESTRequest(&httd.Request{
		Method:   httd.MethodDelete,
		Endpoint: endpoint.Invite(inviteCode),
		Ctx:      ctx,
	}, flags)
	r.factory = inviteFactory

	return getInvite(r.Execute)
}

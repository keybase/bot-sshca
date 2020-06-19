package shared

import (
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/keybase1"
)

// CanRoleReadTeam checks if given role can access team data. User has to be an
// actual team member (not implicit admin) to be able to access data, and
// cannot be a RESTRICTED BOT.
func CanRoleReadTeam(role keybase1.TeamRole) bool {
	switch role {
	case keybase1.TeamRole_READER,
		keybase1.TeamRole_WRITER,
		keybase1.TeamRole_ADMIN,
		keybase1.TeamRole_OWNER,
		keybase1.TeamRole_BOT:
		return true
	default:
		return false
	}
}

// GetAllTeams makes an API call and returns list of team names readable for
// current user.
func GetAllTeams(api *kbchat.API) (teams []string, err error) {
	memberships, err := api.ListUserMemberships(api.GetUsername())
	if err != nil {
		return teams, err
	}
	for _, m := range memberships {
		if CanRoleReadTeam(m.Role) {
			teams = append(teams, m.FqName)
		}
	}
	return teams, nil
}

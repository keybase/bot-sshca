package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildAnnouncement(t *testing.T) {
	values := AnnouncementTemplateValues{Username: "my_username", CurrentTeam: "my_cur_team", Teams: []string{"my_team1", "my_team2"}}

	require.Equal(t, "",
		buildAnnouncement("", values))
	require.Equal(t, "no templates",
		buildAnnouncement("no templates", values))
	require.Equal(t, "repeated my_username my_username my_username",
		buildAnnouncement("repeated {USERNAME} {USERNAME} {USERNAME}", values))
	require.Equal(t, "all my_username my_cur_team my_team1, my_team2",
		buildAnnouncement("all {USERNAME} {CURRENT_TEAM} {TEAMS}", values))
	require.Equal(t, "bogus {FOO}",
		buildAnnouncement("bogus {FOO}", values))
	require.Equal(t, "double-is-not-escape {my_username}",
		buildAnnouncement("double-is-not-escape {{USERNAME}}", values))
}

package bot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildAnnouncement(t *testing.T) {
	require.Equal(t, "",
		buildAnnouncement("", "my_username", "my_cur_team", []string{"my_team1", "my_team2"}))
	require.Equal(t, "no templates",
		buildAnnouncement("no templates", "my_username", "my_cur_team", []string{"my_team1", "my_team2"}))
	require.Equal(t, "repeated my_username my_username my_username",
		buildAnnouncement("repeated $USERNAME $USERNAME $USERNAME", "my_username", "my_cur_team", []string{"my_team1", "my_team2"}))
	require.Equal(t, "all my_username my_cur_team my_team1, my_team2",
		buildAnnouncement("all $USERNAME $CURRENT_TEAM $TEAMS", "my_username", "my_cur_team", []string{"my_team1", "my_team2"}))
	require.Equal(t, "bogus $FOO",
		buildAnnouncement("bogus $FOO", "my_username", "my_cur_team", []string{"my_team1", "my_team2"}))
	require.Equal(t, "double-is-not-escape $my_username",
		buildAnnouncement("double-is-not-escape $$USERNAME", "my_username", "my_cur_team", []string{"my_team1", "my_team2"}))
}

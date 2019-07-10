package bot

import (
	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"strings"
)

func GetMembers(conf config.Config, team string) ([]string, error) {
	runOptions := kbchat.RunOptions{}
	if conf.GetUseAlternateAccount() {
		runOptions = kbchat.RunOptions{HomeDir: conf.GetKeybaseHomeDir(), Oneshot: &kbchat.OneshotOptions{PaperKey: conf.GetKeybasePaperKey(), Username: conf.GetKeybaseUsername()}}
	}
	_, err := kbchat.Start(runOptions)
	if err != nil {
		return nil, err
	}
	cmd := runOptions.Command("team", "list-members", team)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var users []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "writer") {
			users = append(users, strings.Split(line, " ")[1])
		}
	}
	return users, nil
}

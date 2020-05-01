package kssh

import (
	"fmt"
	"strings"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

type Bot struct {
	api *kbchat.API
}

// NewBot creates a new Keybase chat API
func NewBot() (bot Bot, err error) {
	runOptions := kbchat.RunOptions{KeybaseLocation: GetKeybaseBinaryPath()}
	api, err := kbchat.Start(runOptions)
	if err != nil {
		return bot, fmt.Errorf("error starting Keybase chat: %v", err)
	}
	return Bot{api: api}, nil
}

// Configure takes a Bot instance and returns a ConfiguredBot, that includes
// the kssh Config it found for the botName, if specified
func (b *Bot) Configure(botName string) (cbot ConfiguredBot, err error) {
	conf, err := b.getConfig(botName)
	if err != nil {
		return cbot, fmt.Errorf("failed to configure bot: %+v\n", err)
	}
	return ConfiguredBot{conf: conf, api: b.api}, nil
}

// Get the kssh config from the KV store. botName is the bot specified via
// --bot, else is an empty string
func (b *Bot) getConfig(botName string) (conf Config, err error) {
	empty := Config{}
	// They specified a bot via `kssh --bot cabot ...`
	if botName != "" {
		fmt.Printf(">>>BOTNAME= %+v\n", botName)
		conf, err = b.LoadConfigForBot(botName)
		if err != nil {
			return empty, fmt.Errorf("Failed to load config file for bot=%s: %v", botName, err)
		}
		return conf, nil
	}

	// Check if they set a default bot and retrieve the config for that bot/team if so
	defaultBot, defaultTeam, err := GetDefaultBotAndTeam()
	if err != nil {
		return empty, err
	}
	if defaultBot != "" && defaultTeam != "" {
		fmt.Printf(">>>DEFUALT BOTJjllwww= %+v\n", defaultBot)
		conf, err := b.LoadConfig(defaultTeam)
		if err != nil || conf == nil {
			return empty, fmt.Errorf("Failed to load config file for default bot=%s, team=%s: %v", defaultBot, defaultTeam, err)
		}
		return *conf, nil
	}

	// No specified bot and no default bot, fallback and load all the configs
	configs, botNames, err := b.LoadConfigs()
	if err != nil {
		return empty, fmt.Errorf("Failed to load config(s): %v", err)
	}
	fmt.Printf(">>>>>>>>>>>>>>>>>>>> LOADED CONFIGS = %+v\n", configs)
	switch len(configs) {
	case 0:
		return empty, fmt.Errorf("Did not find any configs (is `keybaseca service` running?)")
	case 1:
		return configs[0], nil
	default:
		noDefaultTeamMessage := fmt.Sprintf("Found %d config files (%s). No default bot is configured. \n"+
			"Either specify a team via `kssh --bot cabotName` or set a default bot via `kssh --set-default-bot cabotName`", len(configs), strings.Join(botNames, ", "))
		return empty, fmt.Errorf(noDefaultTeamMessage)
	}
}

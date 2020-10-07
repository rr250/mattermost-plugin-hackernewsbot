package main

import (
	"time"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	botUserName    = "hackernews_bot"
	botDisplayName = "Hacker News Bot"
)

// OnActivate register the plugin command
func (p *Plugin) OnActivate() error {
	botUserID, err := p.ensureBotExists()
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot user")
	}
	p.botUserID = botUserID
	job1, cronErr1 := cluster.Schedule(
		p.API,
		"BackgroundJob",
		cluster.MakeWaitForRoundedInterval(24*time.Hour),
		p.SendDailyTechUpdates,
	)
	if cronErr1 != nil {
		return errors.Wrap(cronErr1, "failed to schedule background job")
	}
	p.backgroundJob1 = job1
	job2, cronErr2 := cluster.Schedule(
		p.API,
		"BackgroundJob",
		cluster.MakeWaitForInterval(4*time.Hour),
		p.SendTechUpdate,
	)
	if cronErr2 != nil {
		return errors.Wrap(cronErr2, "failed to schedule background job")
	}
	p.backgroundJob2 = job2
	return nil
}

func (p *Plugin) ensureBotExists() (string, error) {
	bot := &model.Bot{
		Username:    botUserName,
		DisplayName: botDisplayName,
	}

	return p.Helpers.EnsureBot(bot)
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/robfig/cron"
)

func (p *Plugin) InitCRON() *cron.Cron {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	c := cron.NewWithLocation(loc)
	c.AddFunc("@midnight", p.SendDailyTechUpdates)
	c.AddFunc("@every 240m", p.SendTechUpdate)
	return c
}

func (p *Plugin) SendDailyTechUpdates() {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://hacker-news.firebaseio.com/v0/topstories.json", nil)
	res, err := client.Do(req)

	if err != nil {
		p.API.LogError(err.Error())
		return
	}
	var topNewsIDList []interface{}
	err = json.NewDecoder(res.Body).Decode(&topNewsIDList)
	if err != nil {
		p.API.LogError(err.Error())
		return
	}
	res.Body.Close()
	attachments := []*model.SlackAttachment{}
	j := 0
	for i, topNewsID := range topNewsIDList {
		if i+j == 7 {
			break
		}
		req1, _ := http.NewRequest("GET", "https://hacker-news.firebaseio.com/v0/item/"+fmt.Sprintf("%d", int(topNewsID.(float64)))+".json", nil)
		res1, err1 := client.Do(req1)

		if err1 != nil {
			p.API.LogError(err1.Error())
			j = j - 1
			continue
		}
		var body map[string]interface{}
		err1 = json.NewDecoder(res1.Body).Decode(&body)
		if err1 != nil {
			p.API.LogError(err1.Error())
			j = j - 1
			continue
		}
		res1.Body.Close()

		_, ok := body["url"]
		if !ok {
			j = j - 1
			continue
		}

		req2, _ := http.NewRequest("GET", "https://api.urlmeta.org/?url="+body["url"].(string), nil)
		req2.Header.Set("Authorization", "Basic cnJyaXNoYWJoN0BnbWFpbC5jb206dmQ0aHFvZGl0VDFCZ28xenFIRFI=")
		res2, err1 := client.Do(req2)

		if err1 != nil {
			p.API.LogError(err1.Error())
			j = j - 1
			continue
		}
		var techUpdates map[string]interface{}
		err1 = json.NewDecoder(res2.Body).Decode(&techUpdates)
		if err1 != nil {
			p.API.LogError(err1.Error())
			j = j - 1
			continue
		}
		res2.Body.Close()

		_, ok = techUpdates["meta"]
		if !ok {
			j = j - 1
			continue
		}
		meta := techUpdates["meta"].(map[string]interface{})
		_, ok1 := meta["description"]
		_, ok2 := meta["image"]
		if !ok1 || !ok2 {
			j = j - 1
			continue
		}

		attachment := &model.SlackAttachment{}
		attachment.Title = body["title"].(string)
		attachment.TitleLink = body["url"].(string)
		attachment.Text = meta["description"].(string)
		attachment.ImageURL = meta["image"].(string)
		attachments = append(attachments, attachment)
	}
	configuration := p.getConfiguration()
	for _, channelID := range strings.Split(strings.Trim(configuration.ChannelIDList, " "), ",") {
		postModel := &model.Post{
			UserId:    p.botUserID,
			ChannelId: channelID,
			Message:   "For more news visit https://news.ycombinator.com/",
			Props: model.StringInterface{
				"attachments": attachments,
			},
		}
		_, err = p.API.CreatePost(postModel)
		if err != nil {
			p.API.LogError(err.Error())
		}
	}
}

func (p *Plugin) SendTechUpdate() {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://hacker-news.firebaseio.com/v0/topstories.json", nil)
	res, err := client.Do(req)

	if err != nil {
		p.API.LogError(err.Error())
		return
	}
	var topNewsIDList []interface{}
	err = json.NewDecoder(res.Body).Decode(&topNewsIDList)
	if err != nil {
		p.API.LogError(err.Error())
		return
	}
	res.Body.Close()
	attachments := []*model.SlackAttachment{}
	for _, topNewsID := range topNewsIDList {
		req1, _ := http.NewRequest("GET", "https://hacker-news.firebaseio.com/v0/item/"+fmt.Sprintf("%d", int(topNewsID.(float64)))+".json", nil)
		res1, err1 := client.Do(req1)

		if err1 != nil {
			p.API.LogError(err1.Error())
			continue
		}
		var body map[string]interface{}
		err1 = json.NewDecoder(res1.Body).Decode(&body)
		if err1 != nil {
			p.API.LogError(err1.Error())
			continue
		}
		res1.Body.Close()
		loc, _ := time.LoadLocation("Asia/Kolkata")
		createAt := time.Unix(int64(body["time"].(float64)), 0).In(loc)
		now := time.Now().In(loc).Add(time.Duration(-240) * time.Minute)

		if createAt.Before(now) {
			continue
		}

		_, ok := body["url"]
		if !ok {
			continue
		}

		req2, _ := http.NewRequest("GET", "https://api.urlmeta.org/?url="+body["url"].(string), nil)
		req2.Header.Set("Authorization", "Basic cnJyaXNoYWJoN0BnbWFpbC5jb206dmQ0aHFvZGl0VDFCZ28xenFIRFI=")
		res2, err1 := client.Do(req2)

		if err1 != nil {
			p.API.LogError(err1.Error())
			continue
		}
		var techUpdates map[string]interface{}
		err1 = json.NewDecoder(res2.Body).Decode(&techUpdates)
		if err1 != nil {
			p.API.LogError(err1.Error())
			continue
		}
		res2.Body.Close()

		_, ok = techUpdates["meta"]
		if !ok {
			continue
		}
		meta := techUpdates["meta"].(map[string]interface{})
		_, ok1 := meta["description"]
		_, ok2 := meta["image"]
		if !ok1 || !ok2 {
			continue
		}

		attachment := &model.SlackAttachment{}
		attachment.Title = body["title"].(string)
		attachment.TitleLink = body["url"].(string)
		attachment.Text = meta["description"].(string)
		attachment.ImageURL = meta["image"].(string)
		attachments = append(attachments, attachment)
		break
	}
	configuration := p.getConfiguration()
	for _, channelID := range strings.Split(strings.Trim(configuration.ChannelIDList, " "), ",") {
		postModel := &model.Post{
			UserId:    p.botUserID,
			ChannelId: channelID,
			Message:   "For more news visit https://news.ycombinator.com/",
			Props: model.StringInterface{
				"attachments": attachments,
			},
		}
		_, err = p.API.CreatePost(postModel)
		if err != nil {
			p.API.LogError(err.Error())
		}
	}
}

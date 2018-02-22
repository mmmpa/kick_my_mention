package main

import (
	"os"
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"bytes"
	"time"
	"github.com/aws/aws-lambda-go/lambda"
)

type Notifications struct {
	Items []Notification
}

type Notification struct {
	Reason string `json:"reason"`
	Subject struct {
		LatestCommentURL string `json:"latest_comment_url"`
		Title            string `json:"title"`
		Type             string `json:"type"`
		URL              string `json:"url"`
	} `json:"subject"`
}

type Mention struct {
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}

type Payload struct {
	Text string `json:"text"`
}

func main() {
	lambda.Start(Handler)
}

func Handler() {
	now := time.Now().UTC()

	api := fmt.Sprintf(
		"https://api.github.com/notifications?participating=true&since=%s&before=%s",
		now.Add(time.Duration(-1) * time.Hour).Format("2006-01-02T15")+":00:00Z",
		now.Format("2006-01-02T15")+":00:00Z",
	)
	token := os.Getenv("KICK_MY_MENTION_TOKEN")
	hook := os.Getenv("KICK_MY_MENTION_SLACK_HOOK")

	err, notifications := fetchNotifications(api, token)

	if err != nil {
		fmt.Errorf("%v", err)
		return
	}

	for _, notification := range notifications {
		err, mention := fetchMention(notification.Subject.LatestCommentURL)
		if err == nil {
			err = postMessage(hook, Payload{
				Text: fmt.Sprintf(
					"from: %v, message: %v, url: %v",
					mention.User.Login,
					mention.Body,
					notification.Subject.LatestCommentURL,
				),
			})
			fmt.Errorf("%v", err)
		} else {
			fmt.Errorf("%v", err)
		}
	}
}

func postMessage(url string, payload Payload) error {
	payloadString, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(payloadString),
	)

	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return err
}

func fetchNotifications(url, token string) (error, []Notification) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := new(http.Client)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return err, nil
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err, nil
	}

	var notifications []Notification
	err = json.Unmarshal(b, &notifications)

	if err != nil {
		return err, nil
	}

	return nil, notifications
}

func fetchMention(url string) (error, Mention) {
	var mention Mention

	req, _ := http.NewRequest("GET", url, nil)
	client := new(http.Client)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return err, mention
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err, mention
	}

	err = json.Unmarshal(b, &mention)

	return err, mention
}

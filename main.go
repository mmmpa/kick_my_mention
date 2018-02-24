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
	"log"
)

type Attachment struct {
	Color      string `json:"color"`
	AuthorName string `json:"author_name"`
	AuthorLink string `json:"author_link"`
	Title      string `json:"title"`
	TitleLink  string `json:"title_link"`
	Text       string `json:"text"`
	ThumbURL   string `json:"thumb_url"`
}

type Mention struct {
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	User struct {
		Login     string `json:"login"`
		URL       string `json:"url"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

type Notification struct {
	Reason string `json:"reason"`
	Subject struct {
		LatestCommentURL string `json:"latest_comment_url"`
	} `json:"subject"`
}

func (n Notification) IsMention() bool {
	return n.Reason == "mention"
}

type Payload struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

func main() {
	if len(os.Getenv("LOCAL_RUN")) > 0 {
		log.Println(execute())
	} else {
		lambda.Start(execute)
	}
}

func execute() (string, error) {
	api := "https://api.github.com/notifications?participating=true"
	token := os.Getenv("KICK_MY_MENTION_TOKEN")
	hook := os.Getenv("KICK_MY_MENTION_SLACK_HOOK")

	notifications, raw, err := fetchNotifications(api, token)

	if err != nil {
		return "fetchNotifications error", err
	}

	log.Println(string(raw))

	mentions := []Mention{}

	for _, notification := range notifications {
		if !notification.IsMention() {
			continue
		}

		mention, err := fetchMention(notification.Subject.LatestCommentURL)

		if err != nil {
			log.Printf("fetchNotifications error: %v+\n", err)
			continue
		}

		mentions = append(mentions, mention)
	}

	err = postMessage(hook, mentions)

	if err != nil {
		return "postMessage error", err
	}

	return "ok", nil
}

func fetch(url, token string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	if len(token) > 0 {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func fetchMention(url string) (Mention, error) {
	var mention Mention

	b, err := fetch(url, "")

	if err != nil {
		return mention, err
	}

	return mention, json.Unmarshal(b, &mention)
}

func fetchNotifications(url, token string) ([]Notification, []byte, error) {
	var notifications []Notification

	since, before := sinceAndBefore(time.Now().UTC())
	queried := fmt.Sprintf("%s&since=%s&before=%s", url, since, before)
	b, err := fetch(queried, token)

	if err != nil {
		return notifications, b, err
	}

	return notifications, b, json.Unmarshal(b, &notifications)
}

func postMessage(url string, mentions []Mention) error {
	attachments := make([]Attachment, len(mentions))

	for i, mention := range mentions {
		attachments[i] = Attachment{
			Color:      "#1abc9c",
			AuthorName: fmt.Sprintf("from @%v", mention.User.Login),
			AuthorLink: mention.User.URL,
			Title:      mention.HTMLURL,
			TitleLink:  mention.HTMLURL,
			Text:       mention.Body,
			ThumbURL:   mention.User.AvatarURL,
		}
	}

	payloadString, err := json.Marshal(
		Payload{
			Attachments: attachments,
		},
	)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadString))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func sinceAndBefore(now time.Time) (string, string) {
	since := now.Add(time.Duration(-1) * time.Hour).Format("2006-01-02T15") + ":00:00Z"
	before := now.Format("2006-01-02T15:04:05Z")

	return since, before
}

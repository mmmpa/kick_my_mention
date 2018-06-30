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
	AuthorAssociation string `json:"author_association"`
	Body              string `json:"body"`
	CreatedAt         string `json:"created_at"`
	HTMLURL           string `json:"html_url"`
	ID                int    `json:"id"`
	IssueURL          string `json:"issue_url"`
	NodeID            string `json:"node_id"`
	UpdatedAt         string `json:"updated_at"`
	URL               string `json:"url"`
	User struct {
		AvatarURL         string `json:"avatar_url"`
		EventsURL         string `json:"events_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		GravatarID        string `json:"gravatar_id"`
		HTMLURL           string `json:"html_url"`
		ID                int    `json:"id"`
		Login             string `json:"login"`
		NodeID            string `json:"node_id"`
		OrganizationsURL  string `json:"organizations_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		ReposURL          string `json:"repos_url"`
		SiteAdmin         bool   `json:"site_admin"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		URL               string `json:"url"`
	} `json:"user"`
	NotificationRaw string `json:"notification_raw"`
	Raw string `json:"raw"`
}

type Notification struct {
	ID         string      `json:"id"`
	LastReadAt interface{} `json:"last_read_at"`
	Reason     string      `json:"reason"`
	Repository struct {
		ArchiveURL       string      `json:"archive_url"`
		AssigneesURL     string      `json:"assignees_url"`
		BlobsURL         string      `json:"blobs_url"`
		BranchesURL      string      `json:"branches_url"`
		CollaboratorsURL string      `json:"collaborators_url"`
		CommentsURL      string      `json:"comments_url"`
		CommitsURL       string      `json:"commits_url"`
		CompareURL       string      `json:"compare_url"`
		ContentsURL      string      `json:"contents_url"`
		ContributorsURL  string      `json:"contributors_url"`
		DeploymentsURL   string      `json:"deployments_url"`
		Description      interface{} `json:"description"`
		DownloadsURL     string      `json:"downloads_url"`
		EventsURL        string      `json:"events_url"`
		Fork             bool        `json:"fork"`
		ForksURL         string      `json:"forks_url"`
		FullName         string      `json:"full_name"`
		GitCommitsURL    string      `json:"git_commits_url"`
		GitRefsURL       string      `json:"git_refs_url"`
		GitTagsURL       string      `json:"git_tags_url"`
		HooksURL         string      `json:"hooks_url"`
		HTMLURL          string      `json:"html_url"`
		ID               int         `json:"id"`
		IssueCommentURL  string      `json:"issue_comment_url"`
		IssueEventsURL   string      `json:"issue_events_url"`
		IssuesURL        string      `json:"issues_url"`
		KeysURL          string      `json:"keys_url"`
		LabelsURL        string      `json:"labels_url"`
		LanguagesURL     string      `json:"languages_url"`
		MergesURL        string      `json:"merges_url"`
		MilestonesURL    string      `json:"milestones_url"`
		Name             string      `json:"name"`
		NodeID           string      `json:"node_id"`
		NotificationsURL string      `json:"notifications_url"`
		Owner struct {
			AvatarURL         string `json:"avatar_url"`
			EventsURL         string `json:"events_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			GravatarID        string `json:"gravatar_id"`
			HTMLURL           string `json:"html_url"`
			ID                int    `json:"id"`
			Login             string `json:"login"`
			NodeID            string `json:"node_id"`
			OrganizationsURL  string `json:"organizations_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			ReposURL          string `json:"repos_url"`
			SiteAdmin         bool   `json:"site_admin"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			Type              string `json:"type"`
			URL               string `json:"url"`
		} `json:"owner"`
		Private         bool   `json:"private"`
		PullsURL        string `json:"pulls_url"`
		ReleasesURL     string `json:"releases_url"`
		StargazersURL   string `json:"stargazers_url"`
		StatusesURL     string `json:"statuses_url"`
		SubscribersURL  string `json:"subscribers_url"`
		SubscriptionURL string `json:"subscription_url"`
		TagsURL         string `json:"tags_url"`
		TeamsURL        string `json:"teams_url"`
		TreesURL        string `json:"trees_url"`
		URL             string `json:"url"`
	} `json:"repository"`
	Subject struct {
		LatestCommentURL string `json:"latest_comment_url"`
		Title            string `json:"title"`
		Type             string `json:"type"`
		URL              string `json:"url"`
	} `json:"subject"`
	SubscriptionURL string `json:"subscription_url"`
	Unread          bool   `json:"unread"`
	UpdatedAt       string `json:"updated_at"`
	URL             string `json:"url"`
}

func (n Notification) IsMention() bool {
	return n.Reason == "mention"
}

func (n Notification) IsAssign() bool {
	return n.Reason == "assign"
}

func (n Notification) IsAuthor() bool {
	return n.Reason == "author"
}

func (n Notification) IsNotificationRequired() bool {
	return n.Unread && (n.IsMention() || n.IsAssign() || n.IsAuthor())
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

	notifications, _, err := fetchNotifications(api, token)

	if err != nil {
		return "fetchNotifications error", err
	}

	mentions := []Mention{}

	for _, notification := range notifications {
		if !notification.IsNotificationRequired() {
			continue
		}

		mention, err := fetchMention(notification.Subject.LatestCommentURL)

		j, err := json.MarshalIndent(notification, "", "  ")

		if err == nil {
			mention.NotificationRaw = string(j)
		}

		if err != nil {
			log.Printf("fetchNotifications error: %v+\n", err)
			continue
		}

		j, err = json.MarshalIndent(mention, "", "  ")

		if err == nil {
			mention.Raw = string(j)
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
		text := mention.Body

		if text == "" {
			text = fmt.Sprintf("```\n%s\n```", mention.NotificationRaw)
		}

		attachments[i] = Attachment{
			Color:      "#1abc9c",
			AuthorName: fmt.Sprintf("from @%v", mention.User.Login),
			AuthorLink: mention.User.URL,
			Title:      mention.HTMLURL,
			TitleLink:  mention.HTMLURL,
			Text:       text,
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

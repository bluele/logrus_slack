/*
Slack Hooks for Logrus

	package main

	import (
		"github.com/Sirupsen/logrus"
		"github.com/bluele/logrus_slack"
	)

	const (
		// slack webhook url
		hookURL = "https://hooks.slack.com/TXXXXX/BXXXXX/XXXXXXXXXX"
	)

	func main() {
		logrus.SetLevel(logrus.DebugLevel)

		logrus.AddHook(&logrus_slack.SlackHook{
			HookURL:        hookURL,
			AcceptedLevels: logrus_slack.LevelThreshold(logrus.WarnLevel),
			Channel:        "#general",
			IconEmoji:      ":ghost:",
			Username:       "logrus_slack",
		})

		logrus.WithFields(logrus.Fields{"foo": "bar", "foo2": "bar2"}).Warn("this is a warn level message")
		logrus.Debug("this is a debug level message")
		logrus.Info("this is an info level message")
		logrus.Error("this is an error level message")
	}

You can specify hook options via `SlackHook` attributes.
*/
package logrus_slack

import (
	"github.com/Sirupsen/logrus"
	"github.com/bluele/slack"
)

// SlackHook is a logrus Hook for dispatching messages to the specified
// channel on Slack.
type SlackHook struct {
	// Messages with a log level not contained in this array
	// will not be dispatched. If nil, all messages will be dispatched.
	AcceptedLevels []logrus.Level
	HookURL        string // Webhook URL

	// slack post parameters
	Username  string // display name
	Channel   string // `#channel-name`
	IconEmoji string // emoji string ex) ":ghost:":
	IconURL   string // icon url

	FieldHeader string // a header above field data
	Async       bool   // if async is true, send a message asynchronously.

	hook *slack.WebHook
}

// Fire -  Sent event to slack
func (sh *SlackHook) Fire(e *logrus.Entry) error {
	if sh.hook == nil {
		sh.hook = slack.NewWebHook(sh.HookURL)
	}

	payload := &slack.WebHookPostPayload{
		Username:  sh.Username,
		Channel:   sh.Channel,
		IconEmoji: sh.IconEmoji,
		IconUrl:   sh.IconURL,
	}
	color, _ := LevelColorMap[e.Level]

	attachment := slack.Attachment{}
	payload.Attachments = []*slack.Attachment{&attachment}

	// If there are fields we need to render them at attachments
	if len(e.Data) > 0 {
		// Add a header above field data
		attachment.Text = sh.FieldHeader

		for k, v := range e.Data {
			field := &slack.AttachmentField{}

			if str, ok := v.(string); ok {
				field.Title = k
				field.Value = str
				// If the field is <= 20 then we'll set it to short
				if len(str) <= 20 {
					field.Short = true
				}
			}
			attachment.Fields = append(attachment.Fields, field)
		}
		attachment.Pretext = e.Message
	} else {
		attachment.Text = e.Message
	}
	attachment.Fallback = e.Message
	attachment.Color = color

	if sh.Async {
		go sh.hook.PostMessage(payload)
		return nil
	}
	return sh.hook.PostMessage(payload)
}

// Levels sets which levels to sent to slack
func (sh *SlackHook) Levels() []logrus.Level {
	if sh.AcceptedLevels == nil {
		return AllLevels
	}
	return sh.AcceptedLevels
}

var LevelColorMap = map[logrus.Level]string{
	logrus.DebugLevel: "#9B30FF",
	logrus.InfoLevel:  "good",
	logrus.WarnLevel:  "warning",
	logrus.ErrorLevel: "danger",
	logrus.FatalLevel: "danger",
	logrus.PanicLevel: "danger",
}

// Supported log levels
var AllLevels = []logrus.Level{
	logrus.DebugLevel,
	logrus.InfoLevel,
	logrus.WarnLevel,
	logrus.ErrorLevel,
	logrus.FatalLevel,
	logrus.PanicLevel,
}

// LevelThreshold - Returns every logging level above and including the given parameter.
func LevelThreshold(l logrus.Level) []logrus.Level {
	for i := range AllLevels {
		if AllLevels[i] == l {
			return AllLevels[i:]
		}
	}
	return []logrus.Level{}
}

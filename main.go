package main

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"net/http"
	"os"
	"strings"
)

var slack_token = os.Getenv("CATS_LIITLE_KEY")
var api = slack.New(slack_token)

func main() {
	port, ok := os.LookupEnv("PORT")
	if ok == false {
		port = "9292"
	}

	http.HandleFunc("/help", CreatePost)
	http.HandleFunc("/update", UpdatePost)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		fmt.Printf("Could not parse action response JSON: %v", err)
	}

	OriginalMessage := payload.OriginalMessage
	attatchment := OriginalMessage.Attachments[0]
	attatchmentText := &attatchment.Text

	if !(strings.Contains(*attatchmentText, payload.User.Name)) {
		if strings.Contains(*attatchmentText, "Thanks in advance") {
			*attatchmentText += fmt.Sprintf(", %s", payload.User.Name)
		} else {
			*attatchmentText += fmt.Sprintf("Thanks in advance to:  %s", payload.User.Name)
		}
	}

	channel := payload.Channel.GroupConversation.Conversation.ID
	timestamp := payload.OriginalMessage.Msg.Timestamp
	api.UpdateMessage(channel, timestamp, slack.MsgOptionAttachments(attatchment))
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	s, err := slack.SlashCommandParse(r)

	if err != nil {
		fmt.Fprintf(w, "SlashCommandParse error", err)
	}

	args := strings.SplitN(s.Text, "-", 4)
	n_people, action, location, time := args[0], args[1], args[2], args[3]

	attachment := slack.Attachment{
		Pretext:    fmt.Sprintf("<!here> Are there %s people available to chip in on:\n* Task: %s \n* Location: %s \n* Time: %s", n_people, action, location, time),
		Text:       "Would you like to help?\n",
		Color:      "#3AA3E3",
		CallbackID: "accept",
		MarkdownIn: []string{"text"},
		Actions: []slack.AttachmentAction{
			{
				Name:  "yes",
				Value: "yes",
				Text:  "Yes",
				Type:  "button",
			},
		},
	}

	channelID, timestamp, err := api.PostMessage(s.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Fprintf(w, "PostMessage error", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

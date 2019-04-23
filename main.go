package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/nlopes/slack"

	"github.com/jokaorgua/trello"

	"github.com/op/go-logging"

	"github.com/joho/godotenv"
)

var (
	listenPort                string
	serverAddr                string
	log                       logging.Logger
	trelloClient              *trello.Client
	trelloUsername            string
	trelloWebhookCallbackURL  string
	trelloSlackLoginRelations []loginRelation
	slackApi                  *slack.Client
)

type loginRelation struct {
	trello string
	slack  string
}

type slackApiInterface interface {
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}

func init() {

	trelloSlackLoginRelations = append(trelloSlackLoginRelations, loginRelation{trello: "@dmytro489", slack: "UFYFUF5DM"})
	log := logging.MustGetLogger("trello_to_slack")
	logFormatter, err := logging.NewStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.6s}%{color:reset} %{message}`)
	logging.SetFormatter(logFormatter)

	err = godotenv.Load()
	if err != nil {
		log.Panic("Can not load .env file")
	}

	listenPort = GetEnvVar("LISTEN_PORT", "80")
	serverAddr = GetEnvVar("LISTEN_IP", "0.0.0.0")
	trelloUsername = GetEnvVar("TRELLO_USERNAME", "anonymous")
	trelloWebhookCallbackURL = GetEnvVar("TRELLO_WEBHOOK_URL", "")
	if trelloWebhookCallbackURL == "" {
		log.Panic("Please set TRELLO_WEBHOOK_URL")
	}

	slackApi = slack.New(GetEnvVar("SLACK_TOKEN", ""))

}
func setupTrelloWebhook() {
	trelloClient = trello.NewClient(GetEnvVar("TRELLO_APIKEY", ""), GetEnvVar("TRELLO_TOKEN", ""))
	log.Info("I'm trello user " + trelloUsername)
	log.Info("Going to setup trello webhooks for URL: " + trelloWebhookCallbackURL)

	member, err := trelloClient.GetMember(trelloUsername, trello.Defaults())

	if err != nil {
		log.Panic("Can not get trello member")
	}

	memberBoards, err := member.GetBoards(trello.Defaults())
	if err != nil {
		log.Panic("Can not get boards from trello")
	}
	token, _ := trelloClient.GetToken(trelloClient.Token, trello.Defaults())
	for _, board := range memberBoards {
		webHooks, err := token.GetWebhooks(trello.Defaults())
		if err != nil {
			log.Error("Can not get webhooks for board " + board.ID + "(" + board.Desc + "). Error: " + err.Error())
		}
		if GetEnvVar("TRELLO_CLEAR_PREVIOUS_WEBHOOKS", "0") != "0" {
			log.Info("Clearing old webhooks for board " + board.ID + "(" + board.Name + ")")
			for _, webhook := range webHooks {
				err = webhook.Delete(trello.Defaults())
				if err != nil {
					log.Error("Can not delete webhook " + webhook.ID)
				}
			}
			webHooks, err = token.GetWebhooks(trello.Defaults())
		}
		webhooksUrls := getWebHooksUrls(webHooks)
		log.Debug("Current webhooks urls for board "+board.ID+"("+board.Name+")", webhooksUrls)

		if !SliceContains(webhooksUrls, trelloWebhookCallbackURL) {
			log.Debug("Creating webhook for board " + board.ID + "(" + board.Name + ")")

			webhook := &trello.Webhook{IDModel: board.ID, Description: "Test webhook", CallbackURL: trelloWebhookCallbackURL}
			err := trelloClient.CreateWebhook(webhook)
			if webhook.Active == false {
				log.Error("Can not create webhooks for board " + board.ID + "(" + board.Name + "). Active: false")
				continue
			}
			if err != nil {
				log.Error("Can not create webhooks for board " + board.ID + "(" + board.Name + "). Error: " + err.Error())
				continue
			}
			log.Info("Created webhook for board " + board.ID + "(" + board.Name + ")")
		}
	}
	log.Info("Setup of trello webhooks was made")
}

func main() {
	log.Info("Will listen " + serverAddr + ":" + listenPort)
	go http.ListenAndServe(serverAddr+":"+listenPort, handlers())
	setupTrelloWebhook()
}

func getWebHooksUrls(webhooks []*trello.Webhook) []string {
	result := []string{}
	for _, webhook := range webhooks {
		result = append(result, webhook.CallbackURL)
	}

	return result
}

func sendToSlack(userId string, text string, slackApi slackApiInterface) error {
	_, _, err := slackApi.PostMessage(userId, slack.MsgOptionText(text, false))
	if err != nil {
		log.Error("Can not send message to slack: " + err.Error())
		return err
	}
	return err
}

func handlers() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Got request")
		w.WriteHeader(200)
		w.Write([]byte("ok"))

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Panic("Cant read request's body")
		}
		log.Debug("Received request: " + string(bodyBytes))

		var requestJson = map[string]interface{}{}

		err = json.Unmarshal(bodyBytes, &requestJson)
		if err != nil {
			log.Error("Cant parse request")
			return
		}
		parsedRequest := trello.CardWebhookRequest{}
		json.Unmarshal(bodyBytes, &parsedRequest)

		//we will react only on comments
		if parsedRequest.Action.Type != "commentCard" {
			return
		}

		for _, relation := range trelloSlackLoginRelations {
			if strings.Contains(parsedRequest.Action.Data.Text, relation.trello) {
				message := ""
				message = "<https://trello.com/c/" + parsedRequest.Action.Data.Card.ShortLink + "|" + parsedRequest.Action.Data.Card.Name + ">\n" +
					"> " + parsedRequest.Action.Data.Text
				err := sendToSlack(relation.slack, message, slackApi)
				if err != nil {
					log.Error("Can not send message to slack to " + relation.trello + "(" + relation.slack + ") Error: " + err.Error())
					continue
				}
				log.Info("Sent message to slack to " + relation.trello + "(" + relation.slack + ")")
			}
		}

	})
	return r
}

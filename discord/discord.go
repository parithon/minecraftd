package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const username = "photon"
const discordWebhookUrl = "https://discordapp.com/api/webhooks/849143840349749309/CC-ofKpgW7XMYTVpAduDz5al4jyd1Jp7YvYz2RsxgsyF-XJjj5iR3hJbsiEJ6zJjoLoG"
const startMessage = "Minecraft Bedrock server version %s started"
const stopMessage = "Minecraft Bedrock server stopped"

func Started(version string) error {
	body, err := json.Marshal(map[string]string{
		"username": username,
		"content":  fmt.Sprintf(startMessage, version),
	})
	if err != nil {
		return err
	}

	if _, err := http.Post(discordWebhookUrl, "application/json", bytes.NewBuffer(body)); err != nil {
		return err
	}

	return nil
}

func Stopped() error {
	body, err := json.Marshal(map[string]string{
		"username": username,
		"content":  stopMessage,
	})
	if err != nil {
		return err
	}

	if _, err := http.Post(discordWebhookUrl, "application/json", bytes.NewBuffer(body)); err != nil {
		return err
	}

	return nil
}

func HealthChecked(isHealth bool) error {
	var content string
	if isHealth {
		content = "Minecraft Bedrock server reported healthy"
	} else {
		content = "Minecraft Bedrock server reported unhealthy"
	}
	body, err := json.Marshal(map[string]string{
		"username": username,
		"content":  content,
	})
	if err != nil {
		return err
	}

	if _, err := http.Post(discordWebhookUrl, "application/json", bytes.NewBuffer(body)); err != nil {
		return err
	}

	return nil
}

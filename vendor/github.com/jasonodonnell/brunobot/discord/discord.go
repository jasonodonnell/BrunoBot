package discord

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type payload struct {
	Content string `json:"content"`
}

func Send(data, webhook string) (err error) {
	blob, err := json.Marshal(payload{data})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(blob))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return
}

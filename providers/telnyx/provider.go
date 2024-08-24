package telnyx

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
  "bytes"

	"src.agwa.name/sms-over-xmpp"
)

type Provider struct {
	service      *smsxmpp.Service

	ApiUrl       string
	ApiKey       string

}

type inboundSMS struct {
  Data struct {
    Payload struct {
      From struct {
        PhoneNumber string `json:"phone_number"`
      }
      To []struct {
        PhoneNumber string `json:"phone_number"`
      }
      Text string
      Media []struct {
        Url string
      }
    }
  }
}

func (provider *Provider) Type() string {
	return "telnyx"
}

func (provider *Provider) Send(message *smsxmpp.Message) error {
	postBody, err := json.Marshal(map[string]string{
		"from":                 message.From,
		"to":                   message.To,
		"text":                 message.Body,
	})
	if err != nil {
		return err
	}

	// create request
	req, _ := http.NewRequest("POST", provider.ApiUrl+"/messages", bytes.NewBuffer(postBody))
	req.Header.Add("Authorization", "Bearer "+provider.ApiKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")

	// make request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// read json body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// unserialize json body
	unmarshaled := map[string]interface{}{}
	err = json.Unmarshal(body, &unmarshaled)
	if err != nil {
		return err
	}

	// if status code 200
	if res.StatusCode == 200 {
		return fmt.Errorf("%s",unmarshaled["data"].(map[string]interface{}))
	} else {
		return nil
  }
}

func (provider *Provider) HTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/message", provider.handleMessage)
  return mux
}

func (provider *Provider) handleMessage(w http.ResponseWriter, req *http.Request) {
	requestBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "400 Bad Request: unable to read request body", 400)
		return
	}

  var msg inboundSMS
  if err := json.Unmarshal(requestBytes, &msg); err != nil {
		http.Error(w, fmt.Sprintf("400 Bad Request: malformed JSON: %#v", err), 400)
		return
	}
  var urls []string
  for _, url := range  msg.Data.Payload.Media {
    urls = append(urls, url.Url)
  }

	message := smsxmpp.Message{
		From: msg.Data.Payload.From.PhoneNumber,
		To: msg.Data.Payload.To[0].PhoneNumber,
		Body: msg.Data.Payload.Text,
    MediaURLs: urls,
	}
	if err := provider.service.Receive(&message); err != nil {
		http.Error(w, fmt.Sprintf("500 Internal Server Error: failed to receive message: %#v", err), 500)
		return
	}
	w.WriteHeader(204)
}

func MakeProvider(service *smsxmpp.Service, config smsxmpp.ProviderConfig) (smsxmpp.Provider, error) {
	return &Provider{
		service: service,
		ApiUrl: config["api_url"],
		ApiKey: config["api_key"],
	}, nil
}

func init() {
	smsxmpp.RegisterProviderType("telnyx", MakeProvider)
}

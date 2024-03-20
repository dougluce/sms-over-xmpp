package telnyx

import (
	"net/http"
	"net/url"
	"strings"
	"fmt"
	"io/ioutil"
	"encoding/json"

	"src.agwa.name/sms-over-xmpp"
)

type Provider struct {
	service      *smsxmpp.Service

	apiKey       string
	apiSecret    string
	httpPassword string
}

func (provider *Provider) Type() string {
	return "telnyx"
}

func (provider *Provider) Send(message *smsxmpp.Message) error {
	request := make(url.Values)
	request.Set("api_key", provider.apiKey)
	request.Set("api_secret", provider.apiSecret)
	request.Set("from", strings.TrimPrefix(message.From, "+"))
	request.Set("to", strings.TrimPrefix(message.To, "+"))
	request.Set("text", message.Body)
	if !isASCII(message.Body) {
		// TODO: test non-ASCII messages
		request.Set("type", "unicode")
	}

	response, err := provider.sendSMS(request)
	if err != nil {
		return err
	}

	for _, message := range response.Messages {
		if message.Status != "0" {
			return fmt.Errorf("Error sending SMS (%s): %s", message.Status, sendSMSStatuses[message.Status])
		}
	}

	return nil
}

func (provider *Provider) HTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/message", provider.handleInboundSMS)
  return mux
}

func (provider *Provider) handleInboundSMS(w http.ResponseWriter, req *http.Request) {
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
		apiKey: config["api_key"],
		apiSecret: config["api_secret"],
		httpPassword: config["http_password"],
	}, nil
}

func init() {
	smsxmpp.RegisterProviderType("telnyx", MakeProvider)
}

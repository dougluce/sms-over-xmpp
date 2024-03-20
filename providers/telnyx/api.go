package telnyx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"strings"
)

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

type sendSMSResponse struct {
	MessageCount int `json:"message-count"`
	Messages     []struct {
		Status string `json:"status"`
	} `json:"messages"`
}

func (provider *Provider) sendSMS(form url.Values) (*sendSMSResponse, error) {
	req, err := http.NewRequest("POST", "https://rest.nexmo.com/sms/json", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(httpResp.Body)
	httpResp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Error reading response from Nexmo: %s", err)
	}

	if !(httpResp.StatusCode >= 200 && httpResp.StatusCode <= 299) {
		return nil, fmt.Errorf("HTTP error from Nexmo: %s", httpResp.Status)
	}

	resp := new(sendSMSResponse)
	if err := json.Unmarshal(respBytes, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

var sendSMSStatuses = map[string]string{
	"0": "Success",
	"1": "Throttled",
	"2": "Missing Parameters",
	"3": "Invalid Parameters",
	"4": "Invalid Credentials",
	"5": "Internal Error",
	"6": "Invalid Message",
	"7": "Number Barred",
	"8": "Partner Account Barred",
	"9": "Partner Quota Violation",
	"10": "Too Many Existing Binds",
	"11": "Account Not Enabled For HTTP",
	"12": "Message Too Long",
	"14": "Invalid Signature",
	"15": "Invalid Sender Address",
	"22": "Invalid Network Code",
	"23": "Invalid Callback URL",
	"29": "Non-Whitelisted Destination",
	"32": "Signature And API Secret Disallowed",
	"33": "Number De-activated",
}

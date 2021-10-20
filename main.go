package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/twilio/twilio-go/client"
)

func tweet(str string) error {
	if str == "" {
		return errors.New("tweets have length")
	}
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	twclient := twitter.NewClient(httpClient)
	_, _, err := twclient.Statuses.Update(str, nil)
	return err
}

func fromForm(vals url.Values, key string) (string, bool) {
	val, ok := vals[key]
	if !ok {
		return "", false
	}
	if len(val) == 0 {
		return "", false
	}
	return val[0], true
}

func emptyResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(w, "<Response></Response>")
}

func toParams(vals url.Values) map[string]string {
	ret := map[string]string{}
	for k, v := range vals {
		ret[k] = v[0]
	}
	return ret
}

func sms(w http.ResponseWriter, req *http.Request) {
	v := client.NewRequestValidator(os.Getenv("TWILIO_AUTH_TOKEN"))
	req.ParseForm()
	url := os.Getenv("URL") + req.URL.Path
	valid := v.Validate(url, toParams(req.Form), req.Header.Get("X-Twilio-Signature"))
	if !valid {
		fmt.Println("not twilio")
		emptyResponse(w)
		return
	}
	if number, ok := fromForm(req.Form, "From"); !ok || number != os.Getenv("NUMBER") {
		fmt.Println("bad number", number)
		emptyResponse(w)
		return
	}
	body, ok := fromForm(req.Form, "Body")
	if ok {
		err := tweet(body)
		if err != nil {
			fmt.Println("error tweeting:", err)
		} else {
			fmt.Println("tweeted", body)
		}
	}
	emptyResponse(w)
}

func main() {
	http.HandleFunc("/sms", sms)
	http.ListenAndServe(":8090", nil)
}

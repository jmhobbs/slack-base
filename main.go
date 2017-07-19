package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type slackResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

func NewSlackResponseToChannel(message string) slackResponse {
	return slackResponse{"in_channel", message}
}

func NewSlackResponseToUser(message string) slackResponse {
	return slackResponse{"ephemeral", message}
}

func main() {
	secret := os.Getenv("SLACK_TOKEN")
	if secret == "" {
		log.Fatal("SLACK_TOKEN is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on :%s\n", port)

	http.Handle("/base", handler(secret))
	http.ListenAndServe(":"+port, nil)
}

func handler(secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			resp := NewSlackResponseToUser("The slash command is misconfigured: expected a POST request.")
			resp.MustWrite(w)
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Println("error:", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if r.Form.Get("token") != secret {
			resp := NewSlackResponseToUser("The slash command is misconfigured; token does not match.")
			resp.MustWrite(w)
			return
		}

		parameters := strings.Split(r.Form.Get("text"), " ")

		if len(parameters) != 2 || parameters[0] == "help" {
			resp := NewSlackResponseToUser(fmt.Sprintf("Usage: %s <base to convert to> <integer>", r.Form.Get("command")))
			resp.MustWrite(w)
			return
		}

		base, err := strconv.Atoi(parameters[0])
		if err != nil {
			resp := NewSlackResponseToUser(fmt.Sprintf("Could not convert \"%s\" to an integer.", parameters[0]))
			resp.MustWrite(w)
			return
		}

		n, err := strconv.Atoi(parameters[1])
		if err != nil {
			resp := NewSlackResponseToUser(fmt.Sprintf("Could not convert \"%s\" to an integer.", parameters[1]))
			resp.MustWrite(w)
			return
		}

		if base < 2 || base > 36 {
			resp := NewSlackResponseToUser("Can only convert to bases between 2 and 36, inclusive. not convert \"%s\" to an integer.")
			resp.MustWrite(w)
			return
		}

		resp := NewSlackResponseToChannel(strconv.FormatInt(int64(n), base))
		resp.MustWrite(w)
	})
}

func (s slackResponse) MustWrite(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	err := enc.Encode(s)

	if err != nil {
		log.Println("error:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

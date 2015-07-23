package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

func init() {
	http.HandleFunc("/", handleIndex)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets/"))))
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/api/", handleAPI)
}

func handleAPI(res http.ResponseWriter, req *http.Request) {
	// switch on the api method
	switch strings.SplitN(req.URL.Path, "/", 3)[2] {
	case "tweets":
		switch req.Method {
		case "POST":
			handleUserTweet(res, req)
		default:
			http.Error(res, "method not allowed", 405)
		}
	default:
		http.NotFound(res, req)
	}
}

func handleIndex(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		handleUserProfile(res, req)
		return
	}
	ctx := appengine.NewContext(req)
	tweets, err := getTweets(ctx)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	type Model struct {
		Tweets []*Tweet
	}
	model := Model{
		Tweets: tweets,
	}
	renderTemplate(res, req, "section", model)
}

func handleLogin(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	profile, err := getProfileByEmail(ctx, u.Email)
	if err == nil && profile.Username != "" {
		http.SetCookie(res, &http.Cookie{Name: "logged_in", Value: "true"})
		http.Redirect(res, req, "/"+profile.Username, 302)
		return
	}

	type Model struct {
		Profile *Profile
		Error   string
	}
	model := Model{
		Profile: &Profile{Email: u.Email},
	}

	// create the profile
	username := req.FormValue("username")
	if username != "" {
		_, err = getProfileByUsername(ctx, username)
		// if the username is already taken
		if err == nil {
			model.Error = "username is not available"
		} else {
			model.Profile.Username = username
			// try to create the profile
			err = createProfile(ctx, model.Profile)
			if err != nil {
				model.Error = err.Error()
			} else {
				// on success redirect to the user's timeline
				waitForProfile(ctx, username)
				http.SetCookie(res, &http.Cookie{Name: "logged_in", Value: "true"})
				http.Redirect(res, req, "/"+username, 302)
				return
			}
		}
	}

	// render the login template
	renderTemplate(res, req, "login", model)
}

func handleLogout(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{Name: "logged_in", Value: ""})
	http.Redirect(res, req, "/", 302)
}

func handleUserProfile(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	username := strings.SplitN(req.URL.Path, "/", 2)[1]
	profile, err := getProfileByUsername(ctx, username)
	if err != nil {
		http.Error(res, err.Error(), 404)
		return
	}

	tweets, err := getUserTweets(ctx, profile.Username)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	type Model struct {
		Profile *Profile
		Tweets  []*Tweet
	}
	model := Model{
		Profile: profile,
		Tweets:  tweets,
	}
	renderTemplate(res, req, "usernamefilter", model)
}

func handleUserTweet(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	profile, err := getProfileByEmail(ctx, u.Email)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	var tweet Tweet
	err = json.NewDecoder(req.Body).Decode(&tweet)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	tweet.Time = time.Now()
	tweet.Username = profile.Username

	// create the tweet
	err = createTweet(ctx, profile.Email, &tweet)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	json.NewEncoder(res).Encode(tweet)
}

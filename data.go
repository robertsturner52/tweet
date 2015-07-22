import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func createProfile() {

}

func getProfileByEmail() {

}

func getProfileByUsername(ctx context.Context, username string) (*Profile, error) {
	q := datastore.NewQuery("Profile").Filter("Username =", username).Limit(1)
	var profiles []Profile
	_, err := q.GetAll(ctx, &profiles)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found")
	}
	return &profiles[0], nil
}

func waitForProfile() {

}

func createTweet() {

}

func getTweets() {

}

func getUserTweets(ctx context.Context, username string) ([]*Tweet, error) {
	q := datastore.NewQuery("Tweet")
	if username != "" {
		q = q.Filter("Username =", username)
	}
	q = q.Order("-Time").Limit(10)
	var tweets []*Tweet
	keys, err := q.GetAll(ctx, &tweets)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(tweets); i++ {
		tweets[i].ID = keys[i].IntID()
	}
	return tweets, nil
}

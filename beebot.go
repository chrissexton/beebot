package beebot

import (
	"fmt"
	"time"

	"code.chrissexton.org/cws/BeeBot/config"
	"github.com/jmoiron/sqlx"
	"github.com/jpillora/backoff"
	"github.com/jzelinskie/geddit"
	"github.com/rs/zerolog/log"
)

// DefaultAddr is the HTTP service address
const DefaultAddr = "127.0.0.1:9595"

// BeeBot represents our bot
type BeeBot struct {
	reddit  string
	logPath string

	nav   map[string]string
	debug bool

	db *sqlx.DB
	c  *config.Config

	*geddit.OAuthSession
}

// New creates a BeeBot instance, its database, and connects to reddit
func New(dbFilePath, logFilePath string, debug bool) (*BeeBot, error) {
	db, err := sqlx.Connect("sqlite", dbFilePath)
	if err != nil {
		return nil, err
	}

	c := config.New(db)

	clientID := c.Get("clientid", "")
	clientSecret := c.Get("clientsecret", "")
	userAgent := c.Get("userAgent", "BeeBot")
	baseAddr := c.Get("baseaddr", DefaultAddr)
	userName := c.Get("username", "")
	password := c.Get("password", "")
	reddit := c.Get("reddit", "")

	o, err := geddit.NewOAuthSession(
		clientID,
		clientSecret,
		userAgent,
		fmt.Sprintf("http://%s/cb", baseAddr),
	)
	if err != nil {
		return nil, err
	}

	if err = o.LoginAuth(userName, password); err != nil {
		return nil, err
	}

	b := &BeeBot{
		reddit:       reddit,
		logPath:      logFilePath,
		nav:          make(map[string]string),
		debug:        debug,
		db:           db,
		c:            c,
		OAuthSession: o,
	}

	b.setupDB()

	return b, nil
}

// Serve starts a polling service with exponential backoff
func (b *BeeBot) Serve(dur time.Duration, done chan (bool)) {
	backoff := &backoff.Backoff{
		Min:    time.Duration(b.c.GetInt("backoff.min", 100)) * time.Millisecond,
		Max:    time.Duration(b.c.GetInt("backoff.max", 10)) * time.Second,
		Factor: b.c.GetFloat64("backoff.factor", 2),
		Jitter: b.c.GetInt("backoff.jitter", 1) == 1,
	}

	for {
		timer := time.NewTimer(backoff.Duration())
		select {
		case <-done:
			timer.Stop()
			return
		case <-timer.C:
			if err := b.Run(); err == nil {
				backoff.Reset()
			}
		}
	}

}

// Run triggers a single query and Filter of the reddit
func (b *BeeBot) Run() error {

	offenders := map[string]map[string]bool{}

	filters, err := b.AllFilters()
	if handleErr(err, "Could not get list of filters") {
		return err
	}

	for _, f := range filters {
		offenders[f.Name] = map[string]bool{}
		tmpOffenders := []string{}
		err := b.db.Select(&tmpOffenders, "select offender from offenders where type=?", f.Name)
		if handleErr(err, "could not get %s offenders", f.Name) {
			return err
		}
		for _, o := range tmpOffenders {
			offenders[f.Name][o] = true
		}
	}

	comments, err := b.SubredditComments(b.reddit)
	if handleErr(err, "could not get subreddit comments for %s", b.reddit) {
		return err
	}
	for _, c := range comments {
		for _, f := range filters {
			if f.regex.MatchString(c.Body) {
				if _, ok := offenders[f.Name][c.Author]; ok {
					log.Debug().Msgf("Skipping offender %s", c.Author)
				} else {
					offenders[f.Name][c.Author] = true
					_, err = b.db.Exec(`insert into offenders (offender, type) values (?, ?)`, c.Author, f.Name)
					if handleErr(err, "could not insert raisin offenders") {
						return err
					}
				}
			}
		}
	}
	log.Debug().Msgf("Processed %d comments", len(comments))

	subOpts := geddit.ListingOptions{
		Limit: 10,
	}
	posts, err := b.SubredditSubmissions(b.reddit, geddit.NewSubmissions, subOpts)
	for _, p := range posts {
		for _, f := range filters {
			if f.regex.MatchString(p.Title) || f.regex.MatchString(p.Selftext) {
				if _, ok := offenders[f.Name][p.Author]; ok {
					log.Debug().Msgf("Skipping offender %s", p.Author)
				} else {
					offenders[f.Name][p.Author] = true
					_, err = b.db.Exec(`insert into offenders (offender, type) values (?, ?)`, p.Author, f.Name)
					if handleErr(err, "could not insert raisin offenders") {
						return err
					}
				}
			}
		}
	}
	log.Debug().Msgf("Processed %d posts", len(posts))

	return nil
}

func in(val string, from []string) bool {
	for _, e := range from {
		if e == val {
			return true
		}
	}
	return false
}

func handleErr(err error, message string, extras ...interface{}) bool {
	if err != nil {
		log.Error().
			Err(err).
			Msgf(message, extras...)
		return true
	}
	return false
}

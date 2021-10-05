// © 2013 the CatBase Authors under the WTFPL. See AUTHORS for the list of authors.

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// Config stores any system-wide startup information that cannot be easily configured via
// the database
type Config struct {
	*sqlx.DB
}

// GetFloat64 returns the config value for a string key
// It will first look in the env vars for the key
// It will check the db for the key if an env DNE
// Finally, it will return a zero value if the key does not exist
// It will attempt to convert the value to a float64 if it exists
func (c *Config) GetFloat64(key string, fallback float64) float64 {
	f, err := strconv.ParseFloat(c.GetString(key, fmt.Sprintf("%f", fallback)), 64)
	if err != nil {
		return 0.0
	}
	return f
}

// GetInt64 returns the config value for a string key
// It will first look in the env vars for the key
// It will check the db for the key if an env DNE
// Finally, it will return a zero value if the key does not exist
// It will attempt to convert the value to an int if it exists
func (c *Config) GetInt64(key string, fallback int64) int64 {
	i, err := strconv.ParseInt(c.GetString(key, strconv.FormatInt(fallback, 10)), 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// GetInt returns the config value for a string key
// It will first look in the env vars for the key
// It will check the db for the key if an env DNE
// Finally, it will return a zero value if the key does not exist
// It will attempt to convert the value to an int if it exists
func (c *Config) GetInt(key string, fallback int) int {
	i, err := strconv.Atoi(c.GetString(key, strconv.Itoa(fallback)))
	if err != nil {
		return 0
	}
	return i
}

// Get is a shortcut for GetString
func (c *Config) Get(key, fallback string) string {
	return c.GetString(key, fallback)
}

func envkey(key string) string {
	key = strings.ToUpper(key)
	key = strings.Replace(key, ".", "", -1)
	return key
}

// GetString returns the config value for a string key
// It will first look in the env vars for the key
// It will check the db for the key if an env DNE
// Finally, it will return a zero value if the key does not exist
// It will convert the value to a string if it exists
func (c *Config) GetString(key, fallback string) string {
	key = strings.ToLower(key)
	if v, found := os.LookupEnv(envkey(key)); found {
		return v
	}
	var configValue string
	q := `select value from config where key=?`
	err := c.DB.Get(&configValue, q, key)
	if err != nil {
		log.Debug().Msgf("WARN: Key %s is empty", key)
		return fallback
	}
	return configValue
}

// Unset removes config values from the database
func (c *Config) Unset(key string) error {
	q := `delete from config where key=?`
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(q, key)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Set changes the value for a configuration in the database
// Note, this is always a string. Use the SetArray for an array helper
func (c *Config) Set(key, value string) error {
	key = strings.ToLower(key)
	value = strings.Trim(value, "`")
	q := `insert into config (key,value) values (?, ?)
			on conflict(key) do update set value=?;`
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(q, key, value, value)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// New loads a configuration from the specified database
func New(db *sqlx.DB) *Config {
	c := Config{}
	c.DB = db

	c.MustExec(`create table if not exists config (
		key string,
		value string,
		primary key (key)
	);`)

	return &c
}

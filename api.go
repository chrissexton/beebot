package beebot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

func (b *BeeBot) apiEndpoints() http.Handler {
	r := chi.NewRouter()
	r.Route("/nav", func(r chi.Router) {
		r.Get("/", b.getNav)
	})
	r.Route("/config", func(r chi.Router) {
		r.Get("/", b.getConfig)
		r.Post("/", b.setConfig)
		r.Delete("/", b.deleteConfig)
	})
	r.Route("/filters", func(r chi.Router) {
		r.Get("/", b.getFilters)
		r.Post("/", b.postFilters)
		r.Put("/{name}", b.putFilters)
		r.Delete("/", b.deleteFilters)
	})
	r.Route("/log", func(r chi.Router) {
		r.Get("/", b.getLog)
	})
	return r
}

type configEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (b *BeeBot) getNav(w http.ResponseWriter, r *http.Request) {
	j, _ := json.Marshal(b.nav)
	w.Write(j)
}

func (b *BeeBot) getConfig(w http.ResponseWriter, r *http.Request) {
	entries := []configEntry{}
	err := b.db.Select(&entries, `select key, value from config`)
	if err != nil {
		log.Error().Err(err).Msg("Could not get configuration entries")
		w.WriteHeader(500)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	j, _ := json.Marshal(entries)
	w.Write(j)
}

func (b *BeeBot) setConfig(w http.ResponseWriter, r *http.Request) {
	config := configEntry{}
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &config)
	if err != nil {
		log.Error().Err(err).Msg("Could not get configuration entries")
		w.WriteHeader(400)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	err = b.c.Set(config.Key, config.Value)
	if err != nil {
		log.Error().Err(err).Msg("Could not set configuration entry")
		w.WriteHeader(400)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
}

func (b *BeeBot) deleteConfig(w http.ResponseWriter, r *http.Request) {
	config := configEntry{}
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &config)
	if err != nil {
		log.Error().Err(err).Msg("Could not get configuration entries")
		w.WriteHeader(400)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	log.Info().Msgf("Deleting config: %s", config.Key)
	err = b.c.Unset(config.Key)
	if err != nil {
		log.Error().Err(err).Msg("Could not unset configuration entry")
		w.WriteHeader(400)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	resp, _ := json.Marshal(struct {
		Status string `json:"status"`
	}{"ok"})
	w.Write(resp)
}

func (b *BeeBot) getFilters(w http.ResponseWriter, r *http.Request) {
	filters, err := b.AllFilters()
	if err != nil {
		log.Error().Err(err).Msg("Could not get filters")
	}
	j, _ := json.Marshal(filters)
	w.Write(j)
}

func (b *BeeBot) postFilters(w http.ResponseWriter, r *http.Request) {
	filter := Filter{}
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &filter)
	filter.populate(b.db)
	if err != nil {
		log.Error().Err(err).Msg("Could not read filter entry")
		w.WriteHeader(400)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	err = filter.Save()
	if err != nil {
		log.Error().Err(err).Msg("Could not save filter")
		w.WriteHeader(400)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	out, err := json.Marshal(filter)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal filter output")
		w.WriteHeader(500)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	w.Write(out)
}

func (b *BeeBot) putFilters(w http.ResponseWriter, r *http.Request) {
}

func (b *BeeBot) deleteFilters(w http.ResponseWriter, r *http.Request) {
}

func (b *BeeBot) getLog(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open(b.logPath)
	logs, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error().Err(err).Msg("Could not open logs")
		w.WriteHeader(500)
		j, _ := json.Marshal(err)
		w.Write(j)
	}
	w.Write(logs)
}

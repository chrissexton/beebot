package beebot

import (
	"regexp"
	"text/template"

	"github.com/jmoiron/sqlx"
)

// Filter represents a comment response for the bot
type Filter struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	RegexStr string `json:"regexStr" db:"regex"`
	Template string `json:"template"`

	regex *regexp.Regexp
	db    *sqlx.DB
}

// NewFilter creates and saves a Filter
func (b *BeeBot) NewFilter(name, regex, tpl string) (*Filter, error) {
	// Verify the template
	_, err := template.New("tmp").Parse(tpl)
	if err != nil {
		return nil, err
	}
	_, err = regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	f := &Filter{
		Name:     name,
		RegexStr: regex,
		Template: tpl,
	}
	f.populate(b.db)
	return f, nil
}

func (f *Filter) populate(db *sqlx.DB) {
	f.regex = regexp.MustCompile(f.RegexStr)
	f.db = db
}

func Remove(db *sqlx.DB, name string) error {
	q := `delete from filters where name=?`
	_, err := db.Exec(q, name)
	return err
}

// Save commits a template to the database
func (f *Filter) Save() error {
	q := `insert into filters (name, regex, template) values (?, ? ,?)
			on conflict(name) do update set regex=?, template=?;`
	res, err := f.db.Exec(q, f.Name, f.RegexStr, f.Template,
		f.RegexStr, f.Template)
	if err != nil {
		return err
	}
	f.ID, err = res.LastInsertId()
	return err
}

// AllFilters returns every filter known
func (b *BeeBot) AllFilters() ([]Filter, error) {
	filters := []Filter{}
	err := b.db.Select(&filters, `select * from Filters`)
	if handleErr(err, "could not get Filter list") {
		return nil, err
	}
	for _, f := range filters {
		f.populate(b.db)
	}
	return filters, nil
}

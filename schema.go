package beebot

func (b *BeeBot) setupDB() error {
	if _, err := b.db.Exec(`create table if not exists filters (
		id integer primary key autoincrement,
		name text unique,
		regex text,
		template text
	)`); err != nil {
		return err
	}

	if _, err := b.db.Exec(`create table if not exists offenders (
		id integer primary key autoincrement,
		offender text,
		filter_name text,

		foreign key(filter_name) references filters(name)
	)`); err != nil {
		return err
	}

	return nil
}

package enum

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func newDB() *sql.DB {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return db
}

var db = newDB()

type Plan struct {
	Tier Tier `json:"tier"`
}

func TestMarshal(t *testing.T) {
	p := Plan{
		Tier: TierStandard,
	}

	s, err := json.Marshal(p)

	if err != nil {
		t.Error(err)
	}

	t.Log(string(s))
}

func TestUnmarshal(t *testing.T) {
	j := []byte(`{"tier":"standard"}`)

	var p Plan

	if err := json.Unmarshal(j, &p); err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", p)
}

func TestValuer(t *testing.T) {
	p := Plan{
		Tier: TierStandard,
	}

	query := `
	INSERT INTO premium.test
	SET tier_to_buy = ?`

	_, err := db.Exec(query, p.Tier)

	if err != nil {
		t.Error(err)
	}
}

func TestScanner(t *testing.T) {
	query := `SELECT tier_to_buy AS tier
	FROM premium.test
	LIMIT 1`

	var p Plan

	err := db.QueryRow(query).Scan(&p.Tier)

	if err != nil {
		t.Error(err)
	}

	t.Log(p)
}

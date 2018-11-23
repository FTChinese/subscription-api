package model

import (
	"encoding/json"

	cache "github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Schedule contains discount plans and duration.
// Start and end are all formatted to ISO8601 in UTC: 2006-01-02T15:04:05Z
type Schedule struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Start     string          `json:"startAt"`
	End       string          `json:"endAt"`
	Plans     map[string]Plan `json:"plans"`
	CreatedAt string          `json:"createdAt"`
	createdBy string
}

// RetrieveSchedule finds a lastest discount schedule whose end time is still after now.
func (env Env) RetrieveSchedule() (Schedule, error) {
	query := `
	SELECT id AS id,
		name AS name, 
		start_utc AS start,
		end_utc AS end,
		plans AS plans,
		created_utc AS createdUtc,
		created_by AS createdBy
	FROM premium.discount_schedule
	WHERE end_utc >= UTC_TIMESTAMP() 
	ORDER BY created_utc DESC
	LIMIT 1`

	var s Schedule
	var plans string
	var start string
	var end string

	err := env.DB.QueryRow(query).Scan(
		&s.ID,
		&s.Name,
		&start,
		&end,
		&plans,
		&s.CreatedAt,
		&s.createdBy,
	)

	if err != nil {
		return s, err
	}

	if err := json.Unmarshal([]byte(plans), &s.Plans); err != nil {
		return s, err
	}

	s.Start = util.ISO8601UTC.FromDatetime(start, nil)
	s.End = util.ISO8601UTC.FromDatetime(end, nil)
	s.CreatedAt = util.ISO8601UTC.FromDatetime(s.CreatedAt, nil)

	// Cache it.
	env.cacheSchedule(s)

	return s, nil
}

func (env Env) cacheSchedule(s Schedule) {
	env.Cache.Set(keySchedule, s, cache.NoExpiration)
}

// ScheduleFromCache gets schedule from cache.
func (env Env) ScheduleFromCache() (Schedule, bool) {
	if x, found := env.Cache.Get(keySchedule); found {
		sch, ok := x.(Schedule)

		if ok {
			logger.WithField("location", "ScheduleFromCache").Infof("Cached schedule found %+v", sch)
			return sch, true
		}

		return Schedule{}, false
	}

	return Schedule{}, false
}

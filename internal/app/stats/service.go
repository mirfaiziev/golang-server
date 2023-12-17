package stats

import (
	"fmt"
	"time"
)

const (
	// more than 100 years, should be enough
	MaxWeeks = 6000
	// 60 km/h, which is 16.67 m/s, source https://en.wikipedia.org/wiki/Footspeed, maximal recorded ~45 km/h, so 60 more than enough
	MaxRunnerSpeed = 17
)

type InputItem struct {
	Distance  int       `json:"distance"` // meters
	Time      int       `json:"time"`     // seconds
	Timestamp time.Time `json:"timestamp"`
}

type Output struct {
	MediumDistance       int `json:"medium_distance"`
	MediumTime           int `json:"medium_time"`
	MaxDistance          int `json:"max_distance"`
	MaxTime              int `json:"max_time"`
	MediumWeeklyDistance int `json:"medium_weekly_distance"`
	MediumWeeklyTime     int `json:"medium_weekly_time"`
	MaxWeeklyDistance    int `json:"max_weekly_distance"`
	MaxWeeklyTime        int `json:"max_weekly_time"`
}

type WeeklyStats struct {
	totalDistance int
	totalTime     int
}

// key - start of the week in unix timestamp
type WeeklyStatsBuckets map[int64]WeeklyStats

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Analyze(nweeks int, now time.Time, ii []InputItem) (*Output, error) {
	// no stats - show 0 everywhere
	if len(ii) == 0 {
		return &Output{}, nil
	}

	buckets := make(WeeklyStatsBuckets, nweeks)
	earliest := s.beginOfWeek(now.AddDate(0, 0, -7*(nweeks-1)))

	// calculate common stats and fill weekly buckets, which will be used to calculate weekly stats
	var (
		maxDistance    int
		maxTime        int
		totalDistance  int
		totalTime      int
		mediumDistance int
		mediumTime     int
		workoutsNumber int
	)

	for _, item := range ii {
		// doesn't include into analyze event happened before earliest possible time (Monday of the earliest week)
		if item.Timestamp.Before(earliest) {
			continue
		}

		// validate input item, distance and time must be both positive or both zero
		if (item.Distance > 0 && item.Time == 0) || (item.Distance == 0 && item.Time > 0) {
			return nil,
				fmt.Errorf(
					"data for timestame %s is incorrect, distance and time must be both positive or both zero",
					item.Timestamp,
				)
		}
		if item.Distance > MaxRunnerSpeed*item.Time {
			return nil,
				fmt.Errorf(
					"data for timestame %s is incorrect, speed is too high: %d m/s",
					item.Timestamp,
					int(item.Distance/item.Time),
				)
		}

		workoutsNumber += 1

		totalDistance += item.Distance
		totalTime += item.Time

		if item.Time > maxTime {
			maxTime = item.Time
		}
		if item.Distance > maxDistance {
			maxDistance = item.Distance
		}

		key := s.beginOfWeek(item.Timestamp).Unix()
		weeklyStats, ok := buckets[key]
		if !ok {
			weeklyStats = WeeklyStats{}
		}

		weeklyStats.totalDistance += item.Distance
		weeklyStats.totalTime += item.Time

		buckets[key] = weeklyStats
	}

	if workoutsNumber > 0 {
		mediumDistance = int(totalDistance / workoutsNumber)
		mediumTime = int(totalTime / workoutsNumber)
	}

	// aggregate weekly stats
	var (
		maxWeeklyDistance int
		maxWeeklyTime     int
	)

	for _, bucket := range buckets {
		b := bucket

		if b.totalDistance > maxWeeklyDistance {
			maxWeeklyDistance = b.totalDistance
		}
		if b.totalTime > maxWeeklyTime {
			maxWeeklyTime = b.totalTime
		}
	}

	return &Output{
		MaxDistance:          maxDistance,
		MaxTime:              maxTime,
		MediumDistance:       mediumDistance,
		MediumTime:           mediumTime,
		MaxWeeklyDistance:    maxWeeklyDistance,
		MaxWeeklyTime:        maxWeeklyTime,
		MediumWeeklyDistance: int(totalDistance / nweeks),
		MediumWeeklyTime:     int(totalTime / nweeks),
	}, nil
}

func (s *Service) beginOfWeek(t time.Time) time.Time {
	// because week starts from Monday, we need to do this modification
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekday -= 1

	return t.AddDate(0, 0, -weekday).
		Add(-time.Duration(t.Hour()) * time.Hour).
		Add(-time.Duration(t.Minute()) * time.Minute).
		Add(-time.Duration(t.Second()) * time.Second).
		Add(-time.Duration(t.Nanosecond()) * time.Nanosecond)
}

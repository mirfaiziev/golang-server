package stats

import (
	"fmt"
	"time"
)

const (
	MaxWeeks = 6000 // more than 100 years, should be enough
)

type InputItem struct {
	Distance  int       `json:"distance"`
	Time      int       `json:"time"`
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
	totalDistance  int
	totalTime      int
	maxDistance    int
	maxTime        int
	workoutsNumber int
	mediumTime     int
	mediumDistance int
}

// key - start of the week in unix timestamp
type WeeklyStatsBuckets map[int64]WeeklyStats

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Analyze(nweeks int, now time.Time, ii ...InputItem) (Output, error) {
	// no stats - show 0 everywhere
	if len(ii) == 0 {
		return Output{}, nil
	}

	buckets := make(WeeklyStatsBuckets, nweeks)

	var earliest, weekStart time.Time

	for i := 0; i < nweeks; i++ {
		weekStart = s.beginOfWeek(now.AddDate(0, 0, -7*i))
		earliest = weekStart
		buckets[weekStart.Unix()] = WeeklyStats{}
	}

	var (
		maxDistance    int
		maxTime        int
		totalDistance  int
		totalTime      int
		workoutsNumber int
		mediumDistance int
		mediumTime     int
	)

	for _, item := range ii {
		// doesn't include into analyze event happened before earliest possible time (monday of the earliest week)
		if item.Timestamp.Before(earliest) {
			continue
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
		weeklyStats := buckets[key]
		weeklyStats.totalDistance += item.Distance
		weeklyStats.totalTime += item.Time

		if item.Distance > weeklyStats.maxDistance {
			weeklyStats.maxDistance = item.Distance
		}
		if item.Time > weeklyStats.maxTime {
			weeklyStats.maxTime = item.Time
		}
		weeklyStats.workoutsNumber += 1

		buckets[key] = weeklyStats
	}

	if workoutsNumber > 0 {
		mediumDistance = int(totalDistance / workoutsNumber)
		mediumTime = int(totalTime / workoutsNumber)
	}

	for key, bucket := range buckets {
		b := bucket
		if b.workoutsNumber == 0 {
			b.mediumTime = 0
			b.mediumDistance = 0
		} else {
			b.mediumTime = int(bucket.totalTime / bucket.workoutsNumber)
			b.mediumDistance = int(bucket.totalDistance / bucket.workoutsNumber)
		}

		buckets[key] = b
	}

	return Output{
		MaxDistance:    maxDistance,
		MaxTime:        maxTime,
		MediumDistance: mediumDistance,
		MediumTime:     mediumTime,
	}, nil
}

func (s *Service) beginOfWeek(t time.Time) time.Time {
	// because week starts from Monday, we need to do this modification
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekday -= 1

	fmt.Println(weekday)

	return t.AddDate(0, 0, -weekday).
		Add(-time.Duration(t.Hour()) * time.Hour).
		Add(-time.Duration(t.Minute()) * time.Minute).
		Add(-time.Duration(t.Second()) * time.Second).
		Add(-time.Duration(t.Nanosecond()) * time.Nanosecond)
}

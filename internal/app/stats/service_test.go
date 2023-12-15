package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestService(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

type ServiceSuite struct {
	suite.Suite
	svc *Service
}

func (s *ServiceSuite) SetupTest() {
	s.svc = NewService()
}

func (s *ServiceSuite) TestAnalyze_empty_input() {
	out, err := s.svc.Analyze(1, time.Now(), []InputItem{})

	s.NoError(err)
	s.Equal(&Output{}, out)
}

func (s *ServiceSuite) TestAnalyze_records_are_too_early() {
	now, _ := time.Parse(time.DateOnly, "2023-12-16")
	passedSunday, _ := time.Parse(time.DateOnly, "2023-12-10")
	passedSaturnday, _ := time.Parse(time.DateOnly, "2023-12-09")

	out, err := s.svc.Analyze(1, now, []InputItem{
		{Time: 100, Distance: 1500, Timestamp: passedSunday},
		{Time: 200, Distance: 3000, Timestamp: passedSaturnday},
	})

	s.NoError(err)
	s.Equal(&Output{}, out)
}

func (s *ServiceSuite) TestAnalyze_incorrect_record_empty_distance() {
	now, _ := time.Parse(time.DateOnly, "2023-12-16")

	out, err := s.svc.Analyze(1, now, []InputItem{
		{Time: 100, Distance: 0, Timestamp: now.AddDate(0, 0, -1)},
		{Time: 200, Distance: 3000, Timestamp: now.AddDate(0, 0, -2)},
	})

	s.Nil(out)
	s.ErrorContains(err, "distance and time must be both positive or both zero")
}

func (s *ServiceSuite) TestAnalyze_incorrect_record_too_fast() {
	now, _ := time.Parse(time.DateOnly, "2023-12-16")

	out, err := s.svc.Analyze(1, now, []InputItem{
		{Time: 100, Distance: 10000, Timestamp: now.AddDate(0, 0, -1)},
		{Time: 200, Distance: 3000, Timestamp: now.AddDate(0, 0, -2)},
	})

	s.Nil(out)
	s.ErrorContains(err, " speed is too high: 100 m/s")
}

func (s *ServiceSuite) TestAnalyze_three_weeks_analyze() {
	now, _ := time.Parse(time.DateOnly, "2023-12-16")

	nweeks := 3

	out, err := s.svc.Analyze(nweeks, now, []InputItem{
		// current week, total week distance 1000+1500=2500, total week time 1800+2700=4500
		{Distance: 1000, Time: 1800, Timestamp: now.AddDate(0, 0, -4)}, // 12.12.2023
		{Distance: 1500, Time: 2700, Timestamp: now.AddDate(0, 0, -5)}, // 11.12.2023
		// week before - no records but including into medium weekly stats
		// 1 week earlier - total week distance 900+800+700+600+500=3500, total week time 1200+1200+1000+650+600=4650
		{Distance: 900, Time: 1200, Timestamp: now.AddDate(0, 0, -13)}, // 3.12.2023
		{Distance: 800, Time: 1200, Timestamp: now.AddDate(0, 0, -14)}, // 2.12.2023
		{Distance: 700, Time: 1000, Timestamp: now.AddDate(0, 0, -16)}, // 30.11.2023
		{Distance: 600, Time: 650, Timestamp: now.AddDate(0, 0, -17)},  // 29.11.2023
		{Distance: 500, Time: 600, Timestamp: now.AddDate(0, 0, -19)},  // 27.11.2023
		// 2 week earlier - exclude from analyze
		{Distance: 2500, Time: 3600, Timestamp: now.AddDate(0, 0, -20)}, // 28.11.2023
	})

	s.Nil(err)
	s.Equal(&Output{
		MaxDistance:          1500,
		MaxTime:              2700,
		MediumDistance:       int((2500 + 3500) / 7), // total distance / number of workouts
		MediumTime:           int((4500 + 4650) / 7), // total time / number of workouts
		MaxWeeklyDistance:    3500,
		MaxWeeklyTime:        4650,
		MediumWeeklyDistance: int((2500 + 3500) / nweeks), // total distance / number of weeks
		MediumWeeklyTime:     int((4500 + 4650) / nweeks), // total time / number of weeks
	}, out)
}

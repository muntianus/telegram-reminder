// scheduler_test.go проверяет планировщик ежедневных сообщений.
package main

import (
	"testing"
	"time"

	"github.com/go-co-op/gocron"
)

type fakeTime struct {
	onNow func(*time.Location) time.Time
}

// Now возвращает фиксированное время.
func (f fakeTime) Now(loc *time.Location) time.Time { return f.onNow(loc) }

// Unix создаёт время из секунд и наносекунд.
func (f fakeTime) Unix(sec int64, nsec int64) time.Time { return time.Unix(sec, nsec) }

// Sleep ожидает указанную длительность.
func (f fakeTime) Sleep(d time.Duration) { time.Sleep(d) }

// TestScheduleDailyMessagesTimes проверяет расписание по умолчанию.
func TestScheduleDailyMessagesTimes(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)
	s.CustomTime(fakeTime{onNow: func(l *time.Location) time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, l)
	}})

	scheduleDailyMessages(s, nil, nil, 0)

	s.StartAsync()
	s.Stop()

	times := []string{}
	for _, job := range s.Jobs() {
		times = append(times, job.NextRun().In(loc).Format("15:04"))
	}
	want := map[string]bool{"13:00": false, "20:00": false}
	for _, tm := range times {
		if _, ok := want[tm]; ok {
			want[tm] = true
		}
	}
	for tstr, ok := range want {
		if !ok {
			t.Fatalf("time %s not scheduled", tstr)
		}
	}
}

// TestScheduleDailyMessagesCustomTimes проверяет пользовательские времена.
func TestScheduleDailyMessagesCustomTimes(t *testing.T) {
	t.Setenv("LUNCH_TIME", "10:15")
	t.Setenv("BRIEF_TIME", "21:30")

	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)
	s.CustomTime(fakeTime{onNow: func(l *time.Location) time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, l)
	}})

	scheduleDailyMessages(s, nil, nil, 0)

	s.StartAsync()
	s.Stop()

	times := []string{}
	for _, job := range s.Jobs() {
		times = append(times, job.NextRun().In(loc).Format("15:04"))
	}
	want := map[string]bool{"10:15": false, "21:30": false}
	for _, tm := range times {
		if _, ok := want[tm]; ok {
			want[tm] = true
		}
	}
	for tstr, ok := range want {
		if !ok {
			t.Fatalf("time %s not scheduled", tstr)
		}
	}
}

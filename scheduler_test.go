package main

import (
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	"telegram-reminder/internal/bot"
)

type fakeTime struct {
	onNow func(*time.Location) time.Time
}

func (f fakeTime) Now(loc *time.Location) time.Time     { return f.onNow(loc) }
func (f fakeTime) Unix(sec int64, nsec int64) time.Time { return time.Unix(sec, nsec) }
func (f fakeTime) Sleep(d time.Duration)                { time.Sleep(d) }

func TestScheduleDailyMessagesTimes(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)
	s.CustomTime(fakeTime{onNow: func(l *time.Location) time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, l)
	}})

	bot.ScheduleDailyMessages(s, nil, nil, 0)

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

func TestScheduleDailyMessagesCustomTimes(t *testing.T) {
	t.Setenv("LUNCH_TIME", "10:15")
	t.Setenv("BRIEF_TIME", "21:30")

	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)
	s.CustomTime(fakeTime{onNow: func(l *time.Location) time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, l)
	}})

	bot.ScheduleDailyMessages(s, nil, nil, 0)

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

func TestScheduleDailyMessagesFromTasksJSON(t *testing.T) {
	tasks := `[{"name":"a","prompt":"p1","time":"08:00"},{"name":"b","prompt":"p2","time":"22:45"}]`
	t.Setenv("TASKS_JSON", tasks)

	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)
	s.CustomTime(fakeTime{onNow: func(l *time.Location) time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, l)
	}})

	bot.ScheduleDailyMessages(s, nil, nil, 0)

	s.StartAsync()
	s.Stop()

	times := []string{}
	for _, job := range s.Jobs() {
		times = append(times, job.NextRun().In(loc).Format("15:04"))
	}
	want := map[string]bool{"08:00": false, "22:45": false}
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

package main

import (
	"context"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
)

type fakeTime struct {
	onNow func(*time.Location) time.Time
}

func (f fakeTime) Now(loc *time.Location) time.Time     { return f.onNow(loc) }
func (f fakeTime) Unix(sec int64, nsec int64) time.Time { return time.Unix(sec, nsec) }
func (f fakeTime) Sleep(d time.Duration)                { time.Sleep(d) }

type noopAI struct{}

func (noopAI) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{}, nil
}

func TestSetupSchedulerTimes(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)
	s.CustomTime(fakeTime{onNow: func(l *time.Location) time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, l)
	}})

	setupScheduler(s, noopAI{}, nil, 0)

	s.StartAsync()
	time.Sleep(10 * time.Millisecond)
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

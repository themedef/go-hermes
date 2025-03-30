package hermes

import (
	"context"
	"testing"
)

func helperCreateAPI() (*CommandAPI, context.Context) {
	db := NewStore(Config{})
	ctx := context.Background()
	return NewCommandAPI(db), ctx
}

func TestCommandAPISet(t *testing.T) {
	api, ctx := helperCreateAPI()

	parts := []string{"SET", "myKey", "myValue"}
	got, err := api.Execute(ctx, parts)
	if err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if got != "OK" {
		t.Errorf("Got=%q, want=%q", got, "OK")
	}
}

func TestCommandAPIGet(t *testing.T) {
	api, ctx := helperCreateAPI()

	_, err := api.Execute(ctx, []string{"SET", "key1", "val1"})
	if err != nil {
		t.Fatalf("SET error: %v", err)
	}

	got, err := api.Execute(ctx, []string{"GET", "key1"})
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	if got != "\"val1\"" {
		t.Errorf("Got=%q, want=%q", got, "\"val1\"")
	}

	got2, err2 := api.Execute(ctx, []string{"GET", "noKey"})
	if err2 != nil {
		t.Fatalf("GET error: %v", err2)
	}
	if got2 != "(nil)" {
		t.Errorf("Got=%q, want=%q", got2, "(nil)")
	}
}

func TestCommandAPIDel(t *testing.T) {
	api, ctx := helperCreateAPI()

	_, err := api.Execute(ctx, []string{"SET", "willDelete", "something"})
	if err != nil {
		t.Fatalf("SET error: %v", err)
	}

	got, err := api.Execute(ctx, []string{"DEL", "willDelete"})
	if err != nil {
		t.Fatalf("DEL error: %v", err)
	}
	if got != "true" {
		t.Errorf("Got=%q, want=%q", got, "true")
	}

	got2, err2 := api.Execute(ctx, []string{"DEL", "willDelete"})
	if err2 != nil {
		t.Fatalf("DEL error: %v", err2)
	}
	if got2 != "false" {
		t.Errorf("Got=%q, want=%q", got2, "false")
	}
}

func TestCommandAPIIncrDecr(t *testing.T) {
	api, ctx := helperCreateAPI()

	got, err := api.Execute(ctx, []string{"INCR", "numKey"})
	if err != nil {
		t.Fatalf("INCR error: %v", err)
	}
	if got != "1" {
		t.Errorf("Got=%q, want=%q", got, "1")
	}

	got2, err2 := api.Execute(ctx, []string{"INCR", "numKey"})
	if err2 != nil {
		t.Fatalf("INCR error: %v", err2)
	}
	if got2 != "2" {
		t.Errorf("Got=%q, want=%q", got2, "2")
	}

	got3, err3 := api.Execute(ctx, []string{"DECR", "numKey"})
	if err3 != nil {
		t.Fatalf("DECR error: %v", err3)
	}
	if got3 != "1" {
		t.Errorf("Got=%q, want=%q", got3, "1")
	}
}

func TestCommandAPIListOps(t *testing.T) {
	api, ctx := helperCreateAPI()

	got, err := api.Execute(ctx, []string{"LPUSH", "listKey", "val1"})
	if err != nil || got != "OK" {
		t.Fatalf("LPUSH got=%q err=%v, want=OK", got, err)
	}

	_, err2 := api.Execute(ctx, []string{"LPUSH", "listKey", "val2"})
	if err2 != nil {
		t.Fatalf("LPUSH2 error: %v", err2)
	}

	got3, err3 := api.Execute(ctx, []string{"RPOP", "listKey"})
	if err3 != nil {
		t.Fatalf("RPOP error: %v", err3)
	}
	if got3 != "val1" {
		t.Errorf("Got=%q, want=%q", got3, "val1")
	}

	got4, err4 := api.Execute(ctx, []string{"LPOP", "listKey"})
	if err4 != nil {
		t.Fatalf("LPOP error: %v", err4)
	}
	if got4 != "val2" {
		t.Errorf("Got=%q, want=%q", got4, "val2")
	}

	got5, err5 := api.Execute(ctx, []string{"RPOP", "listKey"})
	if err5 != nil {
		t.Fatalf("RPOP error: %v", err5)
	}
	if got5 != "(nil)" {
		t.Errorf("Got=%q, want=%q", got5, "(nil)")
	}
}

func TestCommandAPIExpireFind(t *testing.T) {
	api, ctx := helperCreateAPI()

	_, err := api.Execute(ctx, []string{"EXPIRE", "someKey"})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	gotF, errF := api.Execute(ctx, []string{"FIND", "aaa"})
	if errF != nil {
		t.Fatalf("FIND error: %v", errF)
	}
	if gotF != "Keys: []" {
		t.Errorf("Got=%q, want=%q", gotF, "Keys: []")
	}

	_, _ = api.Execute(ctx, []string{"SET", "k1", "aaa"})
	gotF2, errF2 := api.Execute(ctx, []string{"FIND", "aaa"})
	if errF2 != nil {
		t.Fatalf("FIND2 error: %v", errF2)
	}
	if gotF2 != "Keys: [k1]" {
		t.Errorf("Got=%q, want=%q", gotF2, "Keys: [k1]")
	}

	gotE, errE := api.Execute(ctx, []string{"EXPIRE", "k1", "30"})
	if errE != nil {
		t.Fatalf("EXPIRE error: %v", errE)
	}
	if gotE != "OK" {
		t.Errorf("Got=%q, want=%q", gotE, "OK")
	}
}

func TestCommandAPIUnknownQuit(t *testing.T) {
	api, ctx := helperCreateAPI()

	_, err := api.Execute(ctx, []string{"FOO"})
	if err == nil {
		t.Error("Expected error for unknown command, got nil")
	}

	got, err2 := api.Execute(ctx, []string{"QUIT"})
	if err2 != nil {
		t.Fatalf("QUIT error: %v", err2)
	}
	if got != "Bye!" {
		t.Errorf("Got=%q, want=%q", got, "Bye!")
	}
}

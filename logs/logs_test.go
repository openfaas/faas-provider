package logs

import (
	"testing"

	"time"
)

func Test_Message_String_WithNS(t *testing.T) {

	ts := time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC)

	m := Message{
		Name:      "figlet",
		Namespace: "openfaas-fn",
		Instance:  "figlet-pod1",
		Text:      "Watchdog started",
		Timestamp: ts,
	}

	got := m.String()
	want := "2019-11-10 23:00:00 +0000 UTC figlet (openfaas-fn figlet-pod1) Watchdog started"
	if want != got {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func Test_Message_String_WithoutNS(t *testing.T) {

	ts := time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC)

	m := Message{
		Name:      "figlet",
		Namespace: "",
		Instance:  "figlet-pod1",
		Text:      "Watchdog started",
		Timestamp: ts,
	}

	got := m.String()
	want := "2019-11-10 23:00:00 +0000 UTC figlet (figlet-pod1) Watchdog started"
	if want != got {
		t.Fatalf("want %q, got %q", want, got)
	}
}

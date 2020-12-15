package types

import (
	"encoding/json"
	"testing"
)

func TestSerializeMinimumValues(t *testing.T) {
	f := FunctionDeployment{
		Image:   "alexellis2/figlet",
		Service: "figlet",
	}

	res, _ := json.Marshal(f)
	got := string(res)

	want := `{"service":"figlet","image":"alexellis2/figlet"}`
	if string(got) != want {
		t.Fatalf("got: %q\nwant: %q", got, want)
	}
}

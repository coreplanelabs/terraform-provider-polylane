package provider

import (
	"reflect"
	"testing"
)

func TestParseCompositeImportID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		value         string
		expectedParts int
		want          []string
		wantError     bool
	}{
		{name: "slash separated", value: "ws_one/team_two/usr_three", expectedParts: 3, want: []string{"ws_one", "team_two", "usr_three"}},
		{name: "comma separated", value: "ws_one, team_two", expectedParts: 2, want: []string{"ws_one", "team_two"}},
		{name: "wrong part count", value: "ws_one/team_two", expectedParts: 3, wantError: true},
		{name: "empty part", value: "ws_one//usr_three", expectedParts: 3, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseCompositeImportID(test.value, test.expectedParts)
			if test.wantError {
				if err == nil {
					t.Fatalf("expected an error, got parts %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCompositeImportID returned an error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("unexpected parts: got %#v, want %#v", got, test.want)
			}
		})
	}
}

package domain

import (
	"encoding/json"
	"testing"
)

func TestUpdateProductImgNullStates(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"absent", `{}`, 0},
		{"explicit null", `{"img": null}`, -1},
		{"value", `{"img": "https://x/y.png"}`, 1},
	}

	for _, tc := range cases {
		var u UpdateProduct
		if err := json.Unmarshal([]byte(tc.in), &u); err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		switch {
		case !u.Img.Set:
			if tc.want != 0 {
				t.Errorf("%s: got absent, want state %d", tc.name, tc.want)
			}
		case u.Img.Value == nil:
			if tc.want != -1 {
				t.Errorf("%s: got null, want state %d", tc.name, tc.want)
			}
		default:
			if tc.want != 1 {
				t.Errorf("%s: got value, want state %d", tc.name, tc.want)
			}
		}
	}
}

func TestUpdateUserPhotoNullStates(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"absent", `{}`, 0},
		{"explicit null", `{"photo": null}`, -1},
		{"value", `{"photo": "https://x/y.png"}`, 1},
	}

	for _, tc := range cases {
		var u UpdateUser
		if err := json.Unmarshal([]byte(tc.in), &u); err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		switch {
		case !u.Photo.Set:
			if tc.want != 0 {
				t.Errorf("%s: got absent, want state %d", tc.name, tc.want)
			}
		case u.Photo.Value == nil:
			if tc.want != -1 {
				t.Errorf("%s: got null, want state %d", tc.name, tc.want)
			}
		default:
			if tc.want != 1 {
				t.Errorf("%s: got value, want state %d", tc.name, tc.want)
			}
		}
	}
}

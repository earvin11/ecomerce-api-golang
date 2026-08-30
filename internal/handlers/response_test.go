package handlers

import "testing"

func TestBuildMeta(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		page     int
		pageSize int
		want     *Meta
	}{
		{
			name:     "empty catalog",
			total:    0,
			page:     1,
			pageSize: 10,
			want:     &Meta{Total: 0, Page: 1, PerPage: 10, LastPage: 1, Remaining: 0},
		},
		{
			name:     "first page complete",
			total:    10,
			page:     1,
			pageSize: 10,
			want:     &Meta{Total: 10, Page: 1, PerPage: 10, LastPage: 1, Remaining: 0},
		},
		{
			name:     "middle page",
			total:    25,
			page:     2,
			pageSize: 10,
			want:     &Meta{Total: 25, Page: 2, PerPage: 10, LastPage: 3, Remaining: 5},
		},
		{
			name:     "last complete page",
			total:    20,
			page:     2,
			pageSize: 10,
			want:     &Meta{Total: 20, Page: 2, PerPage: 10, LastPage: 2, Remaining: 0},
		},
		{
			name:     "page beyond last",
			total:    25,
			page:     5,
			pageSize: 10,
			want:     &Meta{Total: 25, Page: 5, PerPage: 10, LastPage: 3, Remaining: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMeta(tt.total, tt.page, tt.pageSize)
			if *got != *tt.want {
				t.Errorf("buildMeta(%d, %d, %d) = %+v, want %+v", tt.total, tt.page, tt.pageSize, got, tt.want)
			}
		})
	}
}

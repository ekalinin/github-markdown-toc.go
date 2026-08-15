package insertmd

import (
	"errors"
	"testing"
)

func TestReplaceBetweenMarkers(t *testing.T) {
	tests := []struct {
		name    string
		content string
		block   string
		want    string
		wantErr error
	}{
		{
			name:    "replaces an existing block",
			content: "# Title\n\n<!--ts-->\nstale\nlines\n<!--te-->\n\n## Section\n",
			block:   "* [Title](#title)",
			want:    "# Title\n\n<!--ts-->\n* [Title](#title)\n<!--te-->\n\n## Section\n",
		},
		{
			name:    "fills an empty block",
			content: "<!--ts-->\n<!--te-->\n",
			block:   "* [A](#a)\n* [B](#b)",
			want:    "<!--ts-->\n* [A](#a)\n* [B](#b)\n<!--te-->\n",
		},
		{
			name:    "keeps indented markers and their indentation",
			content: "  <!--ts-->\nstale\n  <!--te-->\n",
			block:   "* [A](#a)",
			want:    "  <!--ts-->\n* [A](#a)\n  <!--te-->\n",
		},
		{
			name:    "tolerates CRLF line endings",
			content: "# Title\r\n<!--ts-->\r\nstale\r\n<!--te-->\r\n",
			block:   "* [Title](#title)",
			want:    "# Title\r\n<!--ts-->\r\n* [Title](#title)\n<!--te-->\r\n",
		},
		{
			name:    "no markers",
			content: "# Title\n",
			block:   "* [Title](#title)",
			wantErr: ErrMarkersNotFound,
		},
		{
			name:    "only the start marker",
			content: "<!--ts-->\n# Title\n",
			block:   "* [Title](#title)",
			wantErr: ErrMarkersNotFound,
		},
		{
			name:    "two pairs",
			content: "<!--ts-->\n<!--te-->\n<!--ts-->\n<!--te-->\n",
			block:   "* [A](#a)",
			wantErr: ErrMultipleMarkerPairs,
		},
		{
			name:    "reversed markers",
			content: "<!--te-->\nbody\n<!--ts-->\n",
			block:   "* [A](#a)",
			wantErr: ErrMarkersOutOfOrder,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := replaceBetweenMarkers([]byte(tt.content), []byte(tt.block))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

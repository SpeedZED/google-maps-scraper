package gmaps_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gosom/google-maps-scraper/gmaps"
)

func TestExtractPhotoMetadataFromPayloadFindsNearbyDate(t *testing.T) {
	t.Parallel()

	images := []gmaps.Image{{
		Title: "All",
		Image: "https://lh3.googleusercontent.com/p/abc123=w224-h398-k-no",
	}}
	payload := []any{
		"place",
		[]any{
			"photo-block",
			"https://lh3.googleusercontent.com/p/abc123=w408-h725-k-no",
			[]any{2024.0, 3.0, 13.0},
			"Posted by Local Guide",
			"Jane Example",
		},
	}

	got := gmaps.ExtractPhotoMetadataFromPayload(payload, images)

	require.Len(t, got, 1)
	require.Equal(t, "All", got[0].Title)
	require.Equal(t, "https://lh3.googleusercontent.com/p/abc123=w224-h398-k-no", got[0].Image)
	require.Equal(t, "2024-03-13", got[0].PostedDate)
	require.Equal(t, gmaps.PhotoMetadataStatusFound, got[0].Status)
	require.Equal(t, gmaps.PhotoMetadataSourcePayload, got[0].Source)
}

func TestExtractPhotoMetadataFromPayloadReturnsUnavailableWhenDateMissing(t *testing.T) {
	t.Parallel()

	images := []gmaps.Image{{
		Title: "All",
		Image: "https://lh3.googleusercontent.com/p/no-date=w224-h398-k-no",
	}}

	got := gmaps.ExtractPhotoMetadataFromPayload([]any{"payload"}, images)

	require.Len(t, got, 1)
	require.Equal(t, gmaps.PhotoMetadataStatusUnavailable, got[0].Status)
	require.Empty(t, got[0].PostedDate)
}

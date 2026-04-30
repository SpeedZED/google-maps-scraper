package gmaps

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	PhotoMetadataStatusFound       = "found"
	PhotoMetadataStatusUnavailable = "unavailable"
	PhotoMetadataSourcePayload     = "google_maps_payload"
)

var googleImageSizeSuffix = regexp.MustCompile(`=w\d+(?:-h\d+)?(?:-[a-z0-9]+)*$`)

func ExtractPhotoMetadataFromPayload(root any, images []Image) []PhotoMetadata {
	if len(images) == 0 {
		return nil
	}

	index := make(map[string]Image, len(images)*2)
	order := make([]string, 0, len(images))
	for _, item := range images {
		if item.Image == "" {
			continue
		}
		key := normalizeGoogleImageURL(item.Image)
		if _, ok := index[key]; !ok {
			order = append(order, key)
			index[key] = item
		}
	}

	found := make(map[string]PhotoMetadata, len(index))
	walkPhotoMetadata(root, nil, index, found)

	ans := make([]PhotoMetadata, 0, len(order))
	for _, key := range order {
		item := index[key]
		meta, ok := found[key]
		if !ok {
			meta = PhotoMetadata{
				Title:  item.Title,
				Image:  item.Image,
				Source: PhotoMetadataSourcePayload,
				Status: PhotoMetadataStatusUnavailable,
			}
		}
		if meta.Title == "" {
			meta.Title = item.Title
		}
		if meta.Image == "" {
			meta.Image = item.Image
		}
		if meta.Source == "" {
			meta.Source = PhotoMetadataSourcePayload
		}
		if meta.Status == "" {
			if meta.PostedDate != "" {
				meta.Status = PhotoMetadataStatusFound
			} else {
				meta.Status = PhotoMetadataStatusUnavailable
			}
		}
		ans = append(ans, meta)
	}

	return ans
}

func walkPhotoMetadata(node any, ancestors []any, images map[string]Image, found map[string]PhotoMetadata) {
	switch v := node.(type) {
	case string:
		key := normalizeGoogleImageURL(v)
		if item, ok := images[key]; ok {
			meta := photoMetadataFromContext(item, ancestors)
			if existing, exists := found[key]; !exists || metadataCompleteness(meta) > metadataCompleteness(existing) {
				found[key] = meta
			}
		}
	case []any:
		nextAncestors := append(ancestors, v)
		for _, child := range v {
			walkPhotoMetadata(child, nextAncestors, images, found)
		}
	case map[string]any:
		nextAncestors := append(ancestors, v)
		for _, child := range v {
			walkPhotoMetadata(child, nextAncestors, images, found)
		}
	}
}

func photoMetadataFromContext(item Image, ancestors []any) PhotoMetadata {
	meta := PhotoMetadata{
		Title:  item.Title,
		Image:  item.Image,
		Source: PhotoMetadataSourcePayload,
		Status: PhotoMetadataStatusUnavailable,
	}

	// Google Maps currently keeps the structured photo date tuple in a nearby
	// payload container on the path to the photo image URL. Walk outward from
	// the URL's closest containers and only keep the stable ISO date.
	for i := len(ancestors) - 1; i >= 0; i-- {
		meta.PostedDate = findDateTuple(ancestors[i])
		if meta.PostedDate != "" || len(ancestors)-i >= 6 {
			break
		}
	}

	if meta.PostedDate != "" {
		meta.Status = PhotoMetadataStatusFound
	}

	return meta
}

func normalizeGoogleImageURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return googleImageSizeSuffix.ReplaceAllString(raw, "")
}

func metadataCompleteness(meta PhotoMetadata) int {
	score := 0
	if meta.PostedDate != "" {
		score += 4
	}
	return score
}

func findDateTuple(node any) string {
	switch v := node.(type) {
	case []any:
		if date, ok := parseDateTuple(v); ok {
			return date
		}
		for _, child := range v {
			if date := findDateTuple(child); date != "" {
				return date
			}
		}
	case map[string]any:
		for _, child := range v {
			if date := findDateTuple(child); date != "" {
				return date
			}
		}
	}
	return ""
}

func parseDateTuple(values []any) (string, bool) {
	if len(values) < 3 {
		return "", false
	}

	year, okYear := numericInt(values[0])
	month, okMonth := numericInt(values[1])
	day, okDay := numericInt(values[2])

	if !okYear || !okMonth || !okDay {
		return "", false
	}
	if year < 2005 || year > 2100 || month < 1 || month > 12 || day < 1 || day > 31 {
		return "", false
	}

	return fmt.Sprintf("%04d-%02d-%02d", year, month, day), true
}

func numericInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		if v != float64(int(v)) {
			return 0, false
		}
		return int(v), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(v))
		return i, err == nil
	default:
		return 0, false
	}
}

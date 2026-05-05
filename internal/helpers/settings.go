package helpers

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// GetSetting returns the value stored in site_settings for the given key.
// Returns "" if the key is missing or on error.
func GetSetting(pb *pocketbase.PocketBase, key string) string {
	records, err := pb.FindRecordsByFilter(
		"site_settings",
		"key='"+escapeSettingKey(key)+"'",
		"", 1, 0,
	)
	if err != nil || len(records) == 0 {
		return ""
	}
	return records[0].GetString("value")
}

// SetSetting upserts a key/value pair in site_settings. If the key does not
// exist yet, a new record is created.
func SetSetting(pb *pocketbase.PocketBase, key, value string) error {
	records, _ := pb.FindRecordsByFilter(
		"site_settings",
		"key='"+escapeSettingKey(key)+"'",
		"", 1, 0,
	)
	if len(records) == 0 {
		col, err := pb.FindCollectionByNameOrId("site_settings")
		if err != nil {
			return err
		}
		r := core.NewRecord(col)
		r.Set("key", key)
		r.Set("value", value)
		return pb.Save(r)
	}
	r := records[0]
	r.Set("value", value)
	return pb.Save(r)
}

// escapeSettingKey strips characters that would break a PocketBase filter
// expression (single quote, backslash). Setting keys are short ASCII slugs in
// practice.
func escapeSettingKey(s string) string {
	out := make([]rune, 0, len(s))
	for _, c := range s {
		if c == '\'' || c == '\\' {
			continue
		}
		out = append(out, c)
	}
	return string(out)
}

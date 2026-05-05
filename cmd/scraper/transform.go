package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

type Tienda struct {
	Nombre      string   `json:"nombre"`
	Slug        string   `json:"slug"`
	Cat         string   `json:"cat"`
	Gal         string   `json:"gal"`
	Local       string   `json:"local"`
	Logo        string   `json:"logo"`
	Tags        []string `json:"tags"`
	Descripcion string   `json:"descripcion"`
	About       string   `json:"about"`
	About2      string   `json:"about2"`
	Pay         string   `json:"pay"`
	Photos      []string `json:"photos"`
	Whatsapp    string   `json:"whatsapp"`
	Telefono    string   `json:"telefono"`
	Rating      string   `json:"rating"`
	HorarioLV   string   `json:"horario_lv"`
	HorarioSab  string   `json:"horario_sab"`
	HorarioDom  string   `json:"horario_dom"`
	Status      string   `json:"status"`
	Destacada   bool     `json:"destacada"`
}

type LegacyTienda map[string]interface{}

func transformJSON(data []byte, galeria string) ([]Tienda, error) {
	var raw []LegacyTienda
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make([]Tienda, 0, len(raw))
	for _, legacy := range raw {
		t := Tienda{
			Gal:    galeria,
			Status: "publicado",
		}

		if v, ok := legacy["nombre"].(string); ok {
			t.Nombre = v
		}
		if v, ok := legacy["name"].(string); ok && t.Nombre == "" {
			t.Nombre = v
		}
		t.Slug = slugify(t.Nombre)

		if v, ok := legacy["cat"].(string); ok {
			t.Cat = v
		}
		if v, ok := legacy["categoria"].(string); ok && t.Cat == "" {
			t.Cat = v
		}
		t.Cat = strings.ToLower(t.Cat)

		if v, ok := legacy["local"].(string); ok {
			t.Local = v
		}
		if v, ok := legacy["local_number"].(string); ok && t.Local == "" {
			t.Local = v
		}

		if v, ok := legacy["logo"].(string); ok {
			t.Logo = v
		}
		if v, ok := legacy["logo_url"].(string); ok && t.Logo == "" {
			t.Logo = v
		}

		t.Tags = extractStringSlice(legacy, "tags")
		t.Photos = extractStringSlice(legacy, "photos", "fotos", "images")

		if v, ok := legacy["descripcion"].(string); ok {
			t.Descripcion = v
		}
		if v, ok := legacy["description"].(string); ok && t.Descripcion == "" {
			t.Descripcion = v
		}

		if v, ok := legacy["about"].(string); ok {
			t.About = v
		}
		if v, ok := legacy["about2"].(string); ok {
			t.About2 = v
		}
		if v, ok := legacy["pay"].(string); ok {
			t.Pay = v
		}
		if v, ok := legacy["pago"].(string); ok && t.Pay == "" {
			t.Pay = v
		}

		if v, ok := legacy["whatsapp"].(string); ok {
			t.Whatsapp = v
		}
		if v, ok := legacy["telefono"].(string); ok {
			t.Telefono = v
		}
		if v, ok := legacy["phone"].(string); ok && t.Telefono == "" {
			t.Telefono = v
		}
		if v, ok := legacy["rating"].(string); ok {
			t.Rating = v
		}

		if v, ok := legacy["horario_lv"].(string); ok {
			t.HorarioLV = v
		}
		if v, ok := legacy["horario_sab"].(string); ok {
			t.HorarioSab = v
		}
		if v, ok := legacy["horario_dom"].(string); ok {
			t.HorarioDom = v
		}

		if v, ok := legacy["destacada"].(bool); ok {
			t.Destacada = v
		}
		if v, ok := legacy["featured"].(bool); ok && !t.Destacada {
			t.Destacada = v
		}

		if v, ok := legacy["gal"].(string); ok && v != "" {
			normalized := normalizeGal(v)
			if normalized != "" {
				t.Gal = normalized
			}
		}

		if t.Nombre == "" {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

var multiDashRe = regexp.MustCompile(`-+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '-'
	}, s)
	s = multiDashRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func normalizeGal(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if strings.Contains(v, "placa") || strings.Contains(v, "comercial") {
		return "placa-comercial"
	}
	if strings.Contains(v, "flamenco") || strings.Contains(v, "torre") {
		return "torre-flamenco"
	}
	return ""
}

func extractStringSlice(m map[string]interface{}, keys ...string) []string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case []interface{}:
				result := make([]string, 0, len(val))
				for _, item := range val {
					if s, ok := item.(string); ok && s != "" {
						result = append(result, s)
					}
				}
				return result
			case string:
				if val != "" {
					return []string{val}
				}
			}
		}
	}
	return []string{}
}

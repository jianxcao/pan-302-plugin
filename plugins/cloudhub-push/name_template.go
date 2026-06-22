package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

const (
	movieNameTemplateText = `{{.title}}{{if .year}} ({{.year}}){{end}}/{{.title}}{{if .year}} ({{.year}}){{end}}{{if .part}}-{{.part}}{{end}}{{if .videoFormat}} - {{.videoFormat}}{{end}}{{if .fps}} {{.fps}}fps{{end}}{{if .effect}} {{.effect}}{{end}}{{if .videoCodec}} {{.videoCodec}}{{end}}{{if .audioCodec}} {{.audioCodec}}{{end}}{{.fileExt}}`
	tvNameTemplateText    = `{{.title}}{{if .year}} ({{.year}}){{end}}/Season {{.season}}/{{.title}} {{.season_episode}}{{if .part}}-{{.part}}{{end}}{{if .videoFormat}} - {{.videoFormat}}{{end}}{{if .fps}} {{.fps}}fps{{end}}{{if .effect}} {{.effect}}{{end}}{{if .videoCodec}} {{.videoCodec}}{{end}}{{.fileExt}}`
)

var (
	movieNameTemplate = template.Must(template.New("movie-name").Parse(movieNameTemplateText))
	tvNameTemplate    = template.Must(template.New("tv-name").Parse(tvNameTemplateText))
	seasonDirPattern  = regexp.MustCompile(`(?i)^(?:season\s*0*(\d{1,3})|第\s*0*(\d{1,3})\s*季)$`)
	titleYearPattern  = regexp.MustCompile(`(?:^|[\s._(\[-])(19\d{2}|20\d{2})(?:$|[\s._)\]-])`)
)

func applyCloudHubRecognizableName(resource *Resource) {
	if resource == nil {
		return
	}
	data := cloudHubNameData(resource)
	title := strings.TrimSpace(data["title"].(string))
	if title == "" || strings.TrimSpace(data["fileExt"].(string)) == "" {
		return
	}
	tpl := movieNameTemplate
	if resource.Type == "tv" {
		tpl = tvNameTemplate
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return
	}
	resource.Title = title
	resource.Name = lastPathSegment(out.String())
}

func cloudHubNameData(resource *Resource) map[string]any {
	title, year, pathSeason := titleYearSeasonFromPath(resource)
	if title == "" {
		title = cleanMediaTitle(resource.Title, resource.Type, resource.Year)
	}
	if year == "" {
		year = resource.Year
	}
	season := resource.Season
	if season == 0 {
		season = pathSeason
	}
	fileExt := filepath.Ext(resource.Name)
	if fileExt == "" {
		fileExt = filepath.Ext(resource.Path)
	}
	if fileExt == "" && resource.Container != "" {
		fileExt = "." + strings.TrimPrefix(strings.ToLower(resource.Container), ".")
	}
	return map[string]any{
		"title":          title,
		"year":           year,
		"season":         season,
		"season_episode": seasonEpisodeToken(resource.Season, resource.Episode),
		"part":           "",
		"videoFormat":    videoFormat(resource),
		"fps":            formatFPS(resource.FPS),
		"effect":         videoEffect(resource),
		"videoCodec":     formatCodec(resource.VideoCodec),
		"audioCodec":     formatCodec(resource.AudioCodec),
		"fileExt":        fileExt,
	}
}

func titleYearSeasonFromPath(resource *Resource) (string, string, int) {
	segments := splitPathSegments(resource.Path)
	if len(segments) == 0 {
		return "", "", 0
	}
	if resource.Type == "tv" {
		for i, segment := range segments {
			season := parseSeasonDir(segment)
			if season > 0 && i > 0 {
				title, year := splitTitleYear(segments[i-1])
				return title, year, season
			}
		}
	}
	if len(segments) >= 2 {
		parent := segments[len(segments)-2]
		if title, year := splitTitleYear(parent); title != "" && year != "" {
			return title, year, 0
		}
	}
	return "", "", 0
}

func splitPathSegments(value string) []string {
	cleaned := strings.Trim(normalizeCloudHubPath(value), "/")
	if cleaned == "" {
		return nil
	}
	parts := strings.Split(cleaned, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func lastPathSegment(value string) string {
	segments := splitPathSegments(value)
	if len(segments) == 0 {
		return strings.TrimSpace(value)
	}
	return segments[len(segments)-1]
}

func parseSeasonDir(value string) int {
	matches := seasonDirPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(matches) == 0 {
		return 0
	}
	for _, value := range matches[1:] {
		if value == "" {
			continue
		}
		season, _ := strconv.Atoi(value)
		return season
	}
	return 0
}

func splitTitleYear(value string) (string, string) {
	base := strings.TrimSuffix(strings.TrimSpace(value), filepath.Ext(value))
	matches := titleYearPattern.FindStringSubmatch(base)
	if len(matches) != 2 {
		return cleanSeparators(base), ""
	}
	year := matches[1]
	index := strings.Index(base, year)
	if index < 0 {
		return cleanSeparators(base), year
	}
	return cleanSeparators(strings.TrimSpace(strings.TrimRight(base[:index], "([{- ._"))), year
}

func cleanMediaTitle(value, mediaType, year string) string {
	base := strings.TrimSuffix(strings.TrimSpace(value), filepath.Ext(value))
	if mediaType == "tv" {
		if loc := seasonEpisodePattern.FindStringIndex(base); loc != nil {
			base = strings.TrimSpace(base[:loc[0]])
		}
	}
	if year != "" {
		if index := strings.Index(base, year); index >= 0 {
			base = strings.TrimSpace(base[:index])
		}
	}
	return cleanSeparators(base)
}

func cleanSeparators(value string) string {
	value = strings.ReplaceAll(value, "_", " ")
	value = strings.ReplaceAll(value, ".", " ")
	value = strings.Trim(value, " -_.()[]{}")
	for strings.Contains(value, "  ") {
		value = strings.ReplaceAll(value, "  ", " ")
	}
	return strings.TrimSpace(value)
}

func seasonEpisodeToken(season, episode int) string {
	if season <= 0 || episode <= 0 {
		return ""
	}
	return fmt.Sprintf("S%02dE%02d", season, episode)
}

func videoFormat(resource *Resource) string {
	if resource.VideoResolution != "" {
		return resource.VideoResolution
	}
	for _, token := range strings.Fields(resource.Quality) {
		lower := strings.ToLower(token)
		if strings.HasSuffix(lower, "p") || lower == "4k" || lower == "uhd" {
			return token
		}
	}
	return ""
}

func videoEffect(resource *Resource) string {
	if resource.VideoHDR != "" {
		return resource.VideoHDR
	}
	quality := strings.ToUpper(resource.Quality)
	switch {
	case strings.Contains(quality, "DV"):
		return "DV"
	case strings.Contains(quality, "HDR"):
		return "HDR"
	default:
		return ""
	}
}

func formatCodec(value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "":
		return ""
	case "h264", "h.264", "avc":
		return "H264"
	case "h265", "h.265", "hevc":
		return "HEVC"
	case "eac3", "e-ac-3":
		return "EAC3"
	case "aac":
		return "AAC"
	case "av1":
		return "AV1"
	default:
		return strings.ToUpper(value)
	}
}

func formatFPS(value float64) string {
	if value <= 0 {
		return ""
	}
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", value), "0"), ".")
}

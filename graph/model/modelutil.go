package model

import (
	"strings"
)

func SortableTitle(s string) string {
	s = strings.TrimPrefix(s, "The ")
	s = strings.TrimPrefix(s, "A ")
	s = strings.TrimPrefix(s, "An ")
	return s
}

type ByVideoTitle []*Video

func (a ByVideoTitle) Len() int           { return len(a) }
func (a ByVideoTitle) Less(i, j int) bool { return a[i].SortTitle < a[j].SortTitle }
func (a ByVideoTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type BySeriesName []*Series

func (a BySeriesName) Len() int           { return len(a) }
func (a BySeriesName) Less(i, j int) bool { return a[i].SortName < a[j].SortName }
func (a BySeriesName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByEpisode []*Episode

func (a ByEpisode) Len() int { return len(a) }
func (a ByEpisode) Less(i, j int) bool {
	if a[i].Series.SortName != a[j].Series.SortName {
		return a[i].Series.SortName < a[j].Series.SortName
	}

	if a[i].Season != a[j].Season {
		return a[i].Season < a[j].Season
	}

	return a[i].Episode < a[j].Episode
}
func (a ByEpisode) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

package model

import (
	"strings"
)

func trimArticle(s string) string {
	s = strings.TrimPrefix(s, "The ")
	s = strings.TrimPrefix(s, "A ")
	s = strings.TrimPrefix(s, "An ")
	return s
}

type ByVideoTitle []*Video

func (a ByVideoTitle) Len() int           { return len(a) }
func (a ByVideoTitle) Less(i, j int) bool { return a[i].SortableTitle() < a[j].SortableTitle() }
func (a ByVideoTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (v *Video) SortableTitle() string {
	if v.SortTitle != nil {
		return *v.SortTitle
	}
	return trimArticle(v.Title)
}

type BySeriesName []*Series

func (a BySeriesName) Len() int           { return len(a) }
func (a BySeriesName) Less(i, j int) bool { return a[i].SortableName() < a[j].SortableName() }
func (a BySeriesName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (s *Series) SortableName() string {
	if s.SortName != nil {
		return *s.SortName
	}
	return trimArticle(s.Name)
}

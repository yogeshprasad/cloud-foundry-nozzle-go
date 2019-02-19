package nozzle

import (
	"strings"

	"github.com/gobwas/glob"
)

type Filter interface {
	Match(name string, tags map[string]string) bool
}

type globFilter struct {
	metricWhitelist    glob.Glob
	metricBlacklist    glob.Glob
	metricTagWhitelist map[string]glob.Glob
	metricTagBlacklist map[string]glob.Glob
	// tagInclude         glob.Glob
	// tagExclude         glob.Glob
}

func NewGlobFilter(cfg *FiltersConfig) Filter {
	return &globFilter{
		metricWhitelist:    compile(cfg.MetricsWhiteList),
		metricBlacklist:    compile(cfg.MetricsBlackList),
		metricTagWhitelist: multiCompile(cfg.MetricsTagWhiteList),
		metricTagBlacklist: multiCompile(cfg.MetricsTagBlackList),
		// tagInclude:         compile(cfg.TagInclude),
		// tagExclude:         compile(cfg.TagExclude),
	}
}

func compile(filters []string) glob.Glob {
	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		g, _ := glob.Compile(filters[0])
		return g
	}
	g, _ := glob.Compile("{" + strings.Join(filters, ",") + "}")
	return g
}

func multiCompile(filters map[string][]string) map[string]glob.Glob {
	if len(filters) == 0 {
		return nil
	}
	globs := make(map[string]glob.Glob, len(filters))
	for k, v := range filters {
		g := compile(v)
		if g != nil {
			globs[k] = g
		}
	}
	return globs
}

func (gf *globFilter) Match(name string, tags map[string]string) bool {
	if gf.metricWhitelist != nil && !gf.metricWhitelist.Match(name) {
		return false
	}
	if gf.metricBlacklist != nil && gf.metricBlacklist.Match(name) {
		return false
	}

	if gf.metricTagWhitelist != nil && !matchesTags(gf.metricTagWhitelist, tags) {
		return false
	}
	if gf.metricTagBlacklist != nil && matchesTags(gf.metricTagBlacklist, tags) {
		return false
	}

	// if gf.tagInclude != nil {
	// 	deleteTags(gf.tagInclude, tags, true)
	// }
	// if gf.tagExclude != nil {
	// 	deleteTags(gf.tagExclude, tags, false)
	// }
	return true
}

func matchesTags(matchers map[string]glob.Glob, tags map[string]string) bool {
	for k, matcher := range matchers {
		if val, ok := tags[k]; ok {
			if matcher.Match(val) {
				return true
			}
		}
	}
	return false
}

func matchesTag(matcher glob.Glob, tags map[string]string) bool {
	for k := range tags {
		if matcher.Match(k) {
			return true
		}
	}
	return false
}

func deleteTags(matcher glob.Glob, tags map[string]string, include bool) {
	for k := range tags {
		matches := matcher.Match(k)
		if (include && !matches) || (!include && matches) {
			delete(tags, k)
		}
	}
}

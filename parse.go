package main

import (
	"fmt"
	"strings"
)

func tagFromComment(comment string) (tag string) {
	match := rComment.FindStringSubmatch(comment)
	if len(match) == 2 {
		tag = match[1]
	}
	return
}
func makePointer(comment string) bool {
	match := rPointer.FindStringSubmatch(comment)
	return len(match) == 1
}

type tagItem struct {
	key   string
	value string
}

type tagItems []tagItem

func (ti tagItems) format() string {
	var tags []string
	for _, item := range ti {
		tags = append(tags, fmt.Sprintf(`%s:%s`, item.key, item.value))
	}
	return strings.Join(tags, " ")
}

func (ti tagItems) override(nti tagItems) tagItems {
	var override []tagItem
	for i := range ti {
		var dup = -1
		for j := range nti {
			if ti[i].key == nti[j].key {
				dup = j
				break
			}
		}
		if dup == -1 {
			override = append(override, ti[i])
		} else {
			override = append(override, nti[dup])
			nti = append(nti[:dup], nti[dup+1:]...)
		}
	}
	return append(override, nti...)
}

func newTagItems(tag string) tagItems {
	var items []tagItem
	split := rTags.FindAllString(tag, -1)

	for _, t := range split {
		sepPos := strings.Index(t, ":")
		items = append(items, tagItem{
			key:   t[:sepPos],
			value: t[sepPos+1:],
		})
	}
	return items
}

func injectTag(contents []byte, area fieldInfo) (injected []byte, offset int) {
	expr := make([]byte, area.End-area.Start)
	copy(expr, contents[area.Start-1:area.End-1])
	cti := newTagItems(area.CurrentTag)
	iti := newTagItems(*area.InjectTag)
	ti := cti.override(iti)
	expr = rInject.ReplaceAll(expr, []byte(fmt.Sprintf("`%s`", ti.format())))

	if area.MakePointer {
		injected = append(injected, contents[:area.TypePos-1]...)
		injected = append(injected, byte('*'))
	} else {
		injected = append(injected, contents[:area.Start-1]...)
	}

	injected = append(injected, expr...)
	injected = append(injected, contents[area.End-1:]...)

	return
}

package main

import (
	"fmt"
	"github.com/jucardi/go-logger-lib/log"
	"github.com/jucardi/go-streams/streams"
	"github.com/jucardi/go-strings/stringx"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var (
	rComment = regexp.MustCompile(`^//\s*@inject_tag:\s*(.*)$`)
	rPointer = regexp.MustCompile(`^//\s*@pointer$`)
	rInject  = regexp.MustCompile("`.+`$")
	rTags    = regexp.MustCompile(`[\w_]+:"[^"]+"`)
)

type fieldInfo struct {
	Start       int
	End         int
	Name        string
	TypePos     int
	CurrentTag  string
	InjectTag   *string
	MakePointer bool
}

func parseFile(inputPath string, xxxSkip []string) (areas []fieldInfo, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, inputPath, nil, parser.ParseComments)
	if err != nil {
		return
	}

	for _, decl := range f.Decls {
		// check if is generic declaration
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		var typeSpec *ast.TypeSpec
		for _, spec := range genDecl.Specs {
			if ts, tsOK := spec.(*ast.TypeSpec); tsOK {
				typeSpec = ts
				break
			}
		}

		// skip if can't get type spec
		if typeSpec == nil {
			continue
		}

		// not a struct, skip
		structDecl, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		builder := strings.Builder{}
		if len(xxxSkip) > 0 {
			for i, skip := range xxxSkip {
				builder.WriteString(fmt.Sprintf(`%s:"-"`, skip))
				if i > 0 {
					builder.WriteString(",")
				}
			}
		}

		for _, field := range structDecl.Fields.List {
			// skip if field has no doc
			if len(field.Names) > 0 {
				name := field.Names[0].Name
				if len(xxxSkip) > 0 && strings.HasPrefix(name, "XXX") {
					currentTag := field.Tag.Value
					newTag := builder.String()
					area := fieldInfo{
						Start:      int(field.Pos()),
						End:        int(field.End()),
						CurrentTag: currentTag[1 : len(currentTag)-1],
						InjectTag:  &newTag,
					}
					areas = append(areas, area)
				}
			}
			if field.Doc == nil {
				continue
			}
			var tags []string
			var isPointer bool

			for _, comment := range field.Doc.List {
				if makePointer(comment.Text) {
					isPointer = true
				}

				tag := tagFromComment(comment.Text)
				if tag == "" {
					continue
				}
				tags = append(tags, tag)
			}
			names := streams.From(field.Names).Map(func(i interface{}) interface{} {
				x := i.(*ast.Ident)
				return x.Name
			}).ToArray().([]string)

			if len(tags) > 0 {
				currentTag := field.Tag.Value
				newTag := strings.Join(tags, " ")
				area := fieldInfo{
					Start:       int(field.Type.Pos()),
					End:         int(field.Tag.End()),
					Name:        strings.Join(names, ", "),
					TypePos:     int(field.Type.Pos()),
					CurrentTag:  currentTag[1 : len(currentTag)-1],
					InjectTag:   &newTag,
					MakePointer: isPointer,
				}
				areas = append(areas, area)
			}
		}
	}
	log.Debugf("parsed file '%s', number of fields to inject custom tags: %d", inputPath, len(areas))
	return
}

func writeFile(inputPath string, areas []fieldInfo) (err error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	if err = f.Close(); err != nil {
		return
	}

	offset := 0
	// inject custom tags from tail of file first to preserve order
	for i := range areas {
		area := areas[len(areas)-i-1]
		if area.InjectTag != nil {
			log.Debugf("Injecting custom tag `%s` to '%s'", *area.InjectTag, area.Name)
		}
		area.Start += offset
		area.End += offset
		contents, offset = injectTag(contents, area)
	}

	if cleanup {
		str := string(contents)
		lines := streams.
			From(strings.Split(str, "\n")).
			Filter(func(i interface{}) bool {
				x := i.(string)
				str := stringx.New(x).TrimSpace().Trim("\t").S()
				return tagFromComment(str) == "" && !makePointer(str)
			}).
			ToArray().([]string)

		contents = []byte(strings.Join(lines, "\n"))
	}

	if err = ioutil.WriteFile(inputPath, contents, 0644); err != nil {
		return
	}

	return
}

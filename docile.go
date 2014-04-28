package docile

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var packages = map[string]map[string]string{}

func Get(pack, key string) (result string, found bool) {
	p, found := packages[pack]
	if !found {
		return
	}
	result, found = p[key]
	return
}

func Add(pack, key, doc string) {
	p, found := packages[pack]
	if !found {
		p = map[string]string{}
		packages[pack] = p
	}
	p[key] = doc
}

var funcReg = regexp.MustCompile("^func ")
var typeReg = regexp.MustCompile("^TYPES")

type DocObj struct {
	Pack    string
	Name    string
	Comment string
}
type TmplObj struct {
	Package string
	Docs    []DocObj
}

var tmpl = template.Must(template.New("template").Parse(`
package {{.Package}}
import "github.com/zond/docile"
func init() {
{{range .Docs}}
  docile.Add("{{.Pack}}", "{{.Name}}", "{{.Comment}}")
{{end}}
}
`))

func Generate(pack string, dst string) (err error) {

	if err = os.Remove(dst); err != nil {
		if os.IsNotExist(err) {
			err = nil
		} else {
			return
		}
	}

	ctx := TmplObj{
		filepath.Base(filepath.Dir(dst)),
		[]DocObj{},
	}
	var pkgs map[string]*ast.Package
	for _, godir := range strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator)) {
		if pkgs, err = parser.ParseDir(&token.FileSet{}, filepath.Join(godir, "src", pack), nil, parser.ParseComments); err == nil {
			for _, pkg := range pkgs {
				docPack := doc.New(pkg, filepath.Join(godir, "src", pack), 0)
				for _, f := range docPack.Funcs {
					if strings.TrimSpace(f.Doc) != "" {
						ctx.Docs = append(ctx.Docs, DocObj{pack, f.Name, strings.TrimSpace(f.Doc)})
					}
				}
			}
		}
		if os.IsNotExist(err) {
			err = nil
		}
	}

	if len(ctx.Docs) < 1 {
		return
	}

	dst, err = filepath.Abs(dst)
	if err != nil {
		return
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	err = tmpl.Execute(dstFile, ctx)
	if err != nil {
		return
	}

	return
}

package docile

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func Generate(pack string, dst string) (err error) {
	dst, err = filepath.Abs(dst)
	if err != nil {
		return
	}
	dstFileName := filepath.Join(os.TempDir(), fmt.Sprintf("docile-generated-%v.go", rand.Int63()))
	dstFile, err := os.Create(dstFileName)
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			if err = dstFile.Close(); err == nil {
				err = os.Rename(dstFileName, dst)
			}
		}
	}()
	if _, err = fmt.Fprintf(dstFile, `package %v
import "github.com/zond/docile"
func init() {
`, filepath.Base(filepath.Dir(dst))); err != nil {
		return
	}

	var pkgs map[string]*ast.Package
	for _, godir := range strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator)) {
		if pkgs, err = parser.ParseDir(&token.FileSet{}, filepath.Join(godir, "src", pack), nil, parser.ParseComments); err == nil {
			for _, pkg := range pkgs {
				docPack := doc.New(pkg, filepath.Join(godir, "src", pack), 0)
				for _, f := range docPack.Funcs {
					if strings.TrimSpace(f.Doc) != "" {
						if _, err = fmt.Fprintf(dstFile, "  docile.Add(%#v, %#v, %#v)\n", pack, f.Name, strings.TrimSpace(f.Doc)); err != nil {
							return
						}
					}
				}
			}
		}
		if os.IsNotExist(err) {
			err = nil
		}
	}
	if _, err = fmt.Fprintln(dstFile, "}"); err != nil {
		return
	}
	return

}

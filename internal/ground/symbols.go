// Package ground checks project facts against the code they point at, and
// flags the ones the code no longer backs up.
package ground

import (
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"regexp"
	"strings"
)

// Def is one definition in a source file: a function, method, type, macro,
// constant or variable, with the lines it spans (1-based, inclusive).
type Def struct {
	Name  string
	Start int
	End   int
}

// lang is how a file's comments and definitions are written.
type lang int

const (
	langOther lang = iota
	langGo
	langC
	langPython
	langJS
	langRust
)

func langOf(p string) lang {
	switch strings.ToLower(path.Ext(p)) {
	case ".go":
		return langGo
	case ".c", ".h", ".cc", ".cpp", ".cxx", ".hpp", ".hh", ".hxx", ".ino", ".m", ".mm":
		return langC
	case ".py":
		return langPython
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".mts", ".cts":
		return langJS
	case ".rs":
		return langRust
	}
	return langOther
}

// SymbolLanguages names the languages where Woolly finds definitions, not
// just mentions. Other files are checked by searching for the name.
const SymbolLanguages = "Go, C, C++, Python, JavaScript, TypeScript and Rust"

// Defs lists the definitions in a source file.
func Defs(p string, src []byte) []Def {
	text := strings.ReplaceAll(string(src), "\r\n", "\n")
	switch langOf(p) {
	case langGo:
		if defs, ok := goDefs(text); ok {
			return defs
		}
		return braceDefs(text, goRes)
	case langC:
		return cDefs(text)
	case langPython:
		return pyDefs(text)
	case langJS:
		return braceDefs(text, jsRes)
	case langRust:
		return braceDefs(text, rustRes)
	}
	return nil
}

func goDefs(text string) ([]Def, bool) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", text, parser.SkipObjectResolution)
	if err != nil {
		return nil, false
	}
	line := func(p token.Pos) int { return fset.Position(p).Line }
	var out []Def
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			name := d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				if recv := recvName(d.Recv.List[0].Type); recv != "" {
					name = recv + "." + name
				}
			}
			start := line(d.Pos())
			if d.Doc != nil {
				start = line(d.Type.Pos())
			}
			out = append(out, Def{Name: name, Start: start, End: line(d.End())})
		case *ast.GenDecl:
			for _, s := range d.Specs {
				switch s := s.(type) {
				case *ast.TypeSpec:
					out = append(out, Def{Name: s.Name.Name, Start: line(s.Pos()), End: line(s.End())})
				case *ast.ValueSpec:
					for _, n := range s.Names {
						out = append(out, Def{Name: n.Name, Start: line(s.Pos()), End: line(s.End())})
					}
				}
			}
		}
	}
	return out, true
}

func recvName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return recvName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return recvName(t.X)
	case *ast.IndexListExpr:
		return recvName(t.X)
	}
	return ""
}

var (
	goRes = []*regexp.Regexp{
		regexp.MustCompile(`^func\s+(?:\([^)]*\)\s*)?(\w+)`),
		regexp.MustCompile(`^(?:type|const|var)\s+(\w+)`),
	}
	jsRes = []*regexp.Regexp{
		regexp.MustCompile(`^\s*(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s*\*?\s*(\w+)`),
		regexp.MustCompile(`^\s*(?:export\s+)?(?:default\s+)?(?:abstract\s+)?class\s+(\w+)`),
		regexp.MustCompile(`^\s*(?:export\s+)?(?:declare\s+)?(?:const|let|var)\s+(\w+)`),
		regexp.MustCompile(`^\s*(?:export\s+)?(?:declare\s+)?(?:interface|type|enum|namespace)\s+(\w+)`),
		regexp.MustCompile(`^\s+(?:(?:public|private|protected|static|async|readonly|override|get|set)\s+)*(\w+)\s*(?:<[^>]*>)?\([^)]*\)\s*(?::[^{=]*)?\{`),
	}
	rustRes = []*regexp.Regexp{
		regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?(?:unsafe\s+)?(?:const\s+)?(?:extern\s+"[^"]*"\s+)?fn\s+(\w+)`),
		regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?(?:struct|enum|trait|type|mod|union)\s+(\w+)`),
		regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?(?:const|static)\s+(?:mut\s+)?(\w+)`),
		regexp.MustCompile(`^\s*macro_rules!\s*(\w+)`),
	}
	keywords = map[string]bool{"if": true, "for": true, "while": true, "switch": true, "return": true, "catch": true, "else": true, "do": true, "sizeof": true, "case": true, "new": true, "delete": true, "throw": true}
)

// braceDefs finds definitions line by line and ends each at its closing
// brace, or at the end of its statement when it has no body.
func braceDefs(text string, res []*regexp.Regexp) []Def {
	lines := strings.Split(text, "\n")
	var out []Def
	for i, l := range lines {
		for _, re := range res {
			m := re.FindStringSubmatch(l)
			if m == nil || keywords[m[1]] {
				continue
			}
			out = append(out, Def{Name: m[1], Start: i + 1, End: blockEnd(lines, i)})
			break
		}
	}
	return out
}

var (
	cDefine   = regexp.MustCompile(`^\s*#\s*define\s+(\w+)`)
	cTagged   = regexp.MustCompile(`^\s*(?:typedef\s+)?(?:struct|enum|union|class|namespace)\s+(\w+)\s*(?::[^{;]*)?(?:\{|$)`)
	cTypedef  = regexp.MustCompile(`^\s*typedef\s+(?:struct|enum|union)\b`)
	cTypeName = regexp.MustCompile(`^\s*}\s*(\w+)\s*;`)
	cFunc     = regexp.MustCompile(`^[A-Za-z_][\w\s\*&:<>,~]*?\b([A-Za-z_]\w*(?:::~?\w+)?)\s*\(`)
	cValue    = regexp.MustCompile(`^(?:(?:static|const|constexpr|extern|volatile|inline|unsigned|signed)\s+)*[A-Za-z_][\w:<>]*[\s\*&]+([A-Za-z_]\w*)\s*(?:\[[^\]]*\])*\s*(?:=|;)`)
	cEnumItem = regexp.MustCompile(`^\s*([A-Za-z_]\w*)\s*(?:=[^,]*)?,?\s*(?://.*|/\*.*)?$`)
	cEnumHead = regexp.MustCompile(`^\s*(?:typedef\s+)?enum\b`)
)

func cDefs(text string) []Def {
	lines := strings.Split(text, "\n")
	var out []Def
	inEnum, enumDepth, depth := false, 0, 0
	var typedefStart []int
	for i, l := range lines {
		code := stripLine(l, langC)
		switch {
		case cDefine.MatchString(l):
			end := i
			for end < len(lines)-1 && strings.HasSuffix(strings.TrimRight(lines[end], " \t"), `\`) {
				end++
			}
			out = append(out, Def{Name: cDefine.FindStringSubmatch(l)[1], Start: i + 1, End: end + 1})
		case cTagged.MatchString(code):
			out = append(out, Def{Name: cTagged.FindStringSubmatch(code)[1], Start: i + 1, End: blockEnd(lines, i)})
		case depth == 0 && cFunc.MatchString(code) && !strings.HasSuffix(strings.TrimSpace(code), ";") && opensBody(lines, i):
			name := cFunc.FindStringSubmatch(code)[1]
			if !keywords[name] {
				out = append(out, Def{Name: name, Start: i + 1, End: blockEnd(lines, i)})
				if j := strings.LastIndex(name, "::"); j >= 0 {
					out = append(out, Def{Name: name[j+2:], Start: i + 1, End: blockEnd(lines, i)})
				}
			}
		case depth == 0 && cValue.MatchString(code) && !strings.HasPrefix(strings.TrimSpace(code), "return"):
			out = append(out, Def{Name: cValue.FindStringSubmatch(code)[1], Start: i + 1, End: stmtEnd(lines, i)})
		case inEnum && depth == enumDepth && cEnumItem.MatchString(code) && strings.TrimSpace(code) != "":
			out = append(out, Def{Name: cEnumItem.FindStringSubmatch(code)[1], Start: i + 1, End: i + 1})
		}
		if cTypedef.MatchString(code) {
			typedefStart = append(typedefStart, i)
		}
		if cEnumHead.MatchString(code) {
			inEnum, enumDepth = true, depth+1
		}
		depth += strings.Count(code, "{") - strings.Count(code, "}")
		if depth < 0 {
			depth = 0
		}
		if m := cTypeName.FindStringSubmatch(code); m != nil && len(typedefStart) > 0 && depth == 0 {
			out = append(out, Def{Name: m[1], Start: typedefStart[len(typedefStart)-1] + 1, End: i + 1})
			typedefStart = typedefStart[:len(typedefStart)-1]
		}
		if inEnum && depth < enumDepth {
			inEnum = false
		}
	}
	return out
}

// opensBody reports whether a function header at line i has a body: a "{"
// comes before any ";".
func opensBody(lines []string, i int) bool {
	for j := i; j < len(lines) && j < i+6; j++ {
		code := stripLine(lines[j], langC)
		b, s := strings.Index(code, "{"), strings.Index(code, ";")
		switch {
		case b >= 0 && (s < 0 || b < s):
			return true
		case s >= 0:
			return false
		}
	}
	return false
}

// blockEnd is the line (1-based) where the block opened at or after line i
// closes, or where the statement ends if it has no block.
func blockEnd(lines []string, i int) int {
	depth, opened := 0, false
	for j := i; j < len(lines); j++ {
		code := stripLine(lines[j], langC)
		for _, r := range code {
			switch r {
			case '{':
				depth++
				opened = true
			case '}':
				depth--
				if opened && depth <= 0 {
					return j + 1
				}
			case ';':
				if !opened {
					return j + 1
				}
			}
		}
		if !opened && j > i+8 {
			return i + 1
		}
	}
	return i + 1
}

func stmtEnd(lines []string, i int) int {
	for j := i; j < len(lines) && j < i+200; j++ {
		if strings.Contains(stripLine(lines[j], langC), ";") {
			return j + 1
		}
	}
	return i + 1
}

var (
	pyDef  = regexp.MustCompile(`^(\s*)(?:async\s+)?(?:def|class)\s+(\w+)`)
	pyName = regexp.MustCompile(`^([A-Za-z_]\w*)\s*(?::[^=]*)?=[^=]`)
)

func pyDefs(text string) []Def {
	lines := strings.Split(text, "\n")
	var out []Def
	for i, l := range lines {
		if m := pyDef.FindStringSubmatch(l); m != nil {
			indent := len(m[1])
			end := i
			for j := i + 1; j < len(lines); j++ {
				t := strings.TrimSpace(lines[j])
				if t == "" || strings.HasPrefix(t, "#") {
					continue
				}
				if len(lines[j])-len(strings.TrimLeft(lines[j], " \t")) <= indent {
					break
				}
				end = j
			}
			out = append(out, Def{Name: m[2], Start: i + 1, End: end + 1})
		} else if m := pyName.FindStringSubmatch(l); m != nil {
			out = append(out, Def{Name: m[1], Start: i + 1, End: i + 1})
		}
	}
	return out
}

// Find returns the definition of sym, which may be qualified like
// "Store.Remember" or "Motor::spin".
func Find(defs []Def, sym string) (Def, bool) {
	for _, d := range defs {
		if d.Name == sym {
			return d, true
		}
	}
	short := sym
	if i := strings.LastIndexAny(sym, ".:"); i >= 0 {
		short = sym[i+1:]
	}
	for _, d := range defs {
		name := d.Name
		if i := strings.LastIndexAny(name, ".:"); i >= 0 {
			name = name[i+1:]
		}
		if name == short {
			return d, true
		}
	}
	return Def{}, false
}

// mention finds the first line where name appears as a whole word.
func mention(text, name string) (Def, bool) {
	re, err := regexp.Compile(`(^|[^\w])` + regexp.QuoteMeta(name) + `($|[^\w])`)
	if err != nil {
		return Def{}, false
	}
	for i, l := range strings.Split(text, "\n") {
		if re.MatchString(l) {
			return Def{Name: name, Start: i + 1, End: i + 1}, true
		}
	}
	return Def{}, false
}

// Lines returns lines start to end (1-based, inclusive) of text.
func Lines(text string, start, end int) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if end < start {
		end = start
	}
	if start < 1 || start > len(lines) {
		return ""
	}
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start-1:end], "\n")
}

var (
	blockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
	stringLit    = regexp.MustCompile(`"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)'`)
)

// stripLine removes a line's comment and string contents, for counting braces.
func stripLine(l string, lg lang) string {
	l = stringLit.ReplaceAllString(l, `""`)
	l = blockComment.ReplaceAllString(l, "")
	if lg == langPython {
		if i := strings.Index(l, "#"); i >= 0 {
			l = l[:i]
		}
		return l
	}
	if i := strings.Index(l, "//"); i >= 0 {
		l = l[:i]
	}
	if i := strings.Index(l, "/*"); i >= 0 {
		l = l[:i]
	}
	return l
}

// normalize drops comments and spacing, so edits to comments or layout
// don't count as changes to the code.
func normalize(text string, lg lang) string {
	if lg != langPython && lg != langOther {
		text = blockComment.ReplaceAllString(text, " ")
	}
	var b strings.Builder
	for _, l := range strings.Split(text, "\n") {
		if lg == langPython || lg == langOther {
			if t := strings.TrimSpace(l); strings.HasPrefix(t, "#") && lg == langPython {
				continue
			}
		} else if i := strings.Index(l, "//"); i >= 0 && !strings.Contains(l[:i], `"`) {
			l = l[:i]
		}
		b.WriteString(strings.Join(strings.Fields(l), " "))
		b.WriteString(" ")
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// Fingerprint is a short hash of code, ignoring comments and spacing.
func Fingerprint(p, code string) string {
	h := sha256.Sum256([]byte(normalize(code, langOf(p))))
	return hex.EncodeToString(h[:4])
}

// shape is the fingerprint of code with one name blanked out, so a renamed
// function still matches its old self.
func shape(p, code, name string) string {
	if name != "" {
		short := name
		if i := strings.LastIndexAny(name, ".:"); i >= 0 {
			short = name[i+1:]
		}
		code = regexp.MustCompile(`\b`+regexp.QuoteMeta(short)+`\b`).ReplaceAllString(code, "ǂ")
	}
	return Fingerprint(p, code)
}

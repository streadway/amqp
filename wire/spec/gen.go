package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var (
	ErrUnknownType   = errors.New("Unknown field type in gen")
	ErrUnknownDomain = errors.New("Unknown domain type in gen")
)

var amqpTypeToNative = map[string]string{
	"bit":        "bool",
	"octet":      "byte",
	"shortshort": "uint8",
	"short":      "uint16",
	"long":       "uint32",
	"longlong":   "uint64",
	"timestamp":  "Timestamp",
	"table":      "Table",
	"shortstr":   "string",
	"longstr":    "string",
}

type Rule struct {
	Name string   `xml:"name,attr"`
	Docs []string `xml:"doc"`
}

type Doc struct {
	Type string `xml:"type,attr"`
	Body string `xml:",innerxml"`
}

type Chassis struct {
	Name      string `xml:"name,attr"`
	Implement string `xml:"implement,attr"`
}

type Assert struct {
	Check  string `xml:"check,attr"`
	Value  string `xml:"value,attr"`
	Method string `xml:"method,attr"`
}

type Field struct {
	Name     string   `xml:"name,attr"`
	Domain   string   `xml:"domain,attr"`
	Type     string   `xml:"type,attr"`
	Label    string   `xml:"label,attr"`
	Reserved bool     `xml:"reserved,attr"`
	Docs     []Doc    `xml:"doc"`
	Asserts  []Assert `xml:"assert"`
}

type Method struct {
	Name        string    `xml:"name,attr"`
	Response    string    `xml:"response>name,attr"`
	Synchronous bool      `xml:"synchronous,attr"`
	Content     bool      `xml:"content,attr"`
	Index       string    `xml:"index,attr"`
	Label       string    `xml:"label,attr"`
	Docs        []Doc     `xml:"doc"`
	Rules       []Rule    `xml:"rule"`
	Fields      []Field   `xml:"field"`
	Chassis     []Chassis `xml:"chassis"`
}

type Class struct {
	Name    string    `xml:"name,attr"`
	Handler string    `xml:"handler,attr"`
	Index   string    `xml:"index,attr"`
	Label   string    `xml:"label,attr"`
	Docs    []Doc     `xml:"doc"`
	Methods []Method  `xml:"method"`
	Chassis []Chassis `xml:"chassis"`
}

type Domain struct {
	Name  string `xml:"name,attr"`
	Type  string `xml:"type,attr"`
	Label string `xml:"label,attr"`
	Rules []Rule `xml:"rule"`
	Docs  []Doc  `xml:"doc"`
}

type Constant struct {
	Name  string `xml:"name,attr"`
	Value int    `xml:"value,attr"`
	Doc   []Doc  `xml:"doc"`
}

type Amqp struct {
	Major   int    `xml:"major,attr"`
	Minor   int    `xml:"minor,attr"`
	Port    int    `xml:"port,attr"`
	Comment string `xml:"comment,attr"`

	Constants []Constant `xml:"constant"`
	Domains   []Domain   `xml:"domain"`
	Classes   []Class    `xml:"class"`
}

type renderer struct {
	Root       Amqp
	bitcounter int
}

var (
	helpers = template.FuncMap{
		"camel": camel,
		"clean": clean,
	}
	packageTemplate = template.Must(template.New("package").Funcs(helpers).Parse(`
	/* GENERATED FILE - DO NOT EDIT */
	/* Rebuild from the protocol/gen.go tool */

	{{with .Root}}
	package wire

	import (
		"fmt"
		"io"
	)

	const (
	{{range .Constants}}
	{{range .Doc}}
	/* {{.Body | clean}} */
	{{end}}{{.Name | camel}} = {{.Value}} {{end}}
	)

	{{range .Classes}}
		{{$class := .}}
		{{range .Methods}}
			{{$method := .}}
			{{if .Docs}}/* {{range .Docs}} {{.Body | clean}} {{end}} */{{end}}
			type {{camel $class.Name $method.Name}} struct {
				{{range .Fields}} {{if .Reserved }} {{else}}
				{{.Name | camel}} {{$.FieldType . | $.NativeType}} {{if .Label}}// {{.Label}}{{end}}{{end}}{{end}}
			}

			func (me {{camel $class.Name $method.Name}}) WriteTo(w io.Writer) (int64, error) {
				var buf buffer
				buf.PutShort({{$class.Index}})
				buf.PutShort({{$method.Index}})
				{{range .Fields}}{{$.FieldEncode .}}
				{{end}}{{$.FinishEncode}}
				fmt.Println("encode: {{$class.Name}}.{{$method.Name}} {{$class.Index}} {{$method.Index}}", buf.Bytes())
				wn, err := w.Write(buf.Bytes())
				return int64(wn), err
			}

			func (me {{camel $class.Name $method.Name}}) HasContent() bool {
				return {{.Content}}
			}

			func (me {{camel $class.Name $method.Name}}) IsSynchronous() bool {
				return {{.Synchronous}}
			}

			func (me {{camel $class.Name $method.Name}}) Class() uint16 {
				return {{$class.Index}}
			}
		{{end}}
	{{end}}

	func (me *buffer) NextMethod() (value Method) {
		class := me.NextShort()
		method := me.NextShort()

		switch class {
		{{range .Classes}}
		{{$class := .}}
		case {{.Index}}: // {{.Name}}
			switch method {
			{{range .Methods}}
			case {{.Index}}: // {{$class.Name}} {{.Name}}
				fmt.Println("NextMethod: class:{{$class.Index}} method:{{.Index}}")
				message := {{camel $class.Name .Name}}{}
				{{range .Fields}}{{$.FieldDecode "message" .}}
				{{end}}{{$.FinishDecode}}
				return message
			{{end}}
			}
		{{end}}
		}
				
		return nil
	}

	{{end}}
	`))
)

/*

	{{range .Classes}}
		{{$class := .}}
		{{range .Methods}}
			{{$method := .}}
			func (me *{{camel $class.Name $method.Name}}) Bytes() []byte {
				var buf buffer
				buf.PutShort({{$class.Index}})
				buf.PutShort({{$method.Index}})
				{{range .Fields}}{{$.FieldEncode .}}
				{{end}}
				log.Println("encode: {{$class.Index}} {{$method.Index}}", buf.Bytes())
				return buf.Bytes()
			}

		{{end}}
	{{end}}
*/

func (me *renderer) BitIncrement(field Field, bits *int) {
	t, _ := me.FieldType(field)
	if t == "bit" {
		*bits = *bits + 1
	}
}

func (me *renderer) FinishEncode() (str string, err error) {
	me.bitcounter = 0
	return
}

func (me *renderer) FieldEncode(field Field) (str string, err error) {
	var fieldType, nativeType, fieldName string

	if fieldType, err = me.FieldType(field); err != nil {
		return "", err
	}

	if nativeType, err = me.NativeType(fieldType); err != nil {
		return "", err
	}

	if field.Reserved {
		fieldName = camel(field.Name)
		str += fmt.Sprintf("var %s %s\n", fieldName, nativeType)
	} else {
		fieldName = fmt.Sprintf("me.%s", camel(field.Name))
	}

	if fieldType == "bit" {
		if me.bitcounter == 0 {
			str += fmt.Sprintf("buf.PutOctet(0)\n")
		}
		str += fmt.Sprintf("buf.Put%s(%s, %d)", camel(fieldType), fieldName, me.bitcounter)
		me.bitcounter = me.bitcounter + 1
		return
	}

	me.bitcounter = 0
	str += fmt.Sprintf("buf.Put%s(%s)", camel(fieldType), fieldName)

	return
}

func (me *renderer) FinishDecode() (string, error) {
	if me.bitcounter > 0 {
		me.bitcounter = 0
		// The last field in the fieldset was a bit field
		// which means we need to consume this word.  This would
		// be better done with object scoping
		return "me.NextOctet()", nil
	}
	return "", nil
}

func (me *renderer) FieldDecode(name string, field Field) (string, error) {
	var str string

	t, err := me.FieldType(field)
	if err != nil {
		return "", err
	}

	if field.Reserved {
		str = "_ = "
	} else {
		str = fmt.Sprintf("%s.%s = ", name, camel(field.Name))
	}

	if t == "bit" {
		str += fmt.Sprintf("me.Next%s(%d)", camel(t), me.bitcounter)
		me.bitcounter = me.bitcounter + 1
		return str, nil
	}

	if me.bitcounter > 0 {
		// We've advanced past a bit word, so consume it before the real decoding
		str = "me.NextOctet() // reset\n" + str
		me.bitcounter = 0
	}

	return str + fmt.Sprintf("me.Next%s()", camel(t)), nil

}

func (me *renderer) Domain(field Field) (domain Domain, err error) {
	for _, domain = range me.Root.Domains {
		if field.Domain == domain.Name {
			return
		}
	}
	return domain, nil
	//return domain, ErrUnknownDomain
}

func (me *renderer) FieldType(field Field) (t string, err error) {
	t = field.Type

	if t == "" {
		var domain Domain
		domain, err = me.Domain(field)
		if err != nil {
			return "", err
		}
		t = domain.Type
	}

	return
}

func (me *renderer) NativeType(amqpType string) (t string, err error) {
	if t, ok := amqpTypeToNative[amqpType]; ok {
		return t, nil
	}
	return "", ErrUnknownType
}

func (me *renderer) Tag(d Domain) string {
	label := "`"

	label += `domain:"` + d.Name + `"`

	if len(d.Type) > 0 {
		label += `,type:"` + d.Type + `"`
	}

	label += "`"

	return label
}

func clean(body string) (res string) {
	return strings.Replace(body, "\r", "", -1)
}

func camel(parts ...string) (res string) {
	for _, in := range parts {
		delim := regexp.MustCompile(`^\w|[-_]\w`)

		res += delim.ReplaceAllStringFunc(in, func(match string) string {
			switch len(match) {
			case 1:
				return strings.ToUpper(match)
			case 2:
				return strings.ToUpper(match[1:])
			}
			panic("unreachable")
		})
	}

	return
}

func main() {
	var r renderer

	spec, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln("Please pass spec on stdin", err)
	}

	err = xml.Unmarshal(spec, &r.Root)

	if err != nil {
		log.Fatalln("Could not parse XML:", err)
	}

	if err = packageTemplate.Execute(os.Stdout, &r); err != nil {
		log.Fatalln("Generate error: ", err)
	}
}

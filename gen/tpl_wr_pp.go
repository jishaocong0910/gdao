package gen

const tplWrPp = `{{if eq .DbType 1}}.Write("?"){{end}}{{if eq .DbType 2}}.Write(b.Pp(":")){{end}}{{if eq .DbType 3}}.Write(b.Pp("$")){{end}}{{if eq .DbType 4}}.Write(b.Pp(":")){{end}}{{if eq .DbType 5}}.Write("?"){{end}}`

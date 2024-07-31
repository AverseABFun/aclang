package preprocessor

import (
	"embed"
	"regexp"
	"strings"
)

var preReg = regexp.MustCompile(`&(use)\((.*?)\)`)

//go:embed standard_library/*.acl
var standard_library embed.FS

func Run(str string) string {
	for _, matches := range preReg.FindAllStringSubmatch(str, -1) {
		var toReplace = matches[0]
		var keyword = matches[1]
		var arguments = strings.Split(strings.TrimSpace(matches[2]), ",")
		if keyword == "use" {
			var location = "standard_library/" + arguments[0] + ".acl"
			var data, err = standard_library.ReadFile(location)
			if err != nil {
				panic("Invalid preprocessor step of &use(" + arguments[0] + ") (got error \"" + err.Error() + "\")")
			}
			var stringData = string(data)
			str = strings.ReplaceAll(str, toReplace, stringData)
		}
	}
	return str
}

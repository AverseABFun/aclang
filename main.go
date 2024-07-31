package main

import (
	"flag"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/averseabfun/aclang/keywords"
	"github.com/averseabfun/aclang/preprocessor"
	"github.com/averseabfun/aclang/syntaxtree"
	"github.com/averseabfun/logger"
)

func main() {
	logger.Log(logger.LogInfo, "ACLang compiler 1.0.0")
	flag.Parse()
	if flag.Arg(0) == "" {
		logger.Log(logger.LogFatal, "No argument passed")
	}
	var stat, err = os.Stat(flag.Arg(0))
	if os.IsNotExist(err) {
		logger.Log(logger.LogFatal, "File does not exist")
	}
	if err != nil {
		logger.Logf(logger.LogFatal, "Encountered error %w when trying to stat the file", err)
	}
	var tempData = make([]byte, stat.Size())
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		logger.Logf(logger.LogFatal, "Encountered error %w when trying to open the file", err)
	}
	_, err = file.Read(tempData)
	if err != nil {
		logger.Logf(logger.LogFatal, "Encountered error %w when trying to read from the file", err)
	}
	var data = string(tempData)
	data = preprocessor.Run(data)
	var dataArr = strings.Split(data, "\n")
	var out = make([]string, 0)
	for _, val := range dataArr {
		var commentReg = regexp.MustCompile(`(?ms)(//.*?$)|(/\*.*?\*/)`)
		val = commentReg.ReplaceAllString(val, "")
		val = strings.TrimSpace(val)
		if val != "" {
			out = append(out, val)
		}
	}
	var tree = syntaxtree.CreateTree()
	dataArr = strings.Split(strings.Join(out, ""), ";")
	var depth = uint(0)
	var nodeToAddTo = tree.RootNode
	var length = len(dataArr)
	for f := 0; f < length; f++ {
		var val = dataArr[f]
		var newNode = tree.CreateNode()
		newNode.Data.Depth = depth
		newNode.Data.Type = syntaxtree.TYPE_VALUE
		var keyword = ""
		var arguments = []string{}
		for i, val := range strings.SplitN(val, ":", 2) {
			if i == 0 {
				keyword = val
				if slices.Index(keywords.Keywords, keyword) != -1 {
					newNode.Data.Type = syntaxtree.TYPE_COMMAND
				}
				continue
			}
			for _, val := range strings.SplitN(val, " ", 2) {
				if !strings.HasPrefix(val, "{") {
					arguments = append(arguments, val)
					continue
				}
				val = strings.Replace(val, "{", "", 1)
				dataArr = slices.Insert(dataArr, f+1, val)
				depth++
				if newNode.Data.Type == syntaxtree.TYPE_COMMAND {
					newNode.Data.Type = syntaxtree.TYPE_GROUPING_COMMAND
				} else {
					newNode.Data.Type = syntaxtree.TYPE_GROUPING
				}
			}
		}
		if strings.Contains(keyword, "}") {
			if depth > 0 {
				depth--
			}
			nodeToAddTo = *nodeToAddTo.Parent
			continue
		}
		newNode.Data.Keyword = keyword
		newNode.Data.Arguments = arguments
		nodeToAddTo.AddChild(newNode)
		if newNode.Data.Type == syntaxtree.TYPE_GROUPING || newNode.Data.Type == syntaxtree.TYPE_GROUPING_COMMAND {
			nodeToAddTo = newNode
		}
	}
	logger.Log(logger.LogDebug, tree.RootNode.String())
}

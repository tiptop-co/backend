//go:build tools
// +build tools

//go:generate go build -mod=mod -o ../bin/testgen github.com/LaHainee/go_test_template_gen/cmd

package tools

import _ "github.com/LaHainee/go_test_template_gen/cmd"

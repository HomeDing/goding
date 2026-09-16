// Package main provides the command line parsing and command execution for the GoDing application.
package assets

import (
	_ "embed"
	_ "image/png"
)

//go:embed goding.png
var GodingIcon []byte

//go:embed goding.png
var GodingIconDark []byte

// // g o : e m b e d  templates/*.html
// var emFS embed.FS

// func GetIcon(name string) []byte {
// 	icon, err := emFS.ReadFile("./goding.png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	return icon
// } // GetIcon()

// End.

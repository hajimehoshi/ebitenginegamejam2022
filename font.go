// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"embed"
	"path"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed resource
var resource embed.FS

var robotoBold font.Face

func init() {
	bs, err := resource.ReadFile(path.Join("resource", "Roboto-Bold.ttf"))
	if err != nil {
		panic(err)
	}

	font, err := opentype.Parse(bs)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size: 144,
		DPI:  72,
	})
	if err != nil {
		panic(err)
	}
	robotoBold = face
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"embed"
	"io"
	"io/fs"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed resource
var resourceRootFS embed.FS

var resourceFS fs.FS

func init() {
	f, err := fs.Sub(resourceRootFS, "resource")
	if err != nil {
		panic(err)
	}
	resourceFS = f
}

var (
	spaceAgeBig   font.Face
	spaceAgeSmall font.Face
)

func init() {
	f, err := resourceFS.Open("spaceage.otf")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	font, err := opentype.Parse(bs)
	if err != nil {
		panic(err)
	}

	{
		face, err := opentype.NewFace(font, &opentype.FaceOptions{
			Size: 144,
			DPI:  72,
		})
		if err != nil {
			panic(err)
		}
		spaceAgeBig = face
	}
	{
		face, err := opentype.NewFace(font, &opentype.FaceOptions{
			Size: 108,
			DPI:  72,
		})
		if err != nil {
			panic(err)
		}
		spaceAgeSmall = face
	}
}

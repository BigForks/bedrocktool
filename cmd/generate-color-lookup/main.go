package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/bedrock-tool/bedrocktool/utils"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"golang.org/x/exp/maps"
)

func main() {
	if len(os.Args) != 2 {
		println("usage: generate-color-lookup <resource_packs_folder>")
	}
	folder := os.Args[1]
	packNames, err := os.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	var packs []resource.Pack
	for _, fi := range packNames {
		name := fi.Name()
		pack, err := utils.PackFromBase(resource.MustReadPath(folder + "/" + name))
		if err != nil {
			log.Fatal(err)
		}
		packs = append(packs, pack)
	}

	colors := utils.ResolveColors(nil, packs)
	keys := maps.Keys(colors)
	sort.Strings(keys)

	f, err := os.Create("blockcolors.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.WriteString("package utils\n\n")
	f.WriteString("import (\n\t\"image/color\"\n)\n\n")

	f.WriteString("func LookupColor(name string) color.RGBA {\n")
	f.WriteString("\tswitch name {\n")
	for _, name := range keys {
		color := colors[name]
		f.WriteString("\tcase \"" + name + "\":\n")
		f.WriteString(fmt.Sprintf("\t\treturn color.RGBA{0x%x, 0x%x, 0x%x, 0x%x}\n", color.R, color.G, color.B, color.A))
	}
	f.WriteString("\tdefault:\n\t\treturn color.RGBA{0xff, 0x00, 0xff, 0x00}\n")
	f.WriteString("\t}\n")
	f.WriteString("}\n")
}

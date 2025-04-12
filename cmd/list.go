package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/packwiz/packwiz/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the mods in the modpack",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		// Load pack
		pack, err := core.LoadPack()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Load index
		index, err := pack.LoadIndex()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Load mods
		mods, err := index.LoadAllMods()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Filter mods by side
		if viper.IsSet("list.side") {
			side := viper.GetString("list.side")
			if side != core.UniversalSide && side != core.ServerSide && side != core.ClientSide {
				fmt.Printf("Invalid side %q, must be one of client, server, or both (default)\n", side)
				os.Exit(1)
			}

			i := 0
			for _, mod := range mods {
				if mod.Side == side || mod.Side == core.EmptySide || mod.Side == core.UniversalSide || side == core.UniversalSide {
					mods[i] = mod
					i++
				}
			}
			mods = mods[:i]
		}

		sort.Slice(mods, func(i, j int) bool {
			return strings.ToLower(mods[i].Name) < strings.ToLower(mods[j].Name)
		})



		// Print mods
        if viper.GetBool("list.json") {

            type modInfo struct {
                Name string `json:"name"`
                FileName string `json:"fileName"`
                Source string `json:"source"`
                Slug string `json:"slug"`
                DownloadUrl string `json:"downloadUrl"`
            }

            modInfos := make([]modInfo, 0, len(mods))


            for _, mod := range mods {
                _, metaPath := path.Split(mod.GetFilePath())
                updateKeys := make([]string, 0, len(mod.Update))

                for k := range mod.Update {
                    updateKeys = append(updateKeys, k)
                }

                source := strings.Join(updateKeys, ", ")

                info := modInfo{
                    Name: mod.Name,
                    FileName: mod.FileName,
                    Slug: strings.ReplaceAll(metaPath, ".pw.toml", ""),
                    Source: source,
                    DownloadUrl: mod.Download.URL,
                }

                modInfos = append(modInfos, info)
            }

            output, err := json.MarshalIndent(modInfos, "", strings.Repeat(" ", 4))
            if err != nil {
                fmt.Println("Error outputting json")
                return
            }

            fmt.Println(string(output))

        } else if viper.GetBool("list.version") {
			for _, mod := range mods {
				fmt.Printf("%s (%s)\n", mod.Name, mod.FileName)
			}
		} else {
			for _, mod := range mods {
				fmt.Println(mod.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("version", "v", false, "Print version")
	_ = viper.BindPFlag("list.version", listCmd.Flags().Lookup("version"))
	listCmd.Flags().StringP("side", "s", "", "Filter mods by side (e.g., client or server)")
	_ = viper.BindPFlag("list.side", listCmd.Flags().Lookup("side"))
    listCmd.Flags().BoolP("json", "j", false, "Print mod information as json")
    _ = viper.BindPFlag("list.json", listCmd.Flags().Lookup("json"))

}

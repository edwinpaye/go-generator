package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect and visualize tables, enums, partials, and bounded context groups in DBML schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := schemaPath
		if path == "" {
			path = "./schema.dbml"
		}

		doc, _, err := generatorService.InspectDBML(context.Background(), path)
		if err != nil {
			return err
		}

		fmt.Println("================================================================")
		if doc.Project != nil {
			fmt.Printf("📦 Project: %s\n", doc.Project.Name)
			if doc.Project.Note != "" {
				fmt.Printf("   Note: %s\n", doc.Project.Note)
			}
		} else {
			fmt.Println("📦 Project: (Unnamed)")
		}
		fmt.Println("================================================================")

		fmt.Printf("\n📊 Summary:\n")
		fmt.Printf("  • Tables:       %d\n", len(doc.Tables))
		fmt.Printf("  • Partials:     %d\n", len(doc.TablePartials))
		fmt.Printf("  • Enums:        %d\n", len(doc.Enums))
		fmt.Printf("  • TableGroups:  %d\n", len(doc.TableGroups))
		fmt.Printf("  • References:   %d\n", len(doc.Refs))

		if len(doc.Enums) > 0 {
			fmt.Println("\n🏷️  Enums:")
			for _, e := range doc.Enums {
				var valNames []string
				for _, v := range e.Values {
					valNames = append(valNames, v.Name)
				}
				fmt.Printf("  - %s (%s)\n", e.FullName(), strings.Join(valNames, " | "))
			}
		}

		if len(doc.TablePartials) > 0 {
			fmt.Println("\n🧩 TablePartials (Mixins):")
			for _, p := range doc.TablePartials {
				fmt.Printf("  - %s (%d columns)\n", p.Name, len(p.Columns))
			}
		}

		if len(doc.TableGroups) > 0 {
			fmt.Println("\n📁 TableGroups (Bounded Contexts):")
			for _, tg := range doc.TableGroups {
				var items []string
				for _, it := range tg.Items {
					items = append(items, it.FullName())
				}
				fmt.Printf("  - %s: [%s]\n", tg.Name, strings.Join(items, ", "))
			}
		}

		fmt.Println("\n📋 Tables & Columns:")
		for _, t := range doc.Tables {
			fmt.Printf("  • Table: %s\n", t.FullName())
			for _, c := range t.Columns {
				flags := ""
				if c.PK {
					flags += " [PK]"
				}
				if c.Unique {
					flags += " [UNIQUE]"
				}
				if !c.Nullable {
					flags += " [NOT NULL]"
				}
				fmt.Printf("      - %-20s %-15s%s\n", c.Name, c.Type, flags)
			}
		}

		fmt.Println("\n================================================================")
		return nil
	},
}

func init() {
	inspectCmd.Flags().StringVarP(&schemaPath, "schema", "s", "./schema.dbml", "Path to source DBML schema file")
}

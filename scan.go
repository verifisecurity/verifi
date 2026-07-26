package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/verifisecurity/verifi/internal/finding"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/osv"
	"github.com/verifisecurity/verifi/internal/plan"
	"github.com/verifisecurity/verifi/internal/reason"
)

// runScan implements `verifi scan <path>`: the one read-only command. It
// resolves the project's dependency tree, matches it against the OSV database,
// and reports what needs fixing with a reasoned fix per package. npm for now.
//
// The tree views (--inventory, --sbom) need no advisory database and never
// touch it. --json is a modifier on whichever view is selected, not a view of
// its own. By default the scan checks the registry so it only recommends
// versions that exist; --offline skips that and trusts OSV's fixed versions.
func runScan(args []string) error {
	path, dbDir, outPath := "", "", ""
	asJSON, asSBOM, asInventory, offline, download := false, false, false, false, false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--json":
			asJSON = true
		case "--sbom":
			asSBOM = true
		case "--inventory":
			asInventory = true
		case "--offline":
			offline = true
		case "--download":
			download = true
		case "--db":
			if i+1 >= len(args) {
				return fmt.Errorf("--db needs a directory")
			}
			i++
			dbDir = args[i]
		case "--out":
			if i+1 >= len(args) {
				return fmt.Errorf("--out needs a file path")
			}
			i++
			outPath = args[i]
		default:
			if len(a) > 0 && a[0] == '-' {
				return fmt.Errorf("unknown flag %q", a)
			}
			path = a
		}
	}
	if path == "" {
		path = "."
	}
	if download && dbDir != "" {
		return fmt.Errorf("--download fetches into the local cache, so it cannot be combined with --db")
	}

	// The tree views describe what you depend on, not what is wrong with it, so
	// they resolve the lockfile and stop. No advisory database required.
	if asSBOM || asInventory {
		inv, err := loadInventory(path)
		if err != nil {
			return err
		}
		switch {
		case asSBOM:
			out, err := inv.ToCycloneDX()
			if err != nil {
				return err
			}
			fmt.Println(string(out))
		case asJSON:
			out, err := json.MarshalIndent(inv, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
		default:
			printInventory(inv)
		}
		return nil
	}

	if download {
		inv, err := loadInventory(path)
		if err != nil {
			return err
		}
		if err := downloadDB(inv.Ecosystem); err != nil {
			return err
		}
	}

	res, err := analyze(path, dbDir, offline)
	if err != nil {
		return err
	}

	// Every scan writes the plan, so `verifi fix` always has something current to
	// read whichever view was asked for. It goes under the verifi home, never
	// into the scanned project, so scanning cannot dirty someone's working tree.
	p := buildPlan(res, path, version)
	p.Generated = now()
	planPath := outPath
	if planPath == "" {
		planPath = plan.Path(verifiHome(), path)
	}
	if err := plan.Write(planPath, p); err != nil {
		return fmt.Errorf("write plan: %w", err)
	}

	if asJSON {
		out, err := json.MarshalIndent(res.findings, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}

	printFindings(res.inv, res.findings, res.recs, res.imported, res.dbMeta)
	printPlanHint(planPath, p)
	return nil
}

// printPlanHint tells the user where the plan went and what to do with it. It is
// the handover from scan to the human: nothing is applied until someone opens
// this file and marks something.
func printPlanHint(path string, p plan.Plan) {
	actionable := 0
	for _, c := range p.Candidates {
		if c.Actionable() {
			actionable++
		}
	}
	if actionable == 0 {
		return
	}
	fmt.Printf("Plan written to %s\n", path)
	fmt.Printf("Mark the fixes you want with \"apply\": true, then run: verifi fix %s\n", p.Workspace)
}

func printFindings(inv *inventory.Inventory, findings []finding.Finding, recs []reason.Recommendation, imported map[string]bool, dbMeta *osv.Meta) {
	fmt.Printf("%s@%s (%s)\n", inv.Root.Name, inv.Root.Version, inv.Ecosystem)
	if line := freshnessLine(dbMeta); line != "" {
		fmt.Println(line)
	}
	if len(findings) == 0 {
		fmt.Printf("%d packages scanned, none vulnerable.\n", len(inv.Packages))
		return
	}

	byPkg := map[string][]finding.Finding{}
	var order []string
	for _, f := range findings {
		if _, ok := byPkg[f.Name]; !ok {
			order = append(order, f.Name)
		}
		byPkg[f.Name] = append(byPkg[f.Name], f)
	}
	sort.Slice(order, func(i, j int) bool {
		ri, rj := worstRank(byPkg[order[i]]), worstRank(byPkg[order[j]])
		if ri != rj {
			return ri < rj
		}
		return order[i] < order[j]
	})

	recByName := map[string]reason.Recommendation{}
	for _, r := range recs {
		recByName[r.Name] = r
	}

	fmt.Printf("%d packages scanned, %d vulnerable.\n\n", len(inv.Packages), len(order))
	for _, name := range order {
		fs := byPkg[name]
		fmt.Printf("%-8s %s %s   %s\n", sevLabel(worstRank(fs)), name, fs[0].Version, directTag(inv, name))
		fmt.Printf("   used: %s\n", usageLine(inv, name, imported))
		for _, f := range fs {
			id := f.Advisory
			if len(f.Aliases) > 0 {
				id = f.Aliases[0] + " (" + f.Advisory + ")"
			}
			fmt.Printf("   %s   %s\n", id, f.Summary)
		}
		if r, ok := recByName[name]; ok {
			switch r.Action {
			case "upgrade":
				fmt.Printf("   Fix: upgrade to %s   confidence: %s\n", r.Target, r.Confidence)
			case "remove":
				fmt.Printf("   Fix: remove %s   confidence: %s\n", r.Name, r.Confidence)
			default:
				fmt.Printf("   No published fix   confidence: %s\n", r.Confidence)
			}
			fmt.Printf("        %s\n", r.Reason)
			for _, l := range r.Limitations {
				fmt.Printf("        - %s\n", l)
			}
		}
		fmt.Println()
	}
}

func sevRank(s string) int {
	switch strings.ToUpper(s) {
	case "CRITICAL":
		return 0
	case "HIGH":
		return 1
	case "MEDIUM", "MODERATE":
		return 2
	case "LOW":
		return 3
	default:
		return 4
	}
}

func worstRank(fs []finding.Finding) int {
	r := 4
	for _, f := range fs {
		if x := sevRank(f.Severity); x < r {
			r = x
		}
	}
	return r
}

func sevLabel(rank int) string {
	switch rank {
	case 0:
		return "CRITICAL"
	case 1:
		return "HIGH"
	case 2:
		return "MEDIUM"
	case 3:
		return "LOW"
	default:
		return "UNKNOWN"
	}
}

// usageLine states whether the project's own code imports this package. A
// vulnerable direct dependency nothing imports is a strong remove candidate.
// The phrase is bare so it reads correctly both behind the view's "used:" label
// and on its own as the plan's `used` field.
func usageLine(inv *inventory.Inventory, name string, imported map[string]bool) string {
	if directTag(inv, name) != "direct" {
		return "indirectly, pulled in by another dependency"
	}
	if imported[name] {
		return "imported by your code"
	}
	return "not imported by your code, removing it may clear this"
}

func directTag(inv *inventory.Inventory, name string) string {
	for _, p := range inv.Packages {
		if p.Name == name {
			if p.Direct {
				return "direct"
			}
			return "transitive"
		}
	}
	return ""
}

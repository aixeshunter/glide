package cmd

import (
	"os"
	"path"

	"github.com/Masterminds/cookoo"
	"github.com/kylelemons/go-gypsy/yaml"
)

// Recurse does glide installs on dependent packages.
// Recurse looks in all known packages for a glide.yaml files and installs for
// each one it finds.
func Recurse(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	Info("Checking dependencies for updates.\n")
	conf := p.Get("conf", &Config{}).(*Config)
	vend, _ := VendorPath(c)

	if len(conf.Imports) == 0 {
		Info("No imports.\n")
	}

	// Look in each package to see whether it has a glide.yaml, and no vendor/
	for _, imp := range conf.Imports {
		Info("Looking in %s for a glide.yaml file.\n", imp.Name)
		base := path.Join(vend, imp.Name)
		if !needsGlideUp(base) {
			Info("Package %s manages its own dependencies.\n", imp.Name)
		}
		Info("Package %s needs `glide up`\n", imp.Name)
		if err := dependencyGlideUp(base); err != nil {
			Warn("Failed to update dependency %s: %s", imp.Name, err)
		}
	}

	// Run `glide up`
	return nil, nil
}

func dependencyGlideUp(base string) error {
	//conf := new(Config)
	fname := path.Join(base, "glide.yaml")
	f, err := yaml.ReadFile(fname)
	if err != nil {
		return err
	}

	conf, err := FromYaml(f.Root)
	if err != nil {
		return err
	}
	for _, imp := range conf.Imports {
		Info("Importing %s to project %s\n", imp.Name, base)
		// We don't use the global var to find vendor dir name because the
		// user may mis-use that var to modify the local vendor dir, and
		// we don't want that to break the embedded vendor dirs.
		wd := path.Join(base, "vendor", imp.Name)
		if err := ensureDir(wd); err != nil {
			Warn("Skipped getting %s (vendor/ error): %s\n", imp.Name, err)
			continue
		}

		// How do we want to do this? Should we run the glide command,
		// which would allow environmental control, or should we just
		// run the update route in that directory?
		if err := VcsGet(imp, wd); err != nil {
			Warn("Skipped getting %s: %s\n", imp.Name, err)
		}
	}
	return nil
}

func ensureDir(dirpath string) error {
	if fi, err := os.Stat(dirpath); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll(dirpath, 0755)
}

func needsGlideUp(dir string) bool {
	stat, err := os.Stat(path.Join(dir, "glide.yaml"))
	if err != nil || stat.IsDir() {
		return false
	}

	// Should probably see if vendor is there and non-empty.

	return true
}

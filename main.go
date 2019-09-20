package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var config = flag.String("config", "config.yml", "path to config file")

func main() {
	flag.Parse()
	data, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatal(err)
	}
	var conf Config
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal(err)
	}
	if err := run(conf); err != nil {
		log.Fatal(err)
	}
}

func run(conf Config) error {
	if err := os.MkdirAll(conf.OutputDir, 0755); err != nil {
		return err
	}
	for _, plan := range conf.Plans {
		for workspace, flags := range plan.WorkspaceFlags {
			dir := filepath.Join(conf.BaseDir, plan.Dir)
			fmt.Printf("selecting workspace %s in dir %s\n", workspace, dir)
			if err := workspaceSelect(dir, workspace); err != nil {
				return fmt.Errorf("workspace select fail %s %v", workspace, err)
			}
			fmt.Printf("planning workspace %s in dir %s \n\twith flags %v\n", workspace, dir, flags)
			if err := terraformPlan(plan.Dir, dir, plan.Profile, conf.OutputDir, workspace, flags); err != nil {
				return err
			}
		}
	}

	return nil
}

type Config struct {
	BaseDir   string       `toml:"base_dir" yaml:"base_dir"`
	OutputDir string       `toml:"output_dir" yaml:"output_dir"`
	Plans     []PlanConfig `toml:"plan" yaml:"plans"`
}

type PlanConfig struct {
	Dir            string              `toml:"dir" yaml:"dir"`
	Profile        string              `toml:"profile" yaml:"profile"`
	WorkspaceFlags map[string][]string `toml:"workspace_flags" yaml:"workspace_flags"`
}

func terraformPlan(projectName, dir, profile, outputDir, workspace string, flags []string) error {
	args := []string{"plan", "-lock=false", "-no-color"}
	args = append(args, flags...)
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), fmt.Sprintf("AWS_PROFILE=%s", profile))
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			switch int(exitError.ExitCode()) {
			case 1:
				log.Printf("error while planning: terraform %v", args)
			case 2:
				log.Printf("plan succeeded with diff: terraform %v", args)
			case 0:
				// shouldn't happen, right?
				log.Println("shouldn't happen")
			}
		}
		// NOTE: we keep going even if there's an error in the plan
	}
	os.MkdirAll(filepath.Join(outputDir, projectName), 0755)
	planFile := filepath.Join(outputDir, projectName, workspace)
	err = ioutil.WriteFile(planFile, output, 0644)
	if err != nil {
		return err
	}
	return nil
}

func workspaceSelect(dir, workspace string) error {
	args := []string{"workspace", "select", workspace}
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

/*
Terraform plan exit codes:
0 = Succeeded with empty diff (no changes)
1 = Error
2 = Succeeded with non-empty diff (changes present)
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func main() {
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	var conf Config
	err = toml.Unmarshal(data, &conf)
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
	BaseDir   string       `toml:"base_dir"`
	OutputDir string       `toml:"output_dir"`
	Plans     []PlanConfig `toml:"plan"`
}

type PlanConfig struct {
	Dir            string              `toml:"dir"`
	Profile        string              `toml:"profile"`
	WorkspaceFlags map[string][]string `toml:"workspace_flags"`
}

func terraformPlan(projectName, dir, profile, outputDir, workspace string, flags []string) error {
	args := []string{"plan", "-lock=false", "-no-color"}
	args = append(args, flags...)
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), fmt.Sprintf("AWS_PROFILE=%s", profile))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return err
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

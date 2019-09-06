# plan-everything

Run terraform plans.

## Usage

```
./plan-everything -config config.yml
```

Example config

```yml
base_dir: /path/to/terraform/projects
output_dir: /path-where/plans-will-be/written
plans:
- dir: someSubdir
  profile: some_aws_profile
  workspace_flags:
    someworkspace: ["-var-file", "foo.tfvars"]
    another: ["-var-file", "foo2.tfvars"]
- dir: anotherSubdir
  profile: some_aws_profile
  workspace_flags:
    someworkspace: ["-var-file", "foo.tfvars"]
    another: ["-var-file", "foo2.tfvars"]
```



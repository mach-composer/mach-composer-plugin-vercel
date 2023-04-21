# Vercel plugin for MACH composer

This plugin adds an integration for Vercel for use with MACH composer.

This allows you to streamline your configuration and share it as an integration with your MACH components.

## Requirements
- [MACH Composer >=2.5](https://github.com/labd/mach-composer)
- [terraform-provider-vercel](https://github.com/vercel/terraform-provider-vercel)

## Usage

You can set up configuration variables on a global, a site and a component specific level. Inheritance will take place on each level, combining configuration from global
all the way down to a component level.

Example of setting up global and component level configuration:
```yaml
global:
   # ...
   vercel:
    api_token: "token"
    team_id: "team"
    project_config:
        manual_production_deployment: false # Variable to help with setting up manual deployments in Terraform
        serverless_function_region: "fra1"
        environment_variables:
            - key: CUSTOM_GLOBAL_ENVIRONMENT_VARIABLE
              value: custom
              environments: ["production"] # When left empty it will default to ["production", "preview", "development"]
sites:
    - identifier: my-site
      # ...
      vercel:
        project_config:
          environment_variables:
            - key: SITE_SPECIFIC_ENVIRONMENT_VARIABLE
              value: site
      components:
        - name: my-component
          vercel: # Override defaults on component level
            project_config:
                name: "my-vercel-project"
                framework: "nextjs"
                manual_production_deployment: true
                git_repository:
                  type: "github"
                  repo: "mach-composer/my-vercel-project"
                environment_variables:
                    - key: CUSTOM_COMPONENT_SPECIFIC_ENVIRONMENT_VARIABLE
                      value: custom
                      environments: ["preview"]
                domains:
                  - domain: "cool-plugin.com"
                    git_branch: main
                    redirect: "cool-plugin.vercel.app"
                    redirect_status_code: 307
```

You can then set up Vercel as an integration for a specific component:
```yaml
components:
    - name: my-site
      source: git::https://github.com/mach-composer/my-site//terraform
      version: "1234567"
      integrations: [vercel] # This will prepend the config as variables for your terraform config
```

Then you can set up your terraform resources with the given variables:
```hcl
resource "vercel_project" "project" {
  name                       = var.vercel_project_name
  framework                  = var.vercel_project_framework
  team_id                    = var.vercel_team_id
  serverless_function_region = var.vercel_project_serverless_function_region
  environment                = local.environment
  git_repository = var.vercel_project_git_repository

  build_command  = var.vercel_project_build_command
  root_directory = var.vercel_project_root_directory

  lifecycle {
    # never accidentally destroy this resource
    prevent_destroy = true
  }
}
```

### Manual deployment example

This is an example if you want to manually deploy vercel projects on MACH config updates:
```hcl
resource "vercel_deployment" "manual_production_deployment" {
  count      = var.manual_production_deployment ? 1 : 0
  project_id = vercel_project.my_project.id
  team_id    = var.vercel_team_id
  ref        = var.component_version
  production = true
}
```

## Support

This plugin is in an early stage of development and only supports a tiny subset of the Vercel terraform provider.

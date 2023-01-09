# Vercel plugin for MACH composer

This plugin adds an integration for Vercel for use with MACH composer.

This allows you to streamline your configuration and share it as an integration with your MACH components.

## Requirements
- [MACH Composer 3.x](https://github.com/labd/mach-composer)
- [terraform-provider-vercel](https://github.com/vercel/terraform-provider-vercel)

## Usage

You can set up configuration variables on both a global and a component specific level.
All values set on a global level will be inherited and potentially overwritten by its components.

Example of setting up global and component level configuration:
```yaml
global:
   # ...
   vercel:
    api_token: "token"
    team_id: "team"
    project_config:
        manual_production_deployment: false # Variable to help with setting up manual deployments in Terraform
        environment_variables:
            - key: CUSTOM_GLOBAL_ENVIRONMENT_VARIABLE
              value: custom
              environments: ["production"] # When left empty it will default to ["production", "preview", "development"]
sites:
    - identifier: my-site
      # ...
      components:
        - name: my-component
          vercel: # Override defaults on component level
            project_config:
                manual_production_deployment: true
                environment_variables:
                    - key: CUSTOM_SITE_SPECIFIC_ENVIRONMENT_VARIABLE
                      value: custom
                      environments: ["preview"]
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

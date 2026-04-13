resource "komodo_procedure" "example" {
  name = "deploy-pipeline"

  # Stage 1: Deploy the stack
  stage {
    name = "Deploy"

    execution {
      type = "DeployStack"
      parameters = {
        id        = komodo_stack.app.id
        stop_time = "60"
      }
    }
  }

  # Stage 2: Run health checks in parallel
  stage {
    name = "Health Checks"

    execution {
      type = "RunAction"
      parameters = {
        id = komodo_action.health_check.id
      }
    }

    execution {
      type    = "RunBuild"
      enabled = false
      parameters = {
        id = komodo_build.app.id
      }
    }
  }

  # Stage 3: Notify via another procedure
  stage {
    name = "Notify"

    execution {
      type = "RunProcedure"
      parameters = {
        id = komodo_procedure.send_notification.id
      }
    }
  }

  schedule {
    format     = "Cron"
    expression = "0 2 * * *"
    enabled    = true
    timezone   = "America/New_York"
  }

  failure_alert_enabled = true
}

resource "komodo_procedure" "example" {
  name = "my-procedure"

  stages = jsonencode([
    {
      name     = "Deploy"
      parallel = false
      executions = [
        {
          execution = {
            type = "DeployStack"
            params = {
              stack = "my-stack"
            }
          }
          enabled = true
        }
      ]
    }
  ])
}

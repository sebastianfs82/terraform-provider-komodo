# Import a Server terminal: target_id:name
terraform import komodo_terminal.server 69db2f6e0816ddac8244a5b3:my-terminal

# Import a Container/Stack/Deployment terminal: target_type:target_id:name
terraform import komodo_terminal.container Container:69db2f6e0816ddac8244a5b3:my-container-terminal
terraform import komodo_terminal.stack Stack:69dba72e0816ddac8244ae18:my-stack-terminal
terraform import komodo_terminal.deployment Deployment:69dba72e0816ddac8244ae18:my-deployment-terminal

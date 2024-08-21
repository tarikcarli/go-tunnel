# Go-tunnel

## How to run go-tunnel
### Example
-   go-tunnel --mode tunnel --host 0.0.0.0:8081 --source 127.0.0.1:8080
-   go-tunnel --mode server --secret bb68e0f3-cd4f-4508-960b-3cfa172796e6 --host 0.0.0.0:5050 --min-idle-connections 10
-   go-tunnel --mode client --secret bb68e0f3-cd4f-4508-960b-3cfa172796e6 --server 186.154.25.1:5050 --target 0.0.0.0:443 --source 127.0.0.1:8080
### Options
- --mode (enum:[tunnel,client,server])
- --host (only valid in server,tunnel)
- --target (only valid in server,tunnel)
- --secret (Shared symmetric encryption key, only valid in server,client)
- --min-idle-connections (only valid in server)
- --server (only valid in client)
- --source (only valid in client)

## Design

![Go-tunnel architecture](https://raw.githubusercontent.com/tarikcarli/go-tunnel/main/design.jpg)

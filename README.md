# Chat TCP en Go

Un sistema cliente-servidor concurrente construido en Go utilizando Protocolo head-body sobre sockets TCP.

Arquitectura 1
→ event loop centralizado

Problema observado
→ archivos grandes bloquean distribución

Arquitectura 2
→ Worker por cliente

Problema restante
→ chat y archivos comparten stream TCP

Arquitectura 3
→ Chat Server + File Server

## Cómo ejecutar

1. Iniciar el servidor:
   ```bash
   go run main.go
   ```

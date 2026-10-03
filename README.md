# Chat TCP en Go

Sistema cliente-servidor concurrente en Go que usa un protocolo head-body sobre sockets TCP.

## Estructura

```text
cmd/
  chat-server/   # Punto de entrada del servidor de chat
  client/        # Punto de entrada del cliente
  file-server/   # Reservado; todavía no implementado
internal/
  chat/          # Cliente y servidor de chat
  files/         # Funcionalidad de archivos (en desarrollo)
  protocol/      # Lectura y construcción de headers
storage/
  uploads/       # Archivos recibidos
```

Los ejecutables viven en `cmd/` y la lógica reutilizable, que no se expone
como API pública del módulo, en `internal/`.

## Cómo ejecutar

Inicia el servidor de chat en una terminal:

```bash
go run ./cmd/chat-server
```

Inicia el cliente en otra terminal:

```bash
go run ./cmd/client
```

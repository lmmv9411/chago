# Chat TCP en Go

Sistema cliente-servidor concurrente en Go que usa sockets TCP para intercambio de mensajes y transferencia de archivos.

## Características

- Chat en tiempo real entre clientes conectados al servidor de chat.
- Transferencia de archivos por un servicio independiente en TCP (puerto 8081).
- Notificación automática en el chat cuando se comparte un archivo.
- Descarga de archivos al directorio `downloads/` desde el cliente.
- Almacenamiento de archivos recibidos en `uploads/` por el servidor de archivos.

## Estructura del proyecto

```text
cmd/
  chat-server/   # Inicia el servidor de chat y el servidor de archivos
  client/        # Punto de entrada del cliente
  file-server/   # Servidor de archivos independiente
internal/
  client/        # Lógica del cliente, lectura del teclado y manejo de mensajes
  serverchat/    # Servidor del chat
  serverfiles/   # Servidor de archivos
  protocolchat/  # Construcción y lectura de headers del protocolo de chat
  protocolfile/  # Construcción y lectura de headers del protocolo de archivos
```

Los ejecutables viven en `cmd/` y la lógica reutilizable, que no se expone como API pública del módulo, en `internal/`.

## Requisitos

- Go instalado y configurado en el sistema.
- Acceso a la red local para que los clientes se conecten al mismo puerto.

## Cómo ejecutar

1. Inicia el servidor principal (chat + archivos):

```bash
go run ./cmd/chat-server
```

Esto levanta:
- servidor de chat en el puerto `8080`
- servidor de archivos en el puerto `8081`

2. En otra terminal, inicia el cliente:

```bash
go run ./cmd/client
```

3. Si quieres arrancar solo el servidor de archivos de forma independiente:

```bash
go run ./cmd/file-server
```

## Uso del cliente

Al ejecutar el cliente se te pedirá:

- una dirección IP del servidor
- un nombre de usuario

Para usar la IP por defecto, escribe `y` cuando te pregunte la dirección. El valor por defecto configurado en el cliente es `192.168.1.33`.

### Enviar mensajes

Escribe directamente el texto y se enviará al servidor del chat.

### Enviar un archivo

Escribe:

```text
/file
```

Luego se te pedirá la ruta del archivo local que deseas compartir. El cliente lo envía al servidor de archivos y luego notifica al chat con el nombre del archivo.

### Descargar un archivo recibido

Cuando otro usuario comparte un archivo, el cliente recibe una notificación del tipo `file/notification` y descarga automáticamente el archivo en el directorio local `downloads/`.

## Directorios de salida

- Archivos subidos al servidor: `uploads/`
- Archivos descargados por el cliente: `downloads/`

> Los directorios se crean al momento de ejecutar los servicios si no existen.

## Nota

Este proyecto está orientado a un entorno de laboratorio o prueba local, por lo que la configuración de la IP y los puertos se asume dentro de la misma red local.

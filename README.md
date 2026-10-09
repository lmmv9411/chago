# Chago

Aplicación de chat cliente-servidor escrita en Go. Los clientes intercambian mensajes mediante TCP y pueden compartir archivos a través de un servicio TCP independiente.

## Características

- Chat en tiempo real entre los clientes conectados.
- Servidor de chat concurrente que retransmite mensajes a los demás clientes.
- Envío y descarga de archivos mediante un servicio de archivos separado.
- Notificaciones en el chat cuando un cliente comparte un archivo.
- Indicadores de progreso para las transferencias activas, conservando el orden en que comenzaron.
- Cliente de terminal interactivo con Bubble Tea.

## Arquitectura del proyecto

```text
cmd/
  chat-server/   # Inicia conjuntamente los servidores de chat y de archivos
  client/        # Inicia el cliente de terminal
  file-server/   # Inicia únicamente el servidor de archivos
internal/
  client/        # Conexiones, envío de mensajes y transferencias de archivos
  protocolchat/  # Lectura y construcción de headers del protocolo de chat
  protocolfile/  # Lectura y construcción de headers del protocolo de archivos
  serverchat/    # Aceptación de clientes y retransmisión de mensajes
  serverfiles/   # Recepción y entrega de archivos
  terminal/      # Interfaz de terminal y progreso de transferencias
```

El código de aplicación está en `internal/`; los programas ejecutables están en `cmd/`.

## Requisitos

- Go instalado. La versión del módulo está especificada en `go.mod`.
- Acceso de red entre los clientes y el equipo donde se ejecutan los servidores.
- Puertos TCP `8080` y `8081` disponibles para los servicios.

## Ejecución

### Iniciar el sistema completo

En una terminal, desde la raíz del repositorio:

```bash
go run ./cmd/chat-server
```

Este comando inicia ambos servicios:

- Chat en TCP `:8080`.
- Archivos en TCP `:8081`.

El proceso se mantiene activo mientras ambos servidores estén ejecutándose.

### Iniciar un cliente

En otra terminal, ejecuta:

```bash
go run ./cmd/client
```

Inicia el comando en cada terminal o equipo que quieras conectar al chat.

### Iniciar solo el servidor de archivos

También se puede ejecutar de forma independiente:

```bash
go run ./cmd/file-server
```

Este comando inicia únicamente el servicio TCP de archivos en `:8081`. Para usar el cliente completo también debe estar disponible el servidor de chat en `:8080`.

## Uso

Al iniciar, el cliente solicita la dirección IP del servidor y un nombre de usuario. Escribe la IP del equipo que ejecuta los servidores. Si se escribe `y` para usar la dirección predeterminada, el cliente intenta conectarse a `192.168.1.33`. Luego solicita el nombre de usuario.

### Enviar mensajes

Escribe el mensaje en la interfaz y presiona Enter. El servidor de chat lo retransmite a los demás clientes conectados.

### Compartir un archivo

Escribe `/file` seguido de la ruta local del archivo:

```text
/file /ruta/al/archivo
```

El cliente sube el archivo al servidor de archivos. Cuando la transferencia termina, notifica a los clientes del chat para que puedan descargarlo. El servidor guarda los archivos compartidos en `uploads/`.

### Recibir un archivo

Al recibir una notificación de archivo, el cliente descarga automáticamente el archivo desde el servidor de archivos y lo guarda en `downloads/`. La interfaz muestra el progreso de cada transferencia activa; si hay varias, las presenta en el orden en que comenzaron.

### Controles de la interfaz

- Enter: enviar el texto escrito.
- Page Up / Page Down: desplazarse por los mensajes.
- Ctrl+C o Esc: salir del cliente.

## Almacenamiento y límites

- `uploads/`: directorio relativo al directorio de trabajo del proceso del servidor de archivos.
- `downloads/`: directorio relativo al directorio de trabajo del proceso cliente.
- El servicio de archivos acepta transferencias de hasta 1 GiB.
- Los mensajes de chat tienen un tamaño máximo de 1 KiB.

Los directorios de almacenamiento se crean si no existen. Los archivos con el mismo nombre se guardan usando ese nombre, por lo que una transferencia posterior puede reemplazar un archivo existente.

## Red y seguridad

Los servidores escuchan en todas las interfaces de red (`:8080` y `:8081`). Asegúrate de que los puertos sean accesibles desde los clientes y de que las reglas de firewall permitan las conexiones necesarias. El proyecto está pensado para pruebas o redes de confianza; no implementa autenticación ni cifrado del tráfico.

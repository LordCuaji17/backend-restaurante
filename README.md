# Backend Restaurante - Mariscos El Rinconcito

API REST en Go (Gin Framework) desplegada en la nube para el sistema POS y control de órdenes de meseros.

## 🚀 Despliegue en Producción (Cloud)
- **Servicio:** Render Web Service (Virginia, US-East).
- **URL Base:** `https://backend-restaurante-b2a3.onrender.com`
- **Base de Datos:** TiDB Cloud Serverless (MySQL 8.0 SSL).
- **Integración POS:** Mercado Pago Terminals API v1 (Point Smart).

## 🛠️ Variables de Entorno de Producción
Configuradas en el Dashboard de Render:
```env
PORT=8080
DB_DSN=b2thQBey2Je3Y3i.root:<PASSWORD>@tcp(gateway01.us-east-1.prod.aws.tidbcloud.com:4000)/db_restaurante?tls=true
MP_ACCESS_TOKEN=<TOKEN_PRODUCCION>
MP_TERMINAL_ID=NEWLAND_N950__N950NCD300083779
```

## 💳 Flujo de pago con tarjeta
`POST /api/ordenes/:id/pagar` con `{"metodo_pago": "Tarjeta"}`:
1. El backend crea la orden de cobro en la terminal Point Smart (`POST /v1/orders`) y espera hasta 90 s, consultando el estado cada 2 s.
2. Solo si el pago queda aprobado (`processed`) marca la orden como **Pagada**, libera la mesa e imprime el ticket.
3. Si el pago falla, se cancela, expira o se agota el tiempo, la orden **sigue abierta**, no se imprime ticket y responde `409` con el motivo. Al agotarse el tiempo el cobro se cancela para no bloquear la terminal.
4. Si la terminal tiene otro cobro pendiente, responde `409` pidiendo cancelarlo en la pantalla de la terminal.

El pago en efectivo no cambia: se registra e imprime de inmediato.

> La app Android usa un timeout de lectura de 120 s para esperar el cobro. Tras cambiar `main.go` hay que recompilar el binario `server` (`CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .`) porque el Dockerfile lo copia directamente.

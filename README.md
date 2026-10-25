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
MERCADOPAGO_ACCESS_TOKEN=<TOKEN_PRODUCCION>
POINT_DEVICE_ID=NEWLAND_N950__N950NCD300083779

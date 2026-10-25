# Meseros Rinconcito — Documentación Completa y Bitácora del Proyecto

App móvil Android y Backend REST para meseros de **Mariscos El Rinconcito**.
Incluye control de 4 mesas en salón, **gestión completa de pedidos Para Llevar**, catálogo POS con categorías y búsqueda rápida, cobro inteligente en Efectivo o Tarjeta con precios unitarios integrados (sin desglose de comisión), y **emisión automática de tickets térmicos en terminal Point Smart de Mercado Pago**.

---

## 📌 Contexto de Entorno y Red (Casa vs Negocio)

> [!IMPORTANT]
> **ESTADO ACTUAL:** La aplicación se encuentra en desarrollo y pruebas locales desde **CASA**.
> - **Red de casa actual:** WiFi local `192.168.100.87`.
> - **Dispositivo físico vinculado:** Xiaomi / POCO X3 Pro (`M2102J20SG`, vayu).
> - **Configuración de cliente:** `BASE_URL = "http://192.168.100.87:8080/"` en `RetrofitClient.kt`.
>
> 🚀 **SIGUIENTE PASO PENDIENTE:** Montar el backend Go y la base de datos MariaDB en la nube (ej. Railway / Render / VPS) para tener un dominio fijo HTTPS, olvidarse de las IPs locales y permitir que la app funcione en cualquier red WiFi o con datos móviles sin depender de la laptop prendida.

### ⚠️ Qué hacer al llevar la aplicación y PC al NEGOCIO (Restaurante):
En el negocio la red WiFi tendrá un rango de IP distinto (ej. `192.168.1.X`, `192.168.0.X`):

1. **Conectar PC y Celular a la misma red WiFi del restaurante** (o al punto de acceso / hotspot portátil).
2. **Obtener la nueva IP de la PC en el negocio:**
   ```bash
   ip -4 addr | grep inet
   ```
   *(Identifica la IP en la tarjeta inalámbrica, ej. `192.168.1.45`).*
3. **Actualizar la IP en la app Android:**
   - Abre `~/AndroidStudioProjects/MeserosRinconcito/app/src/main/java/com/example/meserosrinconcito/RetrofitClient.kt`.
   - Modifica `BASE_URL`:
     ```kotlin
     private const val BASE_URL = "http://192.168.1.45:8080/" // <- Tu nueva IP del negocio
     ```
4. **Reinstalar la app en el teléfono:**
   - Por Android Studio dando clic en **Run ▶**.
   - O por terminal:
     ```bash
     cd ~/AndroidStudioProjects/MeserosRinconcito
     JAVA_HOME=/snap/android-studio/current/jbr ./gradlew assembleDebug
     ~/Android/Sdk/platform-tools/adb install -r app/build/outputs/apk/debug/app-debug.apk
     ```
5. **Iniciar el servidor backend en la PC:**
   ```bash
   cd ~/Documents/backend-restaurante
   ./run.sh
   ```
6. **Si el teléfono no conecta, verificar Firewall en Linux:**
   ```bash
   sudo ufw allow 8080/tcp
   ```

*(Alternativa futura sin cambio de IPs: Desplegar el backend en la nube con HTTPS o túnel Cloudflare/ngrok).*

---

## 1. Arquitectura General del Sistema

```
Teléfono Mesero (Xiaomi / POCO X3 Pro)
       │
       │ HTTP / JSON (Puerto 8080 vía WiFi local)
       ▼
Backend Go (Gin Framework) — ~/Documents/backend-restaurante
       ├── MariaDB Local (db_restaurante, 127.0.0.1:3306)
       └── Mercado Pago Terminals API (Impresión directa a Point Smart N950)
```

| Componente | Ubicación local | Descripción |
|---|---|---|
| **App Android** | `~/AndroidStudioProjects/MeserosRinconcito` | Aplicación móvil para meseros (Kotlin + Retrofit + Material3). |
| **Backend Go** | `~/Documents/backend-restaurante` | API REST (`main.go`, `ticket.go`, `.env`, `run.sh`). |
| **Base de Datos** | MariaDB local (`127.0.0.1:3306`) | Base de datos `db_restaurante`. |
| **Hardware POS** | Point Smart (Mercado Pago) | Terminal térmica de 58 mm conectada a internet vía WiFi/4G. |

---

## 2. Flujo Operativo y Pantallas

### A) MainActivity (Mesas Comedor + Pedidos Para Llevar)
- **Mesas en cuadrícula 2×2 (Mesas 1 a 4):**
  - Verde = Mesa Libre.
  - Rojo = Mesa Ocupada.
  - Al presionar cualquier mesa se abre su cuenta en `MesaActivity`.
- **Sección inferior "🛍️ Para Llevar":**
  - Gestiona pedidos que **no ocupan mesa** ni bloquean el salón.
  - Botón **`＋ Nuevo Pedido`**: Abre un diálogo para ingresar el nombre del cliente (opcional, ej. *"Juan"* o vacío).
  - Tarjetas de pedidos activos mostrando hora de creación, total actual y número de platillos.
  - Al pulsar un pedido para llevar, abre su detalle en `MesaActivity`.

### B) MesaActivity (Detalle de Cuenta)
- **Mesa de comedor libre:** Muestra botón `ABRIR CUENTA`.
- **Mesa con cuenta abierta o Pedido Para Llevar:**
  - Encabezado dinámico: `"Mesa 1"` o `"🛍️ Juan"` / `"🛍️ Para Llevar #18"`.
  - Lista de platillos agrupados por producto con cantidad y subtotal acumulado.
  - Botón `＋ AGREGAR PLATILLO`: Abre el menú POS (`ProductosActivity`).
  - Botón `COBRAR CUENTA`: Despliega el modal de liquidación.

### C) Modal Inteligente de Cobro (Efectivo vs Tarjeta)
- **💵 Efectivo:**
  - Respeta los precios de lista normales de cada producto.
  - Al cobrar con opción de imprimir, emite el ticket térmico de consumo en la terminal.
- **💳 Tarjeta (Absorción de comisión integrada y Cobro en Terminal):**
  - **Fórmula de absorción terminal MP:** `Precio Tarjeta = CEIL(Precio Unitario / 0.9594)`
  - Redondea al entero superior por producto para garantizar números enteros limpios y exactos.
  - **Cumplimiento comercial / PROFECO:** **No se muestra ningún apartado ni renglón que diga "Comisión" o "Ajuste por tarjeta"** ni en la app ni en el ticket. Cada platillo ya refleja su precio para tarjeta y la suma coincide de forma transparente.
  - **Disparo automático de cobro en Point Smart:** Al confirmar el cobro en Tarjeta, el backend envía la orden de cobro vía `POST /v1/orders` a Mercado Pago. La terminal Point Smart pita, se ilumina y solicita al cliente: *"Acerque, inserte o deslice su tarjeta"* por el monto total con tarjeta.
- Al cobrar:
  - Si es mesa de comedor: la mesa pasa automáticamente a **'Libre'**.
  - Si es pedido para llevar: se liquida la orden sin afectar ninguna mesa del salón.

### D) ProductosActivity (Catálogo POS)
- Cuadrícula de 2 columnas con tarjetas de comida.
- Chips superiores para filtrar por categorías (`Especialidades`, `Camarones`, `Bebidas`, etc.).
- Buscador en vivo insensible a acentos y mayúsculas (`Normalizer`).
- Botón directo `+` para seleccionar cantidad con NumberPicker y agregar a la cuenta.

---

## 3. Impresión de Tickets en Point Smart (Mercado Pago)

- **Formato térmico de 58 mm (32 columnas exactas UTF-8):**
  - Implementado en `ticket.go`.
  - Control de ancho por conteo de runas UTF-8 (acentos y letra `ñ` no desfasan columnas).
  - Partición de texto multilínea estricta: nombres largos continúan en la columna de producto sin invadir la columna de precio.
- **Encabezado Adaptativo:**
  - En comedor: `Mesa X   Orden #Y`
  - En para llevar: `*** PARA LLEVAR ***` y `Cliente: [Nombre]`
- **Ejemplo Ticket con Tarjeta:**
  ```text
          MARISCOS
        EL RINCONCITO
     *** PARA LLEVAR ***
      Cliente: Juan Pérez
          Orden #18
       25/10/2026 09:05
  --------------------------------
  CANT PRODUCTO            IMPORTE
  --------------------------------
     2 Coctel de Camarón   $292.00
     3 Tostada de Ceviche  $141.00
  --------------------------------
  TOTAL                    $433.00
  MÉTODO DE PAGO:          Tarjeta
  --------------------------------
       ¡Gracias por su visita!
            Vuelva pronto
  ```

---

## 4. Endpoints del Backend REST (Puerto 8080)

| Método | Ruta | Body / Parámetros | Descripción |
|---|---|---|---|
| `GET` | `/api/mesas` | Ninguno | Estado de las mesas (1 a 4). |
| `GET` | `/api/productos` | Ninguno | Lista de platillos y bebidas activas. |
| `GET` | `/api/mesas/:id/orden` | `:id` (ID de mesa) | Retorna orden activa de la mesa. |
| `GET` | `/api/ordenes/para-llevar` | Ninguno | Retorna lista de pedidos abiertos para llevar. |
| `POST` | `/api/ordenes` | `{"mesa_id": 1}` o `{"tipo": "Para Llevar", "cliente_nombre": "Juan"}` | Crea orden de comedor o para llevar. |
| `GET` | `/api/ordenes/:id` | `:id` (ID de orden) | Retorna detalle completo de cualquier orden. |
| `POST` | `/api/ordenes/:id/agregar` | `{"producto_id": 3, "cantidad": 2}` | Agrega platillo (suma cantidad si ya existe). |
| `POST` | `/api/ordenes/:id/pagar` | `{"metodo_pago": "Tarjeta", "imprimir": true}` | Liquida cuenta, libera mesa (si aplica) e imprime. |
| `POST` | `/api/ordenes/:id/imprimir` | Ninguno | Re-impresión directa a la terminal. |

---

## 5. Esquema de Base de Datos (MariaDB `db_restaurante`)

### Tabla `ordenes`:
```sql
CREATE TABLE `ordenes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `mesa_id` int(11) DEFAULT NULL,             -- NULL si es pedido para llevar
  `cliente_nombre` varchar(100) DEFAULT NULL,  -- Nombre opcional para pedidos para llevar
  `tipo` enum('Comedor','Para Llevar') DEFAULT 'Comedor',
  `estado` enum('Abierta','Impresa','Pagada') DEFAULT 'Abierta',
  `fecha_creacion` datetime DEFAULT current_timestamp(),
  `total` decimal(10,2) DEFAULT 0.00,
  `metodo_pago` varchar(20) DEFAULT NULL,
  `comision` decimal(10,2) DEFAULT 0.00,
  PRIMARY KEY (`id`),
  KEY `mesa_id` (`mesa_id`),
  CONSTRAINT `ordenes_ibfk_1` FOREIGN KEY (`mesa_id`) REFERENCES `mesas` (`id`)
);
```

### Tabla `orden_detalles`:
```sql
CREATE TABLE `orden_detalles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `orden_id` int(11) NOT NULL,
  `producto_id` int(11) NOT NULL,
  `cantidad` int(11) NOT NULL,
  `precio_unitario` decimal(10,2) NOT NULL,
  `subtotal` decimal(10,2) NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `orden_detalles_ibfk_1` FOREIGN KEY (`orden_id`) REFERENCES `ordenes` (`id`),
  CONSTRAINT `orden_detalles_ibfk_2` FOREIGN KEY (`producto_id`) REFERENCES `productos` (`id`)
);
```

---

## 6. Variables de Entorno (`.env`) en Backend
Archivo en `~/Documents/backend-restaurante/.env`:
- `MP_ACCESS_TOKEN`: Token de autenticación de Mercado Pago.
- `MP_TERMINAL_ID`: ID del dispositivo Point Smart.
- `DB_DSN`: Conexión MariaDB (`root:1234@tcp(127.0.0.1:3306)/db_restaurante`).
- `PORT`: `8080`.

---

## 7. Comandos de Verificación Rápida
- **Probar Backend (Unit Tests de tickets):**
  ```bash
  cd ~/Documents/backend-restaurante && go test -v ./...
  ```
- **Arrancar Backend:**
  ```bash
  cd ~/Documents/backend-restaurante && ./run.sh
  ```
- **Compilar e instalar APK en teléfono conectado:**
  ```bash
  cd ~/AndroidStudioProjects/MeserosRinconcito
  JAVA_HOME=/snap/android-studio/current/jbr ./gradlew assembleDebug
  ~/Android/Sdk/platform-tools/adb install -r app/build/outputs/apk/debug/app-debug.apk
  ```

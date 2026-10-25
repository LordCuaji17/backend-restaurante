package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// --- ESTRUCTURAS ---

type Producto struct {
	ID        int     `json:"id"`
	Nombre    string  `json:"nombre"`
	Precio    float64 `json:"precio"`
	Categoria string  `json:"categoria"`
	Activo    bool    `json:"activo"`
}

type Mesa struct {
	ID         int    `json:"id"`
	NumeroMesa int    `json:"numero_mesa"`
	Estado     string `json:"estado"`
}

type NuevaOrdenRequest struct {
	MesaID        int    `json:"mesa_id"`
	ClienteNombre string `json:"cliente_nombre"`
	Tipo          string `json:"tipo"` // "Comedor" o "Para Llevar"
}

type AgregarProductoRequest struct {
	ProductoID int `json:"producto_id" binding:"required"`
	Cantidad   int `json:"cantidad" binding:"required,min=1"`
}

type AgregarPersonalizadoRequest struct {
	Nombre      string  `json:"nombre" binding:"required"`
	Precio      float64 `json:"precio" binding:"required"`
	Cantidad    int     `json:"cantidad"`
	GuardarMenu bool    `json:"guardar_menu"`
}

type PagarOrdenRequest struct {
	MetodoPago string  `json:"metodo_pago" binding:"required"` // "Efectivo" o "Tarjeta"
	Imprimir   bool    `json:"imprimir"`
	Propina    float64 `json:"propina"`
}

type QuitarProductoRequest struct {
	ProductoID int `json:"producto_id" binding:"required"`
	Cantidad   int `json:"cantidad"`
}

type ItemOrden struct {
	ProductoID int     `json:"producto_id"`
	Nombre     string  `json:"nombre"`
	Cantidad   int     `json:"cantidad"`
	Subtotal   float64 `json:"subtotal"`
}

type OrdenParaLlevar struct {
	ID            int     `json:"id"`
	ClienteNombre string  `json:"cliente_nombre"`
	Total         float64 `json:"total"`
	Hora          string  `json:"hora"`
	CantidadItems int     `json:"cantidad_items"`
}

type OrdenResponse struct {
	OrdenID       int         `json:"orden_id"`
	MesaID        int         `json:"mesa_id"`
	NumeroMesa    int         `json:"numero_mesa"`
	Tipo          string      `json:"tipo"`
	ClienteNombre string      `json:"cliente_nombre"`
	Estado        string      `json:"estado"`
	Total         float64     `json:"total"`
	Items         []ItemOrden `json:"items"`
}

var db *sql.DB

func main() {
	cargarEnv(".env")
	var err error

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("Falta la variable de entorno DB_DSN (ver .env.example)")
	}

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error al abrir la base de datos:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Error al conectar con MariaDB:", err)
	}
	fmt.Println("Conexión a MariaDB exitosa")

	router := gin.Default()

	// Rutas de la API
	router.GET("/api/productos", getProductos)
	router.GET("/api/mesas", getMesas)
	router.POST("/api/ordenes", crearOrden)
	router.GET("/api/ordenes/para-llevar", getOrdenesParaLlevar)
	router.POST("/api/ordenes/:id/agregar", agregarProducto)
	router.POST("/api/ordenes/:id/quitar", quitarProducto)
	router.POST("/api/ordenes/:id/agregar-personalizado", agregarProductoPersonalizado)
	router.POST("/api/ordenes/:id/imprimir", imprimirTicket)
	router.GET("/api/mesas/:id/orden", getOrdenAbierta)
	router.GET("/api/ordenes/:id", getOrden)
	router.POST("/api/ordenes/:id/pagar", pagarOrden)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run("0.0.0.0:" + port)
}

// --- CONTROLADORES ---

func getProductos(c *gin.Context) {
	rows, err := db.Query("SELECT id, nombre, precio, categoria, activo FROM productos WHERE activo = TRUE")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar los productos"})
		return
	}
	defer rows.Close()

	var productos []Producto
	for rows.Next() {
		var p Producto
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Precio, &p.Categoria, &p.Activo); err != nil {
			continue
		}
		productos = append(productos, p)
	}
	c.JSON(http.StatusOK, productos)
}

func getMesas(c *gin.Context) {
	rows, err := db.Query("SELECT id, numero_mesa, estado FROM mesas")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar las mesas"})
		return
	}
	defer rows.Close()

	var mesas []Mesa
	for rows.Next() {
		var m Mesa
		if err := rows.Scan(&m.ID, &m.NumeroMesa, &m.Estado); err != nil {
			continue
		}
		mesas = append(mesas, m)
	}
	c.JSON(http.StatusOK, mesas)
}

func crearOrden(c *gin.Context) {
	var req NuevaOrdenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos para la orden"})
		return
	}

	if req.Tipo == "" {
		if req.MesaID > 0 {
			req.Tipo = "Comedor"
		} else {
			req.Tipo = "Para Llevar"
		}
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la orden"})
		return
	}

	var ordenID int64
	if req.Tipo == "Para Llevar" {
		res, err := tx.Exec("INSERT INTO ordenes (mesa_id, cliente_nombre, tipo, estado, total) VALUES (NULL, ?, 'Para Llevar', 'Abierta', 0.00)", req.ClienteNombre)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la orden para llevar"})
			return
		}
		ordenID, _ = res.LastInsertId()
	} else {
		res, err := tx.Exec("INSERT INTO ordenes (mesa_id, cliente_nombre, tipo, estado, total) VALUES (?, ?, 'Comedor', 'Abierta', 0.00)", req.MesaID, req.ClienteNombre)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la orden"})
			return
		}
		ordenID, _ = res.LastInsertId()

		_, err = tx.Exec("UPDATE mesas SET estado = 'Ocupada' WHERE id = ?", req.MesaID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar la mesa"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"mensaje":        "Cuenta abierta exitosamente",
		"orden_id":       ordenID,
		"mesa_id":        req.MesaID,
		"tipo":           req.Tipo,
		"cliente_nombre": req.ClienteNombre,
	})
}

func getOrdenesParaLlevar(c *gin.Context) {
	rows, err := db.Query(`
		SELECT o.id, COALESCE(o.cliente_nombre, ''), o.total,
		       DATE_FORMAT(o.fecha_creacion, '%H:%i'),
		       COALESCE(SUM(od.cantidad), 0)
		FROM ordenes o
		LEFT JOIN orden_detalles od ON o.id = od.orden_id
		WHERE o.tipo = 'Para Llevar' AND o.estado != 'Pagada'
		GROUP BY o.id, o.cliente_nombre, o.total, o.fecha_creacion
		ORDER BY o.id DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar pedidos para llevar"})
		return
	}
	defer rows.Close()

	var lista []OrdenParaLlevar
	for rows.Next() {
		var o OrdenParaLlevar
		if err := rows.Scan(&o.ID, &o.ClienteNombre, &o.Total, &o.Hora, &o.CantidadItems); err == nil {
			lista = append(lista, o)
		}
	}
	if lista == nil {
		lista = []OrdenParaLlevar{}
	}
	c.JSON(http.StatusOK, lista)
}

func agregarProducto(c *gin.Context) {
	ordenID := c.Param("id")

	var req AgregarProductoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltan datos del producto o cantidad"})
		return
	}

	var precioUnitario float64
	err := db.QueryRow("SELECT precio FROM productos WHERE id = ?", req.ProductoID).Scan(&precioUnitario)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "El producto no existe en el menú"})
		return
	}

	subtotal := precioUnitario * float64(req.Cantidad)

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno de base de datos"})
		return
	}

	// Si el producto ya está en la cuenta, se suma a esa línea en vez de duplicarla
	var detalleID int
	err = tx.QueryRow("SELECT id FROM orden_detalles WHERE orden_id = ? AND producto_id = ? LIMIT 1 FOR UPDATE",
		ordenID, req.ProductoID).Scan(&detalleID)
	if err == nil {
		_, err = tx.Exec("UPDATE orden_detalles SET cantidad = cantidad + ?, subtotal = subtotal + ? WHERE id = ?",
			req.Cantidad, subtotal, detalleID)
	} else if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO orden_detalles (orden_id, producto_id, cantidad, precio_unitario, subtotal) VALUES (?, ?, ?, ?, ?)",
			ordenID, req.ProductoID, req.Cantidad, precioUnitario, subtotal)
	}
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo registrar el platillo"})
		return
	}

	_, err = tx.Exec("UPDATE ordenes SET total = total + ? WHERE id = ?", subtotal, ordenID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el total de la cuenta"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"mensaje":  "Platillo agregado a la orden",
		"subtotal": subtotal,
	})
}

func quitarProducto(c *gin.Context) {
	ordenID := c.Param("id")

	var req QuitarProductoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltan datos del producto a quitar"})
		return
	}

	cant := req.Cantidad
	if cant <= 0 {
		cant = 1
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}

	var detalleID int
	var cantidadActual int
	var precioUnitario float64
	err = tx.QueryRow("SELECT id, cantidad, precio_unitario FROM orden_detalles WHERE orden_id = ? AND producto_id = ? LIMIT 1 FOR UPDATE",
		ordenID, req.ProductoID).Scan(&detalleID, &cantidadActual, &precioUnitario)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "El platillo no está en la orden"})
		return
	}

	aRestar := cant
	if aRestar > cantidadActual {
		aRestar = cantidadActual
	}
	montoRestar := precioUnitario * float64(aRestar)

	if cantidadActual > aRestar {
		_, err = tx.Exec("UPDATE orden_detalles SET cantidad = cantidad - ?, subtotal = subtotal - ? WHERE id = ?",
			aRestar, montoRestar, detalleID)
	} else {
		_, err = tx.Exec("DELETE FROM orden_detalles WHERE id = ?", detalleID)
	}
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar la orden"})
		return
	}

	_, err = tx.Exec("UPDATE ordenes SET total = GREATEST(0, total - ?) WHERE id = ?", montoRestar, ordenID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar el total"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"mensaje": "Platillo descontado de la orden", "restado": aRestar})
}

func agregarProductoPersonalizado(c *gin.Context) {
	ordenID := c.Param("id")

	var req AgregarPersonalizadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltan datos del platillo (nombre o precio)"})
		return
	}

	nombre := strings.TrimSpace(req.Nombre)
	if nombre == "" || req.Precio <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre y el precio deben ser válidos"})
		return
	}

	cantidad := req.Cantidad
	if cantidad <= 0 {
		cantidad = 1
	}

	subtotal := req.Precio * float64(cantidad)

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de base de datos"})
		return
	}

	activo := 0
	if req.GuardarMenu {
		activo = 1
	}

	var productoID int64
	// Buscar si ya existe un producto con el mismo nombre y precio exacto
	err = tx.QueryRow("SELECT id FROM productos WHERE LOWER(TRIM(nombre)) = LOWER(?) AND precio = ? LIMIT 1",
		nombre, req.Precio).Scan(&productoID)
	if err != nil {
		res, err := tx.Exec("INSERT INTO productos (nombre, precio, categoria, activo) VALUES (?, ?, 'Personalizados', ?)",
			nombre, req.Precio, activo)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo registrar el nuevo platillo"})
			return
		}
		productoID, _ = res.LastInsertId()
	} else if req.GuardarMenu {
		// Si se pidió guardar en menú pero estaba inactivo, activarlo
		tx.Exec("UPDATE productos SET activo = 1 WHERE id = ?", productoID)
	}

	// Registrar o actualizar en orden_detalles
	var detalleID int
	err = tx.QueryRow("SELECT id FROM orden_detalles WHERE orden_id = ? AND producto_id = ? LIMIT 1 FOR UPDATE",
		ordenID, productoID).Scan(&detalleID)
	if err == nil {
		_, err = tx.Exec("UPDATE orden_detalles SET cantidad = cantidad + ?, subtotal = subtotal + ? WHERE id = ?",
			cantidad, subtotal, detalleID)
	} else if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO orden_detalles (orden_id, producto_id, cantidad, precio_unitario, subtotal) VALUES (?, ?, ?, ?, ?)",
			ordenID, productoID, cantidad, req.Precio, subtotal)
	}
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo agregar a la cuenta"})
		return
	}

	_, err = tx.Exec("UPDATE ordenes SET total = total + ? WHERE id = ?", subtotal, ordenID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el total de la orden"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"mensaje":     "Platillo personalizado agregado con éxito",
		"producto_id": productoID,
		"nombre":      nombre,
		"cantidad":    cantidad,
		"subtotal":    subtotal,
	})
}

func getOrdenAbierta(c *gin.Context) {
	var ordenID int
	err := db.QueryRow("SELECT id FROM ordenes WHERE mesa_id = ? AND estado != 'Pagada' ORDER BY id DESC LIMIT 1", c.Param("id")).Scan(&ordenID)
	if err != nil {
		c.JSON(http.StatusOK, OrdenResponse{OrdenID: 0, Estado: "", Items: []ItemOrden{}})
		return
	}
	o, err := cargarOrden(ordenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al cargar la orden"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func getOrden(c *gin.Context) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	o, err := cargarOrden(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "La orden no existe"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func cargarOrden(ordenID int) (OrdenResponse, error) {
	o := OrdenResponse{OrdenID: ordenID, Items: []ItemOrden{}}
	var mesaID sql.NullInt64
	var numeroMesa sql.NullInt64
	var clienteNombre sql.NullString
	var tipo sql.NullString
	err := db.QueryRow(`
		SELECT o.estado, o.total, o.mesa_id, m.numero_mesa, o.cliente_nombre, o.tipo
		FROM ordenes o
		LEFT JOIN mesas m ON o.mesa_id = m.id
		WHERE o.id = ?`, ordenID).Scan(&o.Estado, &o.Total, &mesaID, &numeroMesa, &clienteNombre, &tipo)
	if err != nil {
		return o, err
	}
	if mesaID.Valid {
		o.MesaID = int(mesaID.Int64)
	}
	if numeroMesa.Valid {
		o.NumeroMesa = int(numeroMesa.Int64)
	}
	if clienteNombre.Valid {
		o.ClienteNombre = clienteNombre.String
	}
	if tipo.Valid && tipo.String != "" {
		o.Tipo = tipo.String
	} else {
		o.Tipo = "Comedor"
	}

	rows, err := db.Query(`SELECT p.id, p.nombre, SUM(od.cantidad), SUM(od.subtotal)
		FROM orden_detalles od JOIN productos p ON od.producto_id = p.id
		WHERE od.orden_id = ?
		GROUP BY p.id, p.nombre ORDER BY MIN(od.id)`, ordenID)
	if err != nil {
		return o, err
	}
	defer rows.Close()
	for rows.Next() {
		var it ItemOrden
		if rows.Scan(&it.ProductoID, &it.Nombre, &it.Cantidad, &it.Subtotal) == nil {
			o.Items = append(o.Items, it)
		}
	}
	return o, nil
}

func pagarOrden(c *gin.Context) {
	ordenID := c.Param("id")

	var req PagarOrdenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.MetodoPago = "Efectivo"
		req.Imprimir = true
	}

	var mesaID sql.NullInt64
	var subtotal float64
	if err := db.QueryRow("SELECT mesa_id, total FROM ordenes WHERE id = ?", ordenID).Scan(&mesaID, &subtotal); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "La orden no existe"})
		return
	}

	// Si es Tarjeta, calculamos el precio de cada artículo con comisión integrada (redondeo al entero superior)
	var comision float64 = 0.0
	totalFinal := subtotal
	if req.MetodoPago == "Tarjeta" {
		rows, err := db.Query("SELECT cantidad, precio_unitario FROM orden_detalles WHERE orden_id = ?", ordenID)
		if err == nil {
			var sumaTarjeta float64 = 0.0
			for rows.Next() {
				var cant int
				var pUnit float64
				if rows.Scan(&cant, &pUnit) == nil {
					pTarj := math.Ceil(pUnit / 0.9594)
					sumaTarjeta += pTarj * float64(cant)
				}
			}
			rows.Close()
			if sumaTarjeta > 0 {
				totalFinal = sumaTarjeta
			} else {
				totalFinal = math.Ceil(subtotal / 0.9594)
			}
		} else {
			totalFinal = math.Ceil(subtotal / 0.9594)
		}
		comision = totalFinal - subtotal
	}

	propina := req.Propina
	if propina < 0 {
		propina = 0
	}
	totalFinal += propina

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}
	if _, err = tx.Exec("UPDATE ordenes SET estado = 'Pagada', metodo_pago = ?, comision = ?, propina = ?, total = ? WHERE id = ?",
		req.MetodoPago, comision, propina, totalFinal, ordenID); err == nil {
		if mesaID.Valid {
			_, err = tx.Exec("UPDATE mesas SET estado = 'Libre' WHERE id = ?", mesaID.Int64)
		}
	}
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo cobrar la cuenta"})
		return
	}
	tx.Commit()

	var ticketErr string
	// Imprime el ticket térmico de consumo con el detalle de platillos, propina y totales
	if req.Imprimir {
		if err := procesarImpresion(ordenID, req.MetodoPago, totalFinal, propina); err != nil {
			ticketErr = err.Error()
		}
	}

	if req.MetodoPago == "Tarjeta" {
		// Envía la orden de cobro a la pantalla de la terminal Point Smart
		time.Sleep(500 * time.Millisecond)
		if err := enviarCobroTerminal(ordenID, totalFinal); err != nil {
			if ticketErr != "" {
				ticketErr += " | " + err.Error()
			} else {
				ticketErr = err.Error()
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"mensaje":      "Cuenta cobrada y mesa liberada",
		"metodo_pago":  req.MetodoPago,
		"subtotal":     subtotal,
		"comision":     comision,
		"propina":      propina,
		"total":        totalFinal,
		"ticket_error": ticketErr,
	})
}

func imprimirTicket(c *gin.Context) {
	ordenID := c.Param("id")
	var total float64
	var metodoPago sql.NullString
	var propina float64
	err := db.QueryRow("SELECT total, metodo_pago, COALESCE(propina, 0.0) FROM ordenes WHERE id = ?", ordenID).Scan(&total, &metodoPago, &propina)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "La orden no existe"})
		return
	}

	metodo := ""
	if metodoPago.Valid {
		metodo = metodoPago.String
	}
	if err := procesarImpresion(ordenID, metodo, total, propina); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensaje": "¡Ticket enviado a la terminal exitosamente!",
		"total":   total,
	})
}

func enviarCobroTerminal(ordenID string, totalFinal float64) error {
	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	deviceID := os.Getenv("MP_TERMINAL_ID")
	if accessToken == "" || deviceID == "" {
		return fmt.Errorf("faltan credenciales de Mercado Pago en variables de entorno")
	}

	idempotencyKey := fmt.Sprintf("pay-orden-%s-%d", ordenID, time.Now().UnixNano())
	payload := map[string]interface{}{
		"type":               "point",
		"external_reference": fmt.Sprintf("orden-%s", ordenID),
		"description":        "Mariscos El Rinconcito",
		"transactions": map[string]interface{}{
			"payments": []map[string]interface{}{
				{
					"amount": fmt.Sprintf("%.2f", totalFinal),
				},
			},
		},
		"config": map[string]interface{}{
			"point": map[string]interface{}{
				"terminal_id":       deviceID,
				"print_on_terminal": "seller_ticket",
			},
		},
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.mercadopago.com/v1/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", idempotencyKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, clientErr := client.Do(req)
	if clientErr != nil {
		return clientErr
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Mercado Pago rechazó el cobro en terminal (status %d)", resp.StatusCode)
	}
	return nil
}

func procesarImpresion(ordenID string, metodoPago string, total float64, propina float64) error {
	var mesaNum sql.NullInt64
	var clienteNombre sql.NullString
	var tipo sql.NullString
	err := db.QueryRow(`
		SELECT m.numero_mesa, o.cliente_nombre, o.tipo
		FROM ordenes o LEFT JOIN mesas m ON o.mesa_id = m.id WHERE o.id = ?`, ordenID).Scan(&mesaNum, &clienteNombre, &tipo)
	if err != nil {
		return err
	}

	mesa := 0
	if mesaNum.Valid {
		mesa = int(mesaNum.Int64)
	}
	cli := ""
	if clienteNombre.Valid {
		cli = clienteNombre.String
	}
	tipoStr := "Comedor"
	if tipo.Valid && tipo.String != "" {
		tipoStr = tipo.String
	}

	rows, err := db.Query(`
		SELECT p.nombre, od.precio_unitario, SUM(od.cantidad)
		FROM orden_detalles od
		JOIN productos p ON od.producto_id = p.id
		WHERE od.orden_id = ?
		GROUP BY p.id, p.nombre, od.precio_unitario ORDER BY MIN(od.id)`, ordenID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var lineas []LineaTicket
	var totalCalculado float64 = 0.0
	esTarjeta := metodoPago == "Tarjeta"
	for rows.Next() {
		var l LineaTicket
		var precioUnit float64
		if rows.Scan(&l.Nombre, &precioUnit, &l.Cantidad) == nil {
			if esTarjeta {
				precioTarj := math.Ceil(precioUnit / 0.9594)
				l.Subtotal = precioTarj * float64(l.Cantidad)
			} else {
				l.Subtotal = precioUnit * float64(l.Cantidad)
			}
			totalCalculado += l.Subtotal
			lineas = append(lineas, l)
		}
	}

	if total <= 0 {
		total = totalCalculado + propina
	}

	ticketText := formatearTicket(ordenID, mesa, tipoStr, cli, lineas, metodoPago, total, propina, time.Now())

	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	deviceID := os.Getenv("MP_TERMINAL_ID")
	if accessToken == "" || deviceID == "" {
		return fmt.Errorf("faltan credenciales de Mercado Pago en variables de entorno")
	}

	idempotencyKey := fmt.Sprintf("print-orden-%s-%d", ordenID, time.Now().UnixNano())
	payload := map[string]interface{}{
		"type":               "print",
		"external_reference": idempotencyKey,
		"config": map[string]interface{}{
			"point": map[string]interface{}{
				"terminal_id": deviceID,
				"subtype":     "custom",
			},
		},
		"content": ticketText,
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.mercadopago.com/terminals/v1/actions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", idempotencyKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, clientErr := client.Do(req)
	if clientErr != nil {
		return clientErr
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Mercado Pago rechazó la impresión (status %d)", resp.StatusCode)
	}
	return nil
}

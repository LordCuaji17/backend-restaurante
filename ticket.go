package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// Ancho del ticket en caracteres (papel térmico de 58 mm ≈ 32 columnas).
const anchoTicket = 32

// LineaTicket es un renglón de la cuenta ya agrupado por producto.
type LineaTicket struct {
	Nombre   string
	Cantidad int
	Subtotal float64
}

// padDer rellena con espacios a la derecha hasta n caracteres (cuenta letras, no bytes).
func padDer(s string, n int) string {
	if l := utf8.RuneCountInString(s); l < n {
		return s + strings.Repeat(" ", n-l)
	}
	return s
}

// padIzq rellena con espacios a la izquierda hasta n caracteres.
func padIzq(s string, n int) string {
	if l := utf8.RuneCountInString(s); l < n {
		return strings.Repeat(" ", n-l) + s
	}
	return s
}

// partirTexto divide un nombre largo en varias líneas de hasta n caracteres, cortando en espacios.
func partirTexto(s string, n int) []string {
	var lineas []string
	actual := ""
	for _, palabra := range strings.Fields(s) {
		for utf8.RuneCountInString(palabra) > n { // palabra más larga que la columna
			r := []rune(palabra)
			if actual != "" {
				lineas = append(lineas, actual)
				actual = ""
			}
			lineas = append(lineas, string(r[:n]))
			palabra = string(r[n:])
		}
		switch {
		case actual == "":
			actual = palabra
		case utf8.RuneCountInString(actual)+1+utf8.RuneCountInString(palabra) <= n:
			actual += " " + palabra
		default:
			lineas = append(lineas, actual)
			actual = palabra
		}
	}
	if actual != "" {
		lineas = append(lineas, actual)
	}
	return lineas
}

// formatearTicket arma el texto con las etiquetas de Mercado Pago Point.
// Cada renglón mide exactamente anchoTicket caracteres para que las columnas queden alineadas:
//
//	CANT PRODUCTO            IMPORTE
//	  2  Piña Rellena de     $386.00
//	     Mariscos
func formatearTicket(ordenID string, mesa int, tipo string, clienteNombre string, lineas []LineaTicket, metodoPago string, total float64, propina float64, fecha time.Time) string {
	const (
		colCant   = 5 // "  2 "
		colPrecio = 9 // "$1370.00" + margen
		colNombre = anchoTicket - colCant - colPrecio
	)
	var sb strings.Builder
	izq := func(texto string) { sb.WriteString("{left}" + texto + "{/left}{br}") }
	sep := strings.Repeat("-", anchoTicket)

	sb.WriteString("{center}{w}{b}MARISCOS{/b}{/w}{/center}{br}")
	sb.WriteString("{center}{w}{b}EL RINCONCITO{/b}{/w}{/center}{br}")
	if tipo == "Para Llevar" {
		sb.WriteString("{center}{b}*** PARA LLEVAR ***{/b}{/center}{br}")
		if clienteNombre != "" {
			sb.WriteString(fmt.Sprintf("{center}Cliente: %s{/center}{br}", clienteNombre))
		}
		sb.WriteString(fmt.Sprintf("{center}Orden #%s{/center}{br}", ordenID))
	} else {
		sb.WriteString(fmt.Sprintf("{center}Mesa %d   Orden #%s{/center}{br}", mesa, ordenID))
	}
	sb.WriteString("{center}" + fecha.Format("02/01/2006 15:04") + "{/center}{br}")
	izq(sep)
	izq(padDer("CANT", colCant) + padDer("PRODUCTO", colNombre) + padIzq("IMPORTE", colPrecio))
	izq(sep)

	for _, l := range lineas {
		partes := partirTexto(l.Nombre, colNombre)
		if len(partes) == 0 {
			partes = []string{""}
		}
		izq(padIzq(fmt.Sprintf("%d", l.Cantidad), colCant-1) + " " +
			padDer(partes[0], colNombre) + padIzq(fmt.Sprintf("$%.2f", l.Subtotal), colPrecio))
		for _, resto := range partes[1:] {
			// El resto del nombre queda confinado estrictamente a la columna de producto
			izq(strings.Repeat(" ", colCant) + padDer(resto, colNombre) + strings.Repeat(" ", colPrecio))
		}
	}

	izq(sep)
	if propina > 0 {
		subtotalComida := total - propina
		izq(padDer("SUBTOTAL:", anchoTicket-10) + padIzq(fmt.Sprintf("$%.2f", subtotalComida), 10))
		izq(padDer("PROPINA:", anchoTicket-10) + padIzq(fmt.Sprintf("$%.2f", propina), 10))
		izq(sep)
	}
	sb.WriteString("{left}{b}{w}" + padDer("TOTAL", 6) + padIzq(fmt.Sprintf("$%.2f", total), anchoTicket/2-6) + "{/w}{/b}{/left}{br}")
	if metodoPago != "" {
		izq(padDer("MÉTODO DE PAGO:", anchoTicket-14) + padIzq(metodoPago, 14))
	}
	izq(sep)
	sb.WriteString("{br}{center}¡Gracias por su visita!{/center}{br}")
	sb.WriteString("{center}{s}Vuelva pronto{/s}{/center}{br}{br}{br}")
	return sb.String()
}

// cargarEnv lee un archivo .env (CLAVE=valor) y define las variables que aún no existan.
// En la nube normalmente las variables ya vienen del entorno y este archivo no se usa.
func cargarEnv(ruta string) {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return
	}
	for _, linea := range strings.Split(string(datos), "\n") {
		linea = strings.TrimSpace(linea)
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		if k, v, ok := strings.Cut(linea, "="); ok {
			k = strings.TrimSpace(k)
			if os.Getenv(k) == "" {
				os.Setenv(k, strings.Trim(strings.TrimSpace(v), `"'`))
			}
		}
	}
}

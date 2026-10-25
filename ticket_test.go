package main

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

var etiquetas = regexp.MustCompile(`\{/?[a-z]+\}`)

func TestFormatearTicketEfectivo(t *testing.T) {
	lineas := []LineaTicket{
		{"Piña Rellena de Mariscos", 3, 555},
		{"Refresco 355 ml", 3, 90},
	}
	out := formatearTicket("3", 1, "Comedor", "", lineas, "Efectivo", 645, 0, time.Date(2026, 10, 24, 22, 40, 0, 0, time.UTC))
	vista := strings.ReplaceAll(out, "{br}", "\n")
	vista = etiquetas.ReplaceAllString(vista, "")
	t.Log("\n" + vista)
}

func TestFormatearTicketTarjeta(t *testing.T) {
	// En tarjeta los importes unitarios ya incluyen el cálculo con tarjeta
	// Piña normal $185 -> tarjeta Ceil(185/0.9594) = $193 -> 3 * 193 = $579
	// Refresco normal $30 -> tarjeta Ceil(30/0.9594) = $32 -> 3 * 32 = $96
	// Total = 579 + 96 = $675
	lineas := []LineaTicket{
		{"Piña Rellena de Mariscos", 3, 579},
		{"Refresco 355 ml", 3, 96},
	}
	out := formatearTicket("3", 1, "Comedor", "", lineas, "Tarjeta", 675, 0, time.Date(2026, 10, 24, 22, 40, 0, 0, time.UTC))
	vista := strings.ReplaceAll(out, "{br}", "\n")
	vista = etiquetas.ReplaceAllString(vista, "")
	t.Log("\n" + vista)
}

func TestFormatearTicketTarjetaConPropina(t *testing.T) {
	// Cuenta comida: $675, Propina: 10% de $645 (normal) = $64.50 -> $65.00
	// Total cobrado: $740.00
	lineas := []LineaTicket{
		{"Piña Rellena de Mariscos", 3, 579},
		{"Refresco 355 ml", 3, 96},
	}
	out := formatearTicket("3", 1, "Comedor", "", lineas, "Tarjeta", 740, 65, time.Date(2026, 10, 24, 22, 40, 0, 0, time.UTC))
	vista := strings.ReplaceAll(out, "{br}", "\n")
	vista = etiquetas.ReplaceAllString(vista, "")
	t.Log("\n" + vista)
}

func TestFormatearTicketParaLlevar(t *testing.T) {
	lineas := []LineaTicket{
		{"Coctel de Camarón", 2, 280},
		{"Tostada de Ceviche", 3, 135},
	}
	out := formatearTicket("15", 0, "Para Llevar", "Juan Pérez", lineas, "Efectivo", 415, 0, time.Date(2026, 10, 25, 8, 55, 0, 0, time.UTC))
	vista := strings.ReplaceAll(out, "{br}", "\n")
	vista = etiquetas.ReplaceAllString(vista, "")
	t.Log("\n" + vista)
}

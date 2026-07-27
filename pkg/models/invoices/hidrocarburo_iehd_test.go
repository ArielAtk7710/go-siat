package invoices_test

import (
	"encoding/xml"
	"regexp"
	"strings"
	"testing"
	"time"

	siat "github.com/ron86i/go-siat/v2"
	"github.com/ron86i/go-siat/v2/pkg/models/invoices"
	"github.com/ron86i/go-siat/v2/pkg/utils"
)

// elementOrder extrae los nombres de los elementos hijos directos de <cabecera> o <detalle>
// en el orden en que aparecen en el XML serializado.
func elementOrder(t *testing.T, xmlStr, parent string) []string {
	t.Helper()

	start := strings.Index(xmlStr, "<"+parent+">")
	if start < 0 {
		t.Fatalf("no se encontró <%s> en el XML", parent)
	}
	end := strings.Index(xmlStr[start:], "</"+parent+">")
	if end < 0 {
		t.Fatalf("no se encontró </%s> en el XML", parent)
	}
	block := xmlStr[start+len(parent)+2 : start+end]

	re := regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9]*)[ />]`)
	var names []string
	for _, m := range re.FindAllStringSubmatch(block, -1) {
		names = append(names, m[1])
	}
	return names
}

func assertOrder(t *testing.T, got, want []string, label string) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("%s: %d elementos, se esperaban %d\n  obtenido: %v\n  esperado: %v",
			label, len(got), len(want), got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: posición %d es %q, se esperaba %q", label, i, got[i], want[i])
		}
	}
}

// TestHidrocarburoAlcanzadaIehd_OrdenXSD verifica que la factura del Sector 19 se serialice
// con los elementos exactamente en el orden del XSD ver.23/08/2021. El SIAT rechaza con
// código 939 cualquier desviación del orden.
func TestHidrocarburoAlcanzadaIehd_OrdenXSD(t *testing.T) {
	fecha := time.Date(2026, 7, 26, 10, 30, 0, 0, time.UTC)
	nombre := "CLIENTE DE PRUEBA"
	montoIehd := 12.50
	porcentaje := 5.0

	cabecera := invoices.NewHidrocarburoAlcanzadaIehdCabeceraBuilder().
		WithNitEmisor(123456789).
		WithRazonSocialEmisor("EMPRESA DE PRUEBA S.R.L.").
		WithMunicipio("La Paz").
		WithTelefono(nil).
		WithNumeroFactura(1).
		WithCuf("CUF-DE-PRUEBA").
		WithCufd("CUFD-DE-PRUEBA").
		WithCodigoSucursal(0).
		WithDireccion("Av. Siempre Viva 742").
		WithCodigoPuntoVenta(utils.IntPtr(0)).
		WithFechaEmision(fecha).
		WithNombreRazonSocial(&nombre).
		WithCodigoTipoDocumentoIdentidad(1).
		WithNumeroDocumento("1234567").
		WithComplemento(nil).
		WithCodigoCliente("CLI-001").
		WithCiudad(nil).
		WithNombrePropietario(nil).
		WithNombreRepresentanteLegal(nil).
		WithCondicionPago(nil).
		WithPeriodoEntrega(nil).
		WithCodigoMetodoPago(1).
		WithNumeroTarjeta(nil).
		WithMontoTotal(100).
		WithMontoIehd(&montoIehd).
		WithMontoTotalSujetoIva(100).
		WithCodigoMoneda(1).
		WithTipoCambio(1).
		WithMontoTotalMoneda(100).
		WithDescuentoAdicional(nil).
		WithCodigoExcepcion(nil).
		WithCafc(nil).
		WithLeyenda("Ley N 453").
		WithUsuario("tester").
		Build()

	detalle := invoices.NewHidrocarburoAlcanzadaIehdDetalleBuilder().
		WithActividadEconomica("477300").
		WithCodigoProductoSin(622539).
		WithCodigoProducto("PROD-001").
		WithDescripcion("Diesel Oil").
		WithCantidad(2).
		WithUnidadMedida(1).
		WithPrecioUnitario(50).
		WithMontoDescuento(nil).
		WithSubTotal(100).
		WithPorcentajeIehd(&porcentaje).
		Build()

	factura := invoices.NewHidrocarburoAlcanzadaIehdBuilder().
		WithModalidad(siat.ModalidadElectronica).
		WithCabecera(cabecera).
		AddDetalle(detalle).
		Build()

	raw, err := xml.Marshal(factura)
	if err != nil {
		t.Fatalf("error al serializar: %v", err)
	}
	got := string(raw)

	if !strings.HasPrefix(got, "<facturaElectronicaHidrocarburoAlcanzadaIehd") {
		t.Errorf("elemento raíz incorrecto:\n%.120s", got)
	}

	assertOrder(t, elementOrder(t, got, "cabecera"), []string{
		"nitEmisor", "razonSocialEmisor", "municipio", "telefono", "numeroFactura",
		"cuf", "cufd", "codigoSucursal", "direccion", "codigoPuntoVenta",
		"fechaEmision", "nombreRazonSocial", "codigoTipoDocumentoIdentidad",
		"numeroDocumento", "complemento", "codigoCliente", "ciudad",
		"nombrePropietario", "nombreRepresentanteLegal", "condicionPago",
		"periodoEntrega", "codigoMetodoPago", "numeroTarjeta", "montoTotal",
		"montoIehd", "montoTotalSujetoIva", "codigoMoneda", "tipoCambio",
		"montoTotalMoneda", "descuentoAdicional", "codigoExcepcion", "cafc",
		"leyenda", "usuario", "codigoDocumentoSector",
	}, "cabecera sector 19")

	assertOrder(t, elementOrder(t, got, "detalle"), []string{
		"actividadEconomica", "codigoProductoSin", "codigoProducto", "descripcion",
		"cantidad", "unidadMedida", "precioUnitario", "montoDescuento", "subTotal",
		"porcentajeIehd",
	}, "detalle sector 19")

	if !strings.Contains(got, "<codigoDocumentoSector>19</codigoDocumentoSector>") {
		t.Error("el código de documento sector no quedó fijado en 19")
	}
	if !strings.Contains(got, `<telefono xsi:nil="true">`) {
		t.Error("un campo nillable sin valor debe emitirse con xsi:nil=\"true\"")
	}
}

// TestHidrocarburoNoAlcanzadaIehd_OrdenXSD verifica el Sector 38, idéntico al 19
// salvo por la ausencia de montoIehd y porcentajeIehd. XSD ver.11/03/2022.
func TestHidrocarburoNoAlcanzadaIehd_OrdenXSD(t *testing.T) {
	fecha := time.Date(2026, 7, 26, 10, 30, 0, 0, time.UTC)
	nombre := "CLIENTE DE PRUEBA"

	cabecera := invoices.NewHidrocarburoNoAlcanzadaIehdCabeceraBuilder().
		WithNitEmisor(123456789).
		WithRazonSocialEmisor("EMPRESA DE PRUEBA S.R.L.").
		WithMunicipio("La Paz").
		WithTelefono(nil).
		WithNumeroFactura(1).
		WithCuf("CUF-DE-PRUEBA").
		WithCufd("CUFD-DE-PRUEBA").
		WithCodigoSucursal(0).
		WithDireccion("Av. Siempre Viva 742").
		WithCodigoPuntoVenta(utils.IntPtr(0)).
		WithFechaEmision(fecha).
		WithNombreRazonSocial(&nombre).
		WithCodigoTipoDocumentoIdentidad(1).
		WithNumeroDocumento("1234567").
		WithComplemento(nil).
		WithCodigoCliente("CLI-001").
		WithCiudad(nil).
		WithNombrePropietario(nil).
		WithNombreRepresentanteLegal(nil).
		WithCondicionPago(nil).
		WithPeriodoEntrega(nil).
		WithCodigoMetodoPago(1).
		WithNumeroTarjeta(nil).
		WithMontoTotal(100).
		WithMontoTotalSujetoIva(100).
		WithCodigoMoneda(1).
		WithTipoCambio(1).
		WithMontoTotalMoneda(100).
		WithDescuentoAdicional(nil).
		WithCodigoExcepcion(nil).
		WithCafc(nil).
		WithLeyenda("Ley N 453").
		WithUsuario("tester").
		Build()

	detalle := invoices.NewHidrocarburoNoAlcanzadaIehdDetalleBuilder().
		WithActividadEconomica("477300").
		WithCodigoProductoSin(622539).
		WithCodigoProducto("PROD-001").
		WithDescripcion("Gas Natural").
		WithCantidad(2).
		WithUnidadMedida(1).
		WithPrecioUnitario(50).
		WithMontoDescuento(nil).
		WithSubTotal(100).
		Build()

	factura := invoices.NewHidrocarburoNoAlcanzadaIehdBuilder().
		WithModalidad(siat.ModalidadElectronica).
		WithCabecera(cabecera).
		AddDetalle(detalle).
		Build()

	raw, err := xml.Marshal(factura)
	if err != nil {
		t.Fatalf("error al serializar: %v", err)
	}
	got := string(raw)

	if !strings.HasPrefix(got, "<facturaElectronicaHidrocarburoNoAlcanzadaIehd") {
		t.Errorf("elemento raíz incorrecto:\n%.120s", got)
	}

	assertOrder(t, elementOrder(t, got, "cabecera"), []string{
		"nitEmisor", "razonSocialEmisor", "municipio", "telefono", "numeroFactura",
		"cuf", "cufd", "codigoSucursal", "direccion", "codigoPuntoVenta",
		"fechaEmision", "nombreRazonSocial", "codigoTipoDocumentoIdentidad",
		"numeroDocumento", "complemento", "codigoCliente", "ciudad",
		"nombrePropietario", "nombreRepresentanteLegal", "condicionPago",
		"periodoEntrega", "codigoMetodoPago", "numeroTarjeta", "montoTotal",
		"montoTotalSujetoIva", "codigoMoneda", "tipoCambio", "montoTotalMoneda",
		"descuentoAdicional", "codigoExcepcion", "cafc", "leyenda", "usuario",
		"codigoDocumentoSector",
	}, "cabecera sector 38")

	assertOrder(t, elementOrder(t, got, "detalle"), []string{
		"actividadEconomica", "codigoProductoSin", "codigoProducto", "descripcion",
		"cantidad", "unidadMedida", "precioUnitario", "montoDescuento", "subTotal",
	}, "detalle sector 38")

	if !strings.Contains(got, "<codigoDocumentoSector>38</codigoDocumentoSector>") {
		t.Error("el código de documento sector no quedó fijado en 38")
	}
	if strings.Contains(got, "montoIehd") || strings.Contains(got, "porcentajeIehd") {
		t.Error("el sector 38 no debe emitir campos IEHD")
	}
}

// TestHidrocarburoIehd_Computarizada verifica que WithModalidad cambie el elemento raíz.
func TestHidrocarburoIehd_Computarizada(t *testing.T) {
	casos := []struct {
		nombre   string
		marshal  func() ([]byte, error)
		esperado string
	}{
		{
			nombre: "sector 19",
			marshal: func() ([]byte, error) {
				return xml.Marshal(invoices.NewHidrocarburoAlcanzadaIehdBuilder().
					WithModalidad(siat.ModalidadComputarizada).Build())
			},
			esperado: "<facturaComputarizadaHidrocarburoAlcanzadaIehd",
		},
		{
			nombre: "sector 38",
			marshal: func() ([]byte, error) {
				return xml.Marshal(invoices.NewHidrocarburoNoAlcanzadaIehdBuilder().
					WithModalidad(siat.ModalidadComputarizada).Build())
			},
			esperado: "<facturaComputarizadaHidrocarburoNoAlcanzadaIehd",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raw, err := c.marshal()
			if err != nil {
				t.Fatalf("error al serializar: %v", err)
			}
			if !strings.HasPrefix(string(raw), c.esperado) {
				t.Errorf("elemento raíz es %.60s, se esperaba %s", raw, c.esperado)
			}
		})
	}
}

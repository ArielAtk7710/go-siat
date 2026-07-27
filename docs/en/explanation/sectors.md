# Understanding invoice sectors

<p align="right">
  <a href="../../es/explanation/sectores.md">🇪🇸 Español</a> · <a href="../README.md">Docs index</a>
</p>

SIAT does not have one invoice format. Its regulatory catalog defines **51 document sector codes**, each with its own XSD schema, its own mandatory fields, and its own validation rules. This page explains what a sector is, how the SDK models them, and how to find the one you need.

---

## What a sector is

A *documento sector* is SIAT's way of saying "this kind of business issues this shape of invoice". A hotel invoice needs the guest's stay dates. A mining export invoice needs mineral concentrations and international quotations. A telecom invoice needs a service period. Rather than one giant optional-everything document, SIAT publishes a separate schema per activity.

The sector code appears in two places, and they must agree:

| Where | Field | Who sets it |
| :--- | :--- | :--- |
| Inside the invoice document | `cabecera.codigoDocumentoSector` | The sector builder, with a correct default |
| In the submission request | `WithCodigoDocumentoSector(n)` | You, explicitly |

A mismatch is rejected with code `932 — CODIGO DOCUMENTO SECTOR NO CORRESPONDE AL SERVICIO`. Code `931` means the value itself is not a valid sector, and `940` means your NIT is not authorized for that sector.

> The SDK ships one builder per covered sector, each pre-filled with the right code. You should almost never call `WithCodigoDocumentoSector` on a *header* builder — the default is already correct. You do have to pass the code to the *request* builder.

---

## Every sector builder shares the same shape

All 51 builders in `pkg/models/invoices` (50 codes, one of which has two layouts) follow one pattern, so learning one teaches you all of them:

```go
cabecera := invoices.NewHotelCabeceraBuilder().
    WithNitEmisor(123456789).
    WithNumeroFactura(1).
    WithCuf(cuf).
    // ... sector-specific fields
    Build()

detalle := invoices.NewHotelDetalleBuilder().
    // ... sector-specific line fields
    Build()

factura := invoices.NewHotelBuilder().
    WithModalidad(siat.ModalidadElectronica).
    WithCabecera(cabecera).
    AddDetalle(detalle).
    Build()
```

Three constructors per sector: `New<Sector>CabeceraBuilder()` for the header, `New<Sector>DetalleBuilder()` for a line item, `New<Sector>Builder()` to assemble them. `AddDetalle` is additive — call it once per line.

`WithModalidad` on the root builder controls the XML namespace and whether a signature slot is emitted. It is separate from the `WithCodigoModalidad` you set on the *submission* request, and both need to be set.

What changes between sectors is only the field list on the header and detail builders. Your editor's autocomplete on the returned builder type is the fastest reference there is.

---

## How a sector picks its service endpoint

SIAT exposes twelve SOAP endpoints. Which one accepts your invoice is not a free choice — it is determined by the sector.

**Six sectors have a dedicated endpoint.** They will not be accepted anywhere else:

| Facade | Sectors it serves |
| :--- | :--- |
| `s.CompraVenta()` | CompraVenta (1), CompraVentaBonificaciones (35), CompraVentaTasas (41) |
| `s.Telecomunicaciones()` | Telecomunicaciones (22), TelecomunicacionesZF (49) |
| `s.ServicioBasico()` | ServicioBasico (13), ServicioBasicoZF (40) |
| `s.EntidadFinanciera()` | EntidadFinanciera (15) |
| `s.BoletoAereo()` | BoletoAereo (30) |
| `s.DocumentoAjuste()` | NotaCreditoDebito (24), NotaFiscalCreditoDebito (24), NotaConciliacion (29), NotaCreditoDebitoDescuento (47), NotaCreditoDebitoIce (48) |

**The other 37 sectors route by modality**, not by activity. Pick the facade that matches how you are invoicing:

```go
s.Electronica().RecepcionFactura(ctx, req)     // ModalidadElectronica — signed
s.Computarizada().RecepcionFactura(ctx, req)   // ModalidadComputarizada — control code, no signature
```

Both expose the same nine-method interface, so switching modality is a one-word change at the call site. See [Architecture](architecture.md) for why the interfaces are shaped that way.

`DocumentoAjuste()` is the exception to the whole model: adjustment notes are not invoices, so its methods are named `RecepcionDocumentoAjuste`, `AnulacionDocumentoAjuste` and `ReversionAnulacionDocumentoAjuste` rather than the `*Factura` names.

---

## The full sector catalog

SIAT's regulatory catalog defines **51 document sector codes**. The SDK ships builders for **50 of them**; the single gap is flagged below.

Builder names drop the `New`/`Builder` wrapper — `CompraVenta` means `invoices.NewCompraVentaBuilder()`. In the **Facade** column, *Modality* means `Electronica()` or `Computarizada()` depending on how you are invoicing.

| Code | Description | Tax credit | Builder | Facade |
| ---: | :--- | :--- | :--- | :--- |
| 1 | Sales and purchases | Yes | `CompraVenta` | `CompraVenta()` |
| 2 | Real estate rental | Yes | `AlquilerBienInmueble` | Modality |
| 3 | Commercial export | No | `ComercialExportacion` | Modality |
| 4 | Commercial export on free consignment | No | `LibreConsignacion` | Modality |
| 5 | Free trade zone sales | No | `ZonaFranca` | Modality |
| 6 | Tourism and lodging services | No | `TurismoHospedaje` | Modality |
| 7 | Food security and supply | No | `SeguridadAlimentaria` | Modality |
| 8 | Zero rate — books and international road freight | No | `TasaCero` | Modality |
| 9 | Foreign currency exchange | No | `MonedaExtranjera` | Modality |
| 10 | Duty free | No | `DuttyFree` | Modality |
| 11 | Education sector | Yes | `SectorEducativo` | Modality |
| 12 | Hydrocarbon retail | Yes | `ComercializacionHidro` | Modality |
| 13 | Basic services | Yes | `ServicioBasico` | `ServicioBasico()` |
| 14 | ICE-liable products | Yes | `AlcanzadaIce` | Modality |
| 15 | Financial institutions | Yes | `EntidadFinanciera` | `EntidadFinanciera()` |
| 16 | Hotels | Yes | `Hotel` | Modality |
| 17 | Hospitals / clinics | Yes | `HospitalClinica` | Modality |
| 18 | Games of chance | Yes | `JuegoAzar` | Modality |
| 19 | Hydrocarbons subject to IEHD | Yes | `HidrocarburoAlcanzadaIehd` | Modality |
| 20 | Mineral export | No | `ComercialExportacionMinera` | Modality |
| 21 | Domestic mineral sales | Yes | `VentaMineral` | Modality |
| 22 | Telecommunications | Yes | `Telecomunicaciones` | `Telecomunicaciones()` |
| 23 | Prevalued | Yes | `Prevalorada` | Modality |
| 24 | Credit/debit note | Adjustment | `NotaCreditoDebito` · `NotaFiscalCreditoDebito` | `DocumentoAjuste()` |
| 28 | Commercial export of services | No | `ComercialExportacionServicio` | Modality |
| 29 | Conciliation note | Adjustment | `NotaConciliacion` | `DocumentoAjuste()` |
| 30 | Air ticket | Equivalent doc | `BoletoAereo` | `BoletoAereo()` |
| 31 | Energy supply | Yes | `SuministroEnergia` | Modality |
| **33** | **Zero rate VAT, Law 1613** | No | **— no builder** | — |
| 34 | Insurance | Yes | `Seguros` | Modality |
| 35 | Sales with bonuses | Yes | `CompraVentaBonificaciones` | `CompraVenta()` |
| 36 | Prevalued without tax credit | No | `PrevaloradaSinDerechoCreditoFiscal` | Modality |
| 37 | CNG retail | Yes | `ComercializacionGnv` | Modality |
| 38 | Hydrocarbons not subject to IEHD | Yes | `HidrocarburoNoAlcanzadaIehd` | Modality |
| 39 | Natural gas and LPG retail | Yes | `ComercializacionGnGlp` | Modality |
| 40 | Basic services, free trade zone | No | `ServicioBasicoZF` | `ServicioBasico()` |
| 41 | Sales with non-creditable fees | Yes | `CompraVentaTasas` | `CompraVenta()` |
| 42 | Rental, free trade zone | No | `AlquilerZF` | Modality |
| 43 | Hydrocarbon export | No | `ComercialExportacionHidro` | Modality |
| 44 | Lubricant import and retail | Yes | `ImportacionComercializacionLubricantes` | Modality |
| 45 | Commercial export at sale price | No | `ComercialExportacionPVenta` | Modality |
| 46 | Education, free trade zone | No | `SectorEducativoZF` | Modality |
| 47 | Credit/debit note with discount | Adjustment | `NotaCreditoDebitoDescuento` | `DocumentoAjuste()` |
| 48 | ICE credit/debit note | Adjustment | `NotaCreditoDebitoIce` | `DocumentoAjuste()` |
| 49 | Telecommunications, free trade zone | No | `TelecomunicacionesZF` | `Telecomunicaciones()` |
| 50 | Hospitals / clinics, free trade zone | No | `HospitalClinicaZF` | Modality |
| 51 | Gas bottling plants | Yes | `Engarrafadoras` | Modality |
| 52 | Mineral sales to the Central Bank | No | `VentaMineralBCB` | `Electronica()` only |
| 53 | Lubricant import and retail, IEHD | Yes | `LubricantesIehd` | Modality |
| 54 | Biodiesel / ecological diesel feedstock | No | `Biodiesel` | Modality |
| 55 | Fuel retail | Yes | `VentaCombustibleSinSubvencion` | Modality |

Four things worth reading off this table:

**The codes are not contiguous.** 25, 26, 27 and 32 are absent from SIAT's own catalog. Do not invent them or use them as filler.

**24 appears twice deliberately.** `NotaCreditoDebito` and `NotaFiscalCreditoDebito` are two document layouts sharing one sector code. Pick the one whose field list matches the note you are issuing.

**52 is electronic-only.** Regulation restricts mineral sales to the Central Bank to electronic modality, so it goes through `Electronica()` and always requires a signature.

**`ZF` means *zona franca*.** It is the free-trade-zone variant of a sector, with a different code and extra fields — and almost always without tax credit even where the base sector grants it (compare 13 with 40, 22 with 49, 17 with 50).

### The four hydrocarbon and lubricant sectors

These are the easiest to mix up, because their names overlap and the difference is regulatory rather than structural:

| Code | Builder | Applies to |
| ---: | :--- | :--- |
| 12 | `ComercializacionHidro` | Fuel retail — diesel, gasoline, automotive |
| 19 | `HidrocarburoAlcanzadaIehd` | Activities **liable** for the IEHD |
| 38 | `HidrocarburoNoAlcanzadaIehd` | Activities **exempt** from the IEHD |
| 44 / 53 | `ImportacionComercializacionLubricantes` / `LubricantesIehd` | Lubricant import and retail, without and with IEHD |

Sectors 19 and 38 share an identical layout except that 19 carries `montoIehd` in the header and `porcentajeIehd` per detail line. If you are exempt, 38 rejects those fields; if you are liable, 19 requires them.

### The one uncovered sector

Code 33 exists in SIAT's catalog but has no builder in the SDK:

| Code | Description |
| ---: | :--- |
| 33 | Zero rate VAT, Law 1613 — capital goods and industrial plants |

If your activity falls under it, there is no shortcut inside the SDK: you must build the document struct yourself and feed it in with `WithArchivo` / `WithHashArchivo` instead of `WithFactura`. See [packaging by hand](../how-to/send-invoices.md#packaging-by-hand).

Do not substitute sector 8 (`TasaCero`) for 33. Both are zero-rate, but 8 covers books and international road freight under a different legal basis, and SIAT rejects the mismatch with code 932.

---

## Finding the fields for your sector

The SDK's own tests are the most reliable examples, because they are the code that actually round-trips against SIAT. Every sector has one:

```
pkg/models/invoices/<sector>_test.go
```

For example, `pkg/models/invoices/hotel_test.go` builds a complete hotel invoice with every required field populated. Copy that, change the values, and you have a working document.

Sectors whose codes you do not recognize are best confirmed against SIAT itself rather than guessed — `Sincronizacion()` exposes `SincronizarParametricaTipoDocumentoSector` for the full catalog and `SincronizarListaActividadesDocumentoSector` for which sectors your registered economic activities allow.

---

## Related

| | |
| :--- | :--- |
| Build and send your first invoice | [Tutorial](../tutorial/first-invoice.md) |
| Why the services are shaped this way | [Explanation: Architecture](architecture.md) |
| Signing electronic-modality invoices | [How-to: Sign invoices](../how-to/sign-invoices.md) |
| Sector rejection codes (931/932/940) | [How-to: Handle errors](../how-to/handle-errors.md) |

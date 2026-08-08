package usecases

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOCRFields_InvoiceFixture(t *testing.T) {
	raw := `
FACTURA DE VENTA
EMPRESA DEMO S.A.S
NIT 900.123.456-7
Fecha: 15/03/2026
Vencimiento: 30/03/2026
Factura No. FV-2026-0042
Total a pagar: $1.250.500
COP
Concepto: Servicio de gas marzo
`
	got := parseOCRFields(raw)
	assert.Equal(t, "2026-03-15", got.DocumentDate)
	assert.Equal(t, "2026-03-30", got.DueDate)
	require.NotNil(t, got.AmountCents)
	assert.Equal(t, int64(125050000), *got.AmountCents)
	assert.Equal(t, "1250500", got.Amount)
	assert.Equal(t, "COP", got.Currency)
	assert.Equal(t, "FV-2026-0042", got.ReferenceNumber)
	assert.NotEmpty(t, got.Issuer)
	assert.NotContains(t, strings.ToLower(got.Issuer), "yovanny")
	assert.Contains(t, strings.ToLower(got.Title), "gas")
	assert.Greater(t, got.Confidence, 0.5)
	assert.NotEmpty(t, got.RawExcerpt)
	assert.Empty(t, got.Notes)
}

func TestParseOCRFields_AmountVariants(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		cents   int64
		display string
	}{
		{"dot_thousands", "Total a pagar $2.500.000", 250000000, "2500000"},
		{"plain_labeled", "Total a pagar: 450000", 45000000, "450000"},
		{"decimal_comma", "Total a pagar $89.900,50", 8990050, "89900.50"},
		{"cop_prefix", "Total a pagar COP 1.100.000", 110000000, "1100000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOCRFields(tt.raw)
			require.NotNil(t, got.AmountCents)
			assert.Equal(t, tt.cents, *got.AmountCents)
			assert.Equal(t, tt.display, got.Amount)
		})
	}
}

func TestParseOCRFields_DatesAndIssuer(t *testing.T) {
	raw := "SERVICIUDAD E.S.P\nFecha documento 01-02-2026\nVence el 10/02/2026\nRecibo ABC-991"
	got := parseOCRFields(raw)
	assert.Equal(t, "2026-02-01", got.DocumentDate)
	assert.Equal(t, "2026-02-10", got.DueDate)
	assert.Equal(t, "ABC-991", got.ReferenceNumber)
	assert.Equal(t, "SERVICIUDAD", got.Issuer)
}

func TestParseOCRFields_Empty(t *testing.T) {
	got := parseOCRFields("   \n\t")
	assert.Empty(t, got.Title)
	assert.Empty(t, got.Amount)
	assert.Equal(t, "COP", got.Currency)
	assert.Empty(t, got.Notes)
}

func TestParseOCRFields_EfigasUtilityBill(t *testing.T) {
	// Realistic (slightly noisy) OCR of an Efigas / Brilla gas bill.
	raw := `
Efigas Ahí siempre
Brilla Creciendo contigo
Fecha de expedición 29/07/2026 6:57:40 a. m.
Fecha de generación 29/07/2026 01:00:50 a. m.
HEBERT YOVANNY LOPEZ CASTANEDA
Nro de identificación 10031632
Número del contrato 1286243
Servicio público $61.029
Brilla
Otros servicios
TOTAL A PAGAR $61.029
Fecha de vencimiento 12/08/2026
Fecha de suspensión 13/08/2026
Forma de pago Crédito
Factura electrónica de venta EFIR18208812
Referencia de pago 199294196
Nro de estado de cuenta 1221015587
SERVICIO DE GAS
Último pago Fecha 06/07/2026 VALOR $62.037 ENTIDAD DE RECAUDO AV VILLAS PSE
Tu cupo aprobado Brilla es de: $5.585.000
`
	got := parseOCRFields(raw)

	assert.Equal(t, "Efigas", got.Issuer, "issuer must be the utility, not the customer")
	assert.NotContains(t, strings.ToLower(got.Issuer), "lopez")
	assert.Equal(t, "2026-07-29", got.DocumentDate)
	assert.Equal(t, "2026-08-12", got.DueDate, "due date must come from vencimiento, not último pago")
	require.NotNil(t, got.AmountCents)
	assert.Equal(t, int64(6102900), *got.AmountCents, "amount must be TOTAL A PAGAR, not último pago nor cupo Brilla")
	assert.Equal(t, "61029", got.Amount)
	assert.Equal(t, "199294196", got.ReferenceNumber)
	assert.Contains(t, strings.ToLower(got.Title), "gas")
	assert.Contains(t, got.Title, "Efigas")
	assert.NotContains(t, strings.ToLower(got.Title), "hebert")
	assert.Empty(t, got.Notes, "notes stay empty when extras are present")
	assert.NotContains(t, got.Notes, "Último pago")

	keys := map[string]string{}
	for _, f := range got.ExtraFields {
		keys[f.Key] = f.Value
	}
	assert.Equal(t, "1286243", keys[extraKeyContract])
	assert.Equal(t, "EFIR18208812", keys[extraKeyElectronicInv])
	assert.Equal(t, "2026-08-13", keys[extraKeySuspension])
	assert.Equal(t, "5585000", keys[extraKeyBrillaCupo])
	assert.Equal(t, "Crédito", keys[extraKeyPaymentMethod])
	assert.NotContains(t, keys, extraKeyPaymentRef, "primary ref already holds payment reference")
}

func TestParseOCRFields_ServiciudadRealOCR(t *testing.T) {
	// Excerpt taken from a real Tesseract response (SERVICIUDAD bill).
	raw := `
Ss PS a iS DOCUMENTO EQUIVALEN TE ELECT! NIT: B16.001.609-1 // NUIR:1-66170000-2 RONICO DE-SERVICIOS PUBLICOS Documento Y — 274711110 Total a pagar => $90.650 AGENTE RESENEION DEWA Fecha expedición o4/ago 2026 11:27:01 Ultimo día de pago 19/ago/2026 = == AA EA SERVICIUDAD: eli Moses de deuda 1 Suspensión el día
`
	got := parseOCRFields(raw)
	assert.Equal(t, "SERVICIUDAD", got.Issuer)
	assert.Equal(t, "90650", got.Amount)
	assert.Equal(t, "2026-08-04", got.DocumentDate)
	assert.Equal(t, "2026-08-19", got.DueDate)
	assert.Equal(t, "274711110", got.ReferenceNumber)
	require.NotEmpty(t, got.ExtraFields)
	keys := map[string]string{}
	for _, f := range got.ExtraFields {
		keys[f.Key] = f.Value
	}
	assert.Equal(t, "B16.001.609-1", keys[extraKeyNIT])
	assert.Equal(t, "1-66170000-2", keys[extraKeyNUIR])
	assert.Contains(t, strings.ToLower(got.Title), "documento equival")
	assert.Empty(t, got.Notes)
}

func TestParseOCRFields_EfigasNoisyOCR(t *testing.T) {
	// Closer to real Tesseract noise: broken lines, typos, symbols.
	raw := `
Y tt Ves €Cce alsa Atendemos todas tus solicitudes
Efigas | Brilla
29/07/2026 6:57:40 a €
AEBERT YOVANNY LOPEZ CASTANEDA
Servicio público $61.029
TOTAL A PAGAR $61.029
Fecha de
vencimiento
12/08/2026
Referencla
de pago
199294196
EFIR18208812
contralo 1286243
cupo aprobado Brilla es de $5.585.000
Último pago Fecha 06/07/2026 VALOR $62.037
vewres Penovevres
`
	got := parseOCRFields(raw)
	assert.Equal(t, "Efigas", got.Issuer)
	assert.Equal(t, "61029", got.Amount)
	assert.Equal(t, "2026-08-12", got.DueDate)
	assert.Equal(t, "199294196", got.ReferenceNumber)
	assert.NotEqual(t, "vewres", got.ReferenceNumber)
	assert.NotContains(t, strings.ToLower(got.Title), "€cce")
	assert.Empty(t, got.Notes)
	require.NotEmpty(t, got.ExtraFields)
	keys := map[string]string{}
	for _, f := range got.ExtraFields {
		keys[f.Key] = f.Value
	}
	assert.Equal(t, "EFIR18208812", keys[extraKeyElectronicInv])
	assert.Equal(t, "1286243", keys[extraKeyContract])
	assert.Equal(t, "5585000", keys[extraKeyBrillaCupo])
}

package usecases

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
)

const (
	ocrExcerptMaxLen       = 400
	ocrTitleMaxLen         = 255
	ocrIssuerMaxLen        = 255
	ocrRefMaxLen           = 128
	ocrIssuerScanLines     = 12
	ocrTitleScanLines      = 20
	ocrMinIssuerLetters    = 3
	ocrIssuerUpperRatio    = 0.55
	ocrMinSubmatchAmount   = 2
	ocrMinSubmatchDate     = 4
	ocrTwoDigitYearCutoff  = 100
	ocrTwoDigitYearBase    = 2000
	ocrYearTokenLen        = 4
	ocrSplitDotParts       = 2
	ocrCentsPerUnit        = 100
	ocrMinAmountCents      = 1000
	ocrDecimalPadThreshold = 10
	ocrFractionDigits      = 2
	ocrThousandsGroupLen   = 3
	ocrMaxIssuerLineLen    = 80
	ocrMinTitleLineLen     = 4
	ocrMinCalendarYear     = 1990
	ocrMaxCalendarYear     = 2100
	ocrConfTitle           = 0.2
	ocrConfIssuer          = 0.2
	ocrConfDocDate         = 0.15
	ocrConfDueDate         = 0.1
	ocrConfAmount          = 0.2
	ocrConfReference       = 0.15
	ocrMaxNoisyRatio       = 0.28
	ocrMinRefDigits        = 1
	ocrMinRefLen           = 3
	ocrContextWindow       = 48
	ocrDueLabelLookahead   = 2
	ocrSubmatchIndexLen    = 4
	ocrExtraFieldsCap      = 8
	ocrKeywordWindow       = 96

	kwReferenciaPago     = "referencia de pago"
	kwFormaPago          = "forma de pago"
	kwVencimiento        = "vencimiento"
	kwExpedicion         = "expedicion"          //nolint:misspell // Spanish OCR (sin tilde)
	kwFechaExpedicion    = "fecha expedicion"    //nolint:misspell // Spanish OCR (sin tilde)
	kwFechaDeExpedicion  = "fecha de expedicion" //nolint:misspell // Spanish OCR (sin tilde)
	kwExpedicionAccent   = "expedición"
	kwFechaExpedicionAcc = "fecha expedición"
	kwFechaDeExpedicionA = "fecha de expedición"

	extraKeyContract       = "contract_number"
	extraKeyServiceNumber  = "service_number"
	extraKeyObligation     = "obligation"
	extraKeyMortgageCredit = "mortgage_credit"
	extraKeyElectronicInv  = "electronic_invoice"
	extraKeyIdentificacion = "identification_number"
	extraKeyAccountStatus  = "account_status"
	extraKeyPaymentRef     = "payment_reference"
	extraKeySuspension     = "suspension_date"
	extraKeyPaymentMethod  = "payment_method"
	extraKeyBrillaCupo     = "brilla_cupo"
	extraKeyNIT            = "nit"
	extraKeyNUIR           = "nuir"
	extraKeyDocNumber      = "document_number"
)

func knownIssuerTokens() []string {
	return []string{
		"efigas", "brilla", "serviciudad", "epm", "vanti", "gas natural",
		"codensa", "enel", "air-e", "afinia", "emcali", "acuavalle",
	}
}

var (
	reDateDMY = regexp.MustCompile(`\b(\d{1,2})[/\-.](\d{1,2})[/\-.](\d{2,4})\b`)
	reDateYMD = regexp.MustCompile(`\b(\d{4})[/\-.](\d{1,2})[/\-.](\d{1,2})\b`)

	reTotalAPagar = regexp.MustCompile(
		`(?i)total\s+a\s+pagar\s*[=:\->\s]*\$?\s*((?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d{1,2})?(?:\.\d{2})?)`,
	)
	reServicioPublicoAmount = regexp.MustCompile(
		`(?i)servicio\s+p[uú]blico\s*[:\-]?\s*\$?\s*((?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d{1,2})?(?:\.\d{2})?)`,
	)
	reAmountMarked = regexp.MustCompile(
		`(?i)(?:\$|COP)\s*((?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d{1,2})?(?:\.\d{2})?)`,
	)
	reAmountNear = regexp.MustCompile(
		`(?i)(?:total(?:\s+factura|\s+servicio|\s+a\s+pagar)?|valor(?:\s+total)?|importe)` +
			`\s*[:\-]?\s*\$?\s*((?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d{1,2})?(?:\.\d{2})?)`,
	)
	reAmountPlain = regexp.MustCompile(
		`(?i)\b((?:\d{1,3}(?:\.\d{3})+)(?:,\d{1,2})?|\d{4,}(?:[.,]\d{2})?)\b`,
	)

	reRefPago = regexp.MustCompile(
		`(?i)referen\w*\s+de\s+pago\s*[:\-]?\s*([A-Z0-9][A-Z0-9\-_/]{3,40})`,
	)
	reFacturaElectronica = regexp.MustCompile(
		`(?i)factura\s+electr\w*(?:\s+de\s+venta)?\s*[:\-]?\s*([A-Z0-9][A-Z0-9\-_/]{4,40})`,
	)
	reEFIRCode = regexp.MustCompile(`(?i)\b(EFIR[A-Z0-9]{5,24})\b`)
	reContrato = regexp.MustCompile(
		`(?i)(?:n[uú]mero\s+del\s+)?contr\w*\s*[:\-]?\s*([0-9]{5,20})`,
	)
	reRefNear = regexp.MustCompile(
		`(?i)(?:factura\s*(?:n[ºo°.]?|no\.?|#)|recibo\s*(?:n[ºo°.]?|no\.?|#)?|` +
			`ref(?:erencia)?)\s*[:#]?\s*([A-Z0-9][A-Z0-9\-_/]{3,40})`,
	)
	reDigitsToken = regexp.MustCompile(`([0-9]{5,20})`)
	reDateToken   = regexp.MustCompile(`(\d{1,2}[/\-.]\d{1,2}[/\-.]\d{2,4})`)
	reDateMonthES = regexp.MustCompile(
		`(?i)\b([oO0-9]?\d)[/\-.\s]+(ene|feb|mar|abr|may|jun|jul|ago|sep|sept|oct|nov|dic)[a-z]*[/\-.\s]+(\d{2,4})\b`,
	)
	reMoneyTokenCap = regexp.MustCompile(`\$?\s*((?:\d{1,3}(?:\.\d{3})+|\d{4,}))`)

	reNITLine  = regexp.MustCompile(`(?i)\bN\.?\s*I\.?\s*T\.?\b`)
	reNITValue = regexp.MustCompile(
		`(?i)\bN\.?\s*I\.?\s*T\.?\s*[:\-]?\s*([A-Z0-9][A-Z0-9.\-]{5,24})`,
	)
	reNUIRValue = regexp.MustCompile(
		`(?i)\bNUIR\s*[:\-]?\s*([0-9][0-9\-.]{5,24})`,
	)
	reDocNumber = regexp.MustCompile(
		`(?i)documento\s+\w*\s*[—\-–:]+\s*([0-9]{6,20})`,
	)
	reDueLabel = regexp.MustCompile(
		`(?i)fecha\s+de\s+vencimiento|vencimiento|vence(?:\s+el)?|fecha\s+l[ií]mite|` +
			`pagar\s+antes|[uú]ltimo\s+d[ií]a\s+de\s+pago|ultimo\s+dia\s+de\s+pago`,
	)
	reDocDateLabel = regexp.MustCompile(
		`(?i)fecha\s+de\s+expedici[oó]n|fecha\s+expedici[oó]n|fecha\s+de\s+generaci[oó]n|` +
			`fecha\s+de\s+emisi[oó]n|expedici[oó]n|fecha\s+documento`,
	)
	// Last payment made. Does not match "último día de pago" (due date): that has words between.
	reLastPaymentCtx = regexp.MustCompile(
		`(?i)[uú]ltimo\s+pago|pago\s+anterior|entidad\s+de\s+recaudo`,
	)
	reAmountExcludeCtx = regexp.MustCompile(
		`(?i)[uú]ltimo\s+pago|pago\s+anterior|cupo\s+aprobado|brilla\s+es\s+de|` +
			`cr[eé]dito\s+aprobado|saldo\s+a\s+favor`,
	)
	rePersonName = regexp.MustCompile(
		`(?i)^[A-ZÁÉÍÓÚÑ]{2,}(?:\s+[A-ZÁÉÍÓÚÑ]{2,}){1,4}$`,
	)
	reCompanyHint = regexp.MustCompile(
		`(?i)\b(s\.?\s*a\.?\s*s\.?|e\.?\s*s\.?\s*p\.?|s\.?\s*a\.?|ltda|inc)\b`,
	)
	reIdentificacion = regexp.MustCompile(
		`(?i)(?:nro\.?\s*de\s*)?identific\w*\s*[:\-]?\s*([0-9]{5,20})`,
	)
	reEstadoCuenta = regexp.MustCompile(
		`(?i)(?:nro\.?\s*de\s*)?estado\s+de\s+cuenta\s*[:\-]?\s*([0-9]{5,20})`,
	)
	reSuspensionLabel = regexp.MustCompile(
		`(?i)fecha\s+de\s+suspens\w*|suspens\w*`,
	)
	reFormaPago = regexp.MustCompile(
		`(?i)forma\s+de\s+pago\s*[:\-]?\s*([A-Za-zÁÉÍÓÚáéíóúñÑ]{3,40})`,
	)
	reCupoBrilla = regexp.MustCompile(
		`(?i)cupo\s+aprobado\s+brilla\s+es\s+de\s*[:\-]?\s*\$?\s*((?:\d{1,3}(?:\.\d{3})+|\d+))`,
	)
	reDueDateInline = regexp.MustCompile(
		`(?i)(?:(?:fecha\s+de\s+)?vencim\w*|[uú]ltimo\s+d[ií]a\s+de\s+pago)` +
			`\s*[:\-=]*\s*` +
			`((?:\d{1,2}[/\-.]\d{1,2}[/\-.]\d{2,4})|(?:[oO0-9]?\d[/\-.\s]+(?:ene|feb|mar|abr|may|jun|jul|ago|sep|sept|oct|nov|dic)[a-z]*[/\-.\s]+\d{2,4}))`,
	)
	reDocDateInline = regexp.MustCompile(
		`(?i)fecha\s+(?:de\s+)?expedici\w*\s*[:\-=]*\s*` +
			`((?:\d{1,2}[/\-.]\d{1,2}[/\-.]\d{2,4})|(?:[oO0-9]?\d[/\-.\s]+(?:ene|feb|mar|abr|may|jun|jul|ago|sep|sept|oct|nov|dic)[a-z]*[/\-.\s]+\d{2,4}))`,
	)
	reNroServicio = regexp.MustCompile(
		`(?i)(?:nro|n[uú]mero)\s*(?:de\s*)?servicio\s*[:\-]?\s*([0-9]{5,12})`,
	)
	reObligacion = regexp.MustCompile(
		`(?i)obligaci\w*\s*[:\-]?\s*([A-Z0-9][A-Z0-9\-_/]{4,40})`,
	)
	reCreditoHip = regexp.MustCompile(
		`(?i)cr[eé]dito\s+hipotecario\s*[:\-]?\s*([A-Z0-9][A-Z0-9\-_/]{4,40})`,
	)
)

type parsedOCRFields struct {
	Title           string
	Issuer          string
	DocumentDate    string
	DueDate         string
	Amount          string
	AmountCents     *int64
	Currency        string
	ReferenceNumber string
	Notes           string
	ExtraFields     []dtos.ExtraFieldDTO
	RawExcerpt      string
	Confidence      float64
}

func parseOCRFields(raw string) parsedOCRFields {
	text := normalizeOCRText(raw)
	out := parsedOCRFields{
		Currency:   "COP",
		RawExcerpt: truncateRunes(collapseWS(text), ocrExcerptMaxLen),
	}
	if text == "" {
		return out
	}

	lines := splitNonEmptyLines(text)
	flat := flattenOCR(text)
	out.Issuer = pickIssuer(text, lines)
	out.ReferenceNumber = pickReference(text, flat)
	out.DocumentDate, out.DueDate = pickDates(text, lines, flat)
	out.AmountCents, out.Amount = pickAmount(text)
	out.Title = pickTitle(text, lines, out.Issuer)
	out.ExtraFields = pickExtraFields(text, lines, flat, out.ReferenceNumber)
	// Do not dump OCR into notes; extras carry structured leftovers.
	out.Notes = ""
	out.Confidence = scoreConfidence(out)
	return out
}

func normalizeOCRText(raw string) string {
	text := strings.ReplaceAll(raw, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "|", " ")
	text = strings.ReplaceAll(text, "€", " ")
	text = strings.TrimSpace(text)
	return text
}

func flattenOCR(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// captureAfterKeywords finds the first keyword and returns the first regex capture in a short window after it.
func captureAfterKeywords(text string, keywords []string, valueRe *regexp.Regexp) string {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		idx := strings.Index(lower, kw)
		if idx < 0 {
			continue
		}
		start := idx + len(kw)
		end := start + ocrKeywordWindow
		if end > len(text) {
			end = len(text)
		}
		if start >= end {
			continue
		}
		window := text[start:end]
		if m := valueRe.FindStringSubmatch(window); len(m) >= ocrMinSubmatchAmount {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func splitNonEmptyLines(text string) []string {
	raw := strings.Split(text, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func pickIssuer(text string, lines []string) string {
	lower := strings.ToLower(text)
	for _, known := range knownIssuerTokens() {
		if strings.Contains(lower, known) {
			return truncateRunes(displayIssuerName(known), ocrIssuerMaxLen)
		}
	}
	for _, line := range lines {
		if reNITLine.MatchString(line) {
			cleaned := reNITLine.ReplaceAllString(line, "")
			cleaned = strings.TrimSpace(strings.Trim(cleaned, ":#-"))
			if cleaned != "" && !isGenericLine(cleaned) && !looksLikePersonName(cleaned) && !isNoisyOCRLine(cleaned) {
				return truncateRunes(cleaned, ocrIssuerMaxLen)
			}
		}
	}
	for _, line := range lines[:min(ocrIssuerScanLines, len(lines))] {
		if reCompanyHint.MatchString(line) && !isNoisyOCRLine(line) {
			return truncateRunes(cleanIssuerLine(line), ocrIssuerMaxLen)
		}
	}
	for _, line := range lines[:min(ocrIssuerScanLines, len(lines))] {
		if looksLikeIssuer(line) && !looksLikePersonName(line) {
			return truncateRunes(cleanIssuerLine(line), ocrIssuerMaxLen)
		}
	}
	return ""
}

func displayIssuerName(known string) string {
	switch known {
	case "efigas":
		return "Efigas"
	case "brilla":
		return "Brilla"
	case "serviciudad":
		return "SERVICIUDAD"
	case "gas natural":
		return "Gas Natural"
	default:
		if known == "" {
			return ""
		}
		r := []rune(known)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
}

func cleanIssuerLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "|·•-–—:")
	return strings.Join(strings.Fields(line), " ")
}

func looksLikeIssuer(line string) bool {
	line = cleanIssuerLine(line)
	if isGenericLine(line) || isNoisyOCRLine(line) || looksLikePersonName(line) {
		return false
	}
	if len(line) < ocrMinIssuerLetters || len(line) > ocrMaxIssuerLineLen {
		return false
	}
	letters := 0
	upper := 0
	for _, r := range line {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	if letters < ocrMinIssuerLetters {
		return false
	}
	return float64(upper)/float64(letters) >= ocrIssuerUpperRatio
}

func looksLikePersonName(line string) bool {
	line = cleanIssuerLine(line)
	if !rePersonName.MatchString(line) {
		return false
	}
	parts := strings.Fields(line)
	return len(parts) >= 2 && len(parts) <= 5
}

func isGenericLine(line string) bool {
	l := strings.ToLower(strings.TrimSpace(line))
	generics := []string{
		"factura", "recibo", "cuenta de cobro", "total a pagar", "total",
		"fecha", kwVencimiento, "nit", "page", "página", "www.", "http",
		"señor", "senor", "cliente", "direccion", "dirección",
		"servicio público", "servicio publico", "atendemos", "brilla creciendo",
		"número del contrato", "numero del contrato", "nro de identificación",
		"nro de identificacion", kwReferenciaPago, kwFormaPago,
	}
	for _, g := range generics {
		if l == g || strings.HasPrefix(l, g+" ") {
			return true
		}
	}
	return false
}

func isNoisyOCRLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return true
	}
	letters := 0
	weird := 0
	for _, r := range line {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), unicode.IsSpace(r):
			if unicode.IsLetter(r) {
				letters++
			}
		case r == '.' || r == ',' || r == '-' || r == '/' || r == ':' || r == '$' || r == '#':
			// normal punctuation
		default:
			weird++
		}
	}
	total := len([]rune(line))
	if total == 0 {
		return true
	}
	if float64(weird)/float64(total) > ocrMaxNoisyRatio {
		return true
	}
	// Mostly symbols / fragments with almost no letters.
	return letters > 0 && letters < ocrMinIssuerLetters && weird > 0
}

func pickReference(text, flat string) string {
	for _, re := range []*regexp.Regexp{reRefPago, reFacturaElectronica, reEFIRCode, reContrato, reDocNumber, reRefNear} {
		if ref := firstCapture(re, text); isPlausibleReference(ref) {
			return truncateRunes(ref, ocrRefMaxLen)
		}
	}
	if ref := captureAfterKeywords(flat, []string{
		kwReferenciaPago, "referen", "ref. pago", "ref pago",
	}, reDigitsToken); isPlausibleReference(ref) {
		return truncateRunes(ref, ocrRefMaxLen)
	}
	if ref := firstCapture(reEFIRCode, flat); isPlausibleReference(ref) {
		return truncateRunes(ref, ocrRefMaxLen)
	}
	if ref := firstCapture(reDocNumber, flat); isPlausibleReference(ref) {
		return truncateRunes(ref, ocrRefMaxLen)
	}
	return ""
}

func firstCapture(re *regexp.Regexp, text string) string {
	m := re.FindStringSubmatch(text)
	if len(m) < ocrMinSubmatchAmount {
		return ""
	}
	return strings.TrimSpace(m[1])
}

type ocrExtraCollector struct {
	out     []dtos.ExtraFieldDTO
	seenVal map[string]struct{}
}

func newOCRExtraCollector(primaryRef string) *ocrExtraCollector {
	c := &ocrExtraCollector{
		out:     make([]dtos.ExtraFieldDTO, 0, ocrExtraFieldsCap),
		seenVal: map[string]struct{}{},
	}
	if primaryRef = strings.TrimSpace(primaryRef); primaryRef != "" {
		c.seenVal[strings.ToLower(primaryRef)] = struct{}{}
	}
	return c
}

func (c *ocrExtraCollector) add(key, label, value string) {
	value = strings.TrimSpace(value)
	if key == "" || label == "" || value == "" || len(c.out) >= maxExtraFields {
		return
	}
	lv := strings.ToLower(value)
	if _, dup := c.seenVal[lv]; dup {
		return
	}
	c.seenVal[lv] = struct{}{}
	c.out = append(c.out, dtos.ExtraFieldDTO{
		Key: key, Label: label, Value: truncateRunes(value, maxExtraValueLen),
	})
}

func (c *ocrExtraCollector) addRef(key, label, value string) {
	if isPlausibleReference(value) {
		c.add(key, label, value)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func pickExtraFields(text string, lines []string, flat, primaryRef string) []dtos.ExtraFieldDTO {
	c := newOCRExtraCollector(primaryRef)
	c.addRef(extraKeyContract, "Contrato", firstNonEmpty(
		firstCapture(reContrato, text),
		captureAfterKeywords(flat, []string{"contrato", "contraro", "contralo"}, reDigitsToken),
	))
	c.addRef(extraKeyElectronicInv, "Factura electrónica", firstNonEmpty(
		firstCapture(reFacturaElectronica, text),
		firstCapture(reEFIRCode, flat),
	))
	c.addRef(extraKeyIdentificacion, "Identificación", firstNonEmpty(
		firstCapture(reIdentificacion, text),
		captureAfterKeywords(flat, []string{"identificacion", "identificación", "nro de identific"}, reDigitsToken),
	))
	c.addRef(extraKeyAccountStatus, "Estado de cuenta", firstNonEmpty(
		firstCapture(reEstadoCuenta, text),
		captureAfterKeywords(flat, []string{"estado de cuenta"}, reDigitsToken),
	))
	refPago := firstNonEmpty(
		firstCapture(reRefPago, text),
		captureAfterKeywords(flat, []string{kwReferenciaPago, "ref. pago", "ref pago"}, reDigitsToken),
	)
	if isPlausibleReference(refPago) && (primaryRef == "" || !strings.EqualFold(refPago, primaryRef)) {
		c.add(extraKeyPaymentRef, "Referencia de pago", refPago)
	}
	susp := dateNearLabel(lines, reSuspensionLabel)
	if susp == "" {
		susp = firstDateIn(captureAfterKeywords(flat, []string{"suspension", "suspensión"}, reDateToken))
	}
	if susp != "" {
		c.add(extraKeySuspension, "Fecha de suspensión", susp)
	}
	method := firstCapture(reFormaPago, text)
	if method == "" || isNoisyOCRLine(method) || len(method) < 3 {
		method = captureAfterKeywords(flat, []string{kwFormaPago}, regexp.MustCompile(`(?i)([A-Za-zÁÉÍÓÚáéíóúñÑ]{3,40})`))
	}
	if method != "" && !isNoisyOCRLine(method) {
		c.add(extraKeyPaymentMethod, "Forma de pago", method)
	}
	cupo := firstNonEmpty(
		firstCapture(reCupoBrilla, text),
		captureAfterKeywords(flat, []string{"cupo aprobado", "cupo brilla", "brilla es de"}, reMoneyTokenCap),
	)
	if _, display, ok := parseMoneyToken(cupo); ok {
		c.add(extraKeyBrillaCupo, "Cupo Brilla", display)
	}
	c.addRef(extraKeyServiceNumber, "N° servicio", firstCapture(reNroServicio, text))
	c.addRef(extraKeyObligation, "Obligación", firstCapture(reObligacion, text))
	c.addRef(extraKeyMortgageCredit, "Crédito hipotecario", firstCapture(reCreditoHip, text))
	if nit := firstCapture(reNITValue, flat); nit != "" {
		c.add(extraKeyNIT, "NIT", nit)
	}
	if nuir := firstCapture(reNUIRValue, flat); nuir != "" {
		c.add(extraKeyNUIR, "NUIR", nuir)
	}
	if docNo := firstCapture(reDocNumber, flat); isPlausibleReference(docNo) &&
		(primaryRef == "" || !strings.EqualFold(docNo, primaryRef)) {
		c.add(extraKeyDocNumber, "N° documento", docNo)
	}
	return c.out
}

func isPlausibleReference(ref string) bool {
	if len(ref) < ocrMinRefLen || isNoisyOCRLine(ref) {
		return false
	}
	digits := 0
	for _, r := range ref {
		if unicode.IsDigit(r) {
			digits++
		}
	}
	// Prefer refs that include digits (invoice / payment / contract numbers).
	return digits >= ocrMinRefDigits
}

func pickDates(text string, lines []string, flat string) (docDate, dueDate string) {
	dueDate = dateNearLabel(lines, reDueLabel)
	if dueDate == "" {
		if m := reDueDateInline.FindStringSubmatch(flat); len(m) >= ocrMinSubmatchAmount {
			dueDate = parseAnyDate(m[1])
		}
	}
	if dueDate == "" {
		dueDate = parseAnyDate(captureAfterKeywords(flat, []string{
			"ultimo dia de pago", "último día de pago", "ultimo día de pago",
			kwVencimiento, "vencim", "vence el", "vence",
		}, reDateMonthES))
	}
	if dueDate == "" {
		dueDate = parseAnyDate(captureAfterKeywords(flat, []string{
			"ultimo dia de pago", "último día de pago", kwVencimiento, "vence",
		}, reDateToken))
	}

	docDate = dateNearLabel(lines, reDocDateLabel)
	if docDate == "" {
		if m := reDocDateInline.FindStringSubmatch(flat); len(m) >= ocrMinSubmatchAmount {
			docDate = parseAnyDate(m[1])
		}
	}
	if docDate == "" {
		docDate = parseAnyDate(captureAfterKeywords(flat, []string{
			kwFechaExpedicion, kwFechaDeExpedicion, kwFechaExpedicionAcc, kwFechaDeExpedicionA,
			kwExpedicion, kwExpedicionAccent,
		}, reDateMonthES))
	}
	if docDate == "" {
		docDate = firstDateOutsideContexts(text, lines, true)
	}
	if dueDate != "" && dueDate == docDate {
		dueDate = ""
	}
	return docDate, dueDate
}

func dateNearLabel(lines []string, label *regexp.Regexp) string {
	for i, line := range lines {
		if !label.MatchString(line) {
			continue
		}
		// Skip "último pago" blocks even if they mention "fecha".
		window := joinLinesWindow(lines, i, ocrDueLabelLookahead)
		if reLastPaymentCtx.MatchString(window) && !reDueLabel.MatchString(line) {
			continue
		}
		// Prefer the date after the label (OCR often packs several dates in one line).
		search := window
		if loc := label.FindStringIndex(window); loc != nil {
			search = window[loc[1]:]
		}
		if iso := firstDateIn(search); iso != "" {
			return iso
		}
		if iso := firstDateIn(window); iso != "" {
			return iso
		}
	}
	return ""
}

func joinLinesWindow(lines []string, i, after int) string {
	end := min(len(lines), i+1+after)
	return strings.Join(lines[i:end], " ")
}

func firstDateOutsideContexts(text string, lines []string, skipLastPayment bool) string {
	for i, line := range lines {
		window := joinLinesWindow(lines, i, 1)
		if skipLastPayment && reLastPaymentCtx.MatchString(window) {
			continue
		}
		if reDueLabel.MatchString(line) {
			continue
		}
		if iso := firstDateIn(line); iso != "" {
			return iso
		}
	}
	_ = text
	return ""
}

func firstDateIn(s string) string {
	return parseAnyDate(s)
}

func parseAnyDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if m := reDateMonthES.FindStringSubmatch(s); len(m) >= ocrMinSubmatchDate {
		if iso := normalizeSpanishMonthDate(m[1], m[2], m[3]); iso != "" {
			return iso
		}
	}
	if m := reDateDMY.FindStringSubmatch(s); len(m) > 0 {
		return normalizeDateParts(m, false)
	}
	if m := reDateYMD.FindStringSubmatch(s); len(m) > 0 {
		return normalizeDateParts(m, true)
	}
	return ""
}

func spanishMonthNumber(mon string) int {
	switch strings.ToLower(strings.TrimSpace(mon)) {
	case "ene", "enero":
		return int(time.January)
	case "feb", "febrero":
		return int(time.February)
	case "mar", "marzo":
		return int(time.March)
	case "abr", "abril":
		return int(time.April)
	case "may", "mayo":
		return int(time.May)
	case "jun", "junio":
		return int(time.June)
	case "jul", "julio":
		return int(time.July)
	case "ago", "agosto":
		return int(time.August)
	case "sep", "sept", "septiembre", "set", "setiembre":
		return int(time.September)
	case "oct", "octubre":
		return int(time.October)
	case "nov", "noviembre":
		return int(time.November)
	case "dic", "diciembre":
		return int(time.December)
	default:
		return 0
	}
}

func normalizeSpanishMonthDate(dayRaw, monthRaw, yearRaw string) string {
	dayRaw = strings.TrimSpace(dayRaw)
	dayRaw = strings.ReplaceAll(dayRaw, "o", "0")
	dayRaw = strings.ReplaceAll(dayRaw, "O", "0")
	d, err := strconv.Atoi(dayRaw)
	if err != nil {
		return ""
	}
	mo := spanishMonthNumber(monthRaw)
	if mo == 0 {
		return ""
	}
	y, err := strconv.Atoi(strings.TrimSpace(yearRaw))
	if err != nil {
		return ""
	}
	if y < ocrTwoDigitYearCutoff {
		y += ocrTwoDigitYearBase
	}
	if d < 1 || d > 31 || y < ocrMinCalendarYear || y > ocrMaxCalendarYear {
		return ""
	}
	t := time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.UTC)
	if t.Day() != d || int(t.Month()) != mo || t.Year() != y {
		return ""
	}
	return t.Format("2006-01-02")
}

func normalizeDateParts(m []string, ymd bool) string {
	if len(m) < ocrMinSubmatchDate {
		return ""
	}
	var y, mo, d int
	var err error
	if ymd {
		y, err = strconv.Atoi(m[1])
		if err != nil {
			return ""
		}
		mo, err = strconv.Atoi(m[2])
		if err != nil {
			return ""
		}
		d, err = strconv.Atoi(m[3])
		if err != nil {
			return ""
		}
	} else {
		d, err = strconv.Atoi(m[1])
		if err != nil {
			return ""
		}
		mo, err = strconv.Atoi(m[2])
		if err != nil {
			return ""
		}
		y, err = strconv.Atoi(m[3])
		if err != nil {
			return ""
		}
		if y < ocrTwoDigitYearCutoff {
			y += ocrTwoDigitYearBase
		}
	}
	if mo < 1 || mo > 12 || d < 1 || d > 31 || y < ocrMinCalendarYear || y > ocrMaxCalendarYear {
		return ""
	}
	t := time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.UTC)
	if t.Day() != d || int(t.Month()) != mo || t.Year() != y {
		return ""
	}
	return t.Format("2006-01-02")
}

func pickAmount(text string) (*int64, string) {
	// Highest priority: explicit TOTAL A PAGAR / servicio público.
	for _, re := range []*regexp.Regexp{reTotalAPagar, reServicioPublicoAmount} {
		if cents, display, ok := firstAmountMatch(text, re, false); ok {
			return cents, display
		}
	}
	if cents, display, ok := firstAmountMatch(text, reAmountNear, true); ok {
		return cents, display
	}
	if cents, display, ok := bestAmountMatch(text, reAmountMarked, true); ok {
		return cents, display
	}
	if cents, display, ok := bestAmountMatch(text, reAmountPlain, true); ok {
		return cents, display
	}
	return nil, ""
}

func firstAmountMatch(text string, re *regexp.Regexp, excludeBadCtx bool) (*int64, string, bool) {
	matches := re.FindAllStringSubmatchIndex(text, -1)
	for _, idx := range matches {
		if len(idx) < ocrSubmatchIndexLen {
			continue
		}
		token := text[idx[2]:idx[3]]
		if excludeBadCtx && amountInExcludedContext(text, idx[0]) {
			continue
		}
		if looksLikeYearOrNITFragment(token, text) {
			continue
		}
		cents, display, ok := parseMoneyToken(token)
		if !ok {
			continue
		}
		c := cents
		return &c, display, true
	}
	return nil, "", false
}

func bestAmountMatch(text string, re *regexp.Regexp, excludeBadCtx bool) (*int64, string, bool) {
	matches := re.FindAllStringSubmatchIndex(text, -1)
	var bestCents int64
	var bestDisplay string
	counts := map[int64]int{}
	displays := map[int64]string{}
	for _, idx := range matches {
		if len(idx) < ocrSubmatchIndexLen {
			continue
		}
		token := text[idx[2]:idx[3]]
		if excludeBadCtx && amountInExcludedContext(text, idx[0]) {
			continue
		}
		if looksLikeYearOrNITFragment(token, text) {
			continue
		}
		cents, display, ok := parseMoneyToken(token)
		if !ok {
			continue
		}
		counts[cents]++
		displays[cents] = display
		if cents > bestCents {
			bestCents = cents
			bestDisplay = display
		}
	}
	// Prefer amounts that appear more than once (e.g. servicio + total).
	var multiCents int64
	var multiDisplay string
	multiCount := 1
	for c, n := range counts {
		if n > multiCount || (n == multiCount && c > multiCents) {
			multiCount = n
			multiCents = c
			multiDisplay = displays[c]
		}
	}
	if multiCount > 1 {
		c := multiCents
		return &c, multiDisplay, true
	}
	if bestDisplay == "" {
		return nil, "", false
	}
	c := bestCents
	return &c, bestDisplay, true
}

func amountInExcludedContext(text string, pos int) bool {
	start := pos - ocrContextWindow
	if start < 0 {
		start = 0
	}
	end := pos + ocrContextWindow
	if end > len(text) {
		end = len(text)
	}
	window := text[start:end]
	return reAmountExcludeCtx.MatchString(window) || reLastPaymentCtx.MatchString(window)
}

func looksLikeYearOrNITFragment(token, full string) bool {
	compact := strings.ReplaceAll(strings.ReplaceAll(token, ".", ""), ",", "")
	if len(compact) == ocrYearTokenLen {
		if y, err := strconv.Atoi(compact); err == nil && y >= ocrMinCalendarYear && y <= ocrMaxCalendarYear {
			return true
		}
	}
	// Only reject when the token is part of the NIT value itself. A single OCR line
	// often contains both NIT and "total a pagar", so matching the whole line is too broad.
	nitCompact := strings.ReplaceAll(strings.ReplaceAll(token, ".", ""), "-", "")
	for _, m := range reNITValue.FindAllStringSubmatch(full, -1) {
		if len(m) < ocrMinSubmatchAmount {
			continue
		}
		nit := m[1]
		if strings.Contains(nit, token) {
			return true
		}
		if nitCompact != "" && strings.Contains(strings.ReplaceAll(strings.ReplaceAll(nit, ".", ""), "-", ""), nitCompact) {
			return true
		}
	}
	return false
}

func parseMoneyToken(raw string) (cents int64, display string, ok bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, "", false
	}
	switch {
	case strings.Contains(s, ","):
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	case strings.Count(s, ".") > 1:
		s = strings.ReplaceAll(s, ".", "")
	case strings.Count(s, ".") == 1:
		parts := strings.SplitN(s, ".", ocrSplitDotParts)
		if len(parts[1]) == ocrThousandsGroupLen && isAllDigits(parts[1]) {
			s = parts[0] + parts[1]
		}
	}
	parts := strings.SplitN(s, ".", ocrSplitDotParts)
	intPart := parts[0]
	if intPart == "" {
		return 0, "", false
	}
	for _, r := range intPart {
		if !unicode.IsDigit(r) {
			return 0, "", false
		}
	}
	whole, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil || whole < 0 {
		return 0, "", false
	}
	frac := int64(0)
	if len(parts) == ocrSplitDotParts {
		fp := parts[1]
		if len(fp) > ocrFractionDigits {
			fp = fp[:ocrFractionDigits]
		}
		for len(fp) < ocrFractionDigits {
			fp += "0"
		}
		frac, err = strconv.ParseInt(fp, 10, 64)
		if err != nil {
			return 0, "", false
		}
	}
	cents = whole*ocrCentsPerUnit + frac
	if cents < ocrMinAmountCents {
		return 0, "", false
	}
	if frac == 0 {
		display = strconv.FormatInt(whole, 10)
	} else {
		display = strconv.FormatInt(whole, 10) + "." + pad2(frac)
	}
	return cents, display, true
}

func pad2(n int64) string {
	if n < ocrDecimalPadThreshold {
		return "0" + strconv.FormatInt(n, 10)
	}
	return strconv.FormatInt(n, 10)
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func pickTitle(text string, lines []string, issuer string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "servicio de gas") || strings.Contains(lower, "consumo de gas"):
		return titledWithIssuer("Factura de gas", issuer)
	case strings.Contains(lower, "documento equival"):
		return titledWithIssuer("Documento equivalente de servicios públicos", issuer)
	case strings.Contains(lower, "servicio público") || strings.Contains(lower, "servicio publico") ||
		strings.Contains(lower, "servicios publicos") || strings.Contains(lower, "servicios públicos"):
		return titledWithIssuer("Factura de servicios públicos", issuer)
	case strings.Contains(lower, "predial"):
		return titledWithIssuer("Predial", issuer)
	}

	for _, line := range lines[:min(ocrTitleScanLines, len(lines))] {
		if isNoisyOCRLine(line) || looksLikePersonName(line) {
			continue
		}
		l := strings.ToLower(line)
		if strings.Contains(l, "factura") || strings.Contains(l, "recibo") ||
			strings.Contains(l, "cuenta de cobro") {
			title := cleanIssuerLine(line)
			return titledWithIssuer(title, issuer)
		}
	}
	if issuer != "" {
		return truncateRunes("Documento — "+issuer, ocrTitleMaxLen)
	}
	for _, line := range lines {
		if isGenericLine(line) || isNoisyOCRLine(line) || looksLikePersonName(line) {
			continue
		}
		if len(line) >= ocrMinTitleLineLen {
			return truncateRunes(cleanIssuerLine(line), ocrTitleMaxLen)
		}
	}
	return ""
}

func titledWithIssuer(base, issuer string) string {
	base = cleanIssuerLine(base)
	if issuer == "" {
		return truncateRunes(base, ocrTitleMaxLen)
	}
	if strings.Contains(strings.ToLower(base), strings.ToLower(issuer)) {
		return truncateRunes(base, ocrTitleMaxLen)
	}
	return truncateRunes(base+" — "+issuer, ocrTitleMaxLen)
}

func scoreConfidence(f parsedOCRFields) float64 {
	score := 0.0
	if f.Title != "" {
		score += ocrConfTitle
	}
	if f.Issuer != "" {
		score += ocrConfIssuer
	}
	if f.DocumentDate != "" {
		score += ocrConfDocDate
	}
	if f.DueDate != "" {
		score += ocrConfDueDate
	}
	if f.Amount != "" {
		score += ocrConfAmount
	}
	if f.ReferenceNumber != "" {
		score += ocrConfReference
	}
	if score > 1 {
		score = 1
	}
	return score
}

func collapseWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

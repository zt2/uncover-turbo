package compiler

import "github.com/zt2/uncover-turbo/pkg/queryir"

// Engine field mappings below are v1 best-effort approximations based on each
// engine's public query syntax. They are intentionally an open set: fields we
// are not confident about for a given engine are omitted, so a query using them
// yields ErrUnsupported (and that engine is skipped) rather than a silently
// wrong query. Refine against official docs by editing these tables.

func init() {
	register(fofa())
	register(hunter())
	register(censys())
	register(zoomeye())
}

// fofa: `field="value"`, `&&` / `||`, field-level `!=`, parenthesized groups.
func fofa() *dialect {
	return &dialect{
		name: "fofa",
		fields: map[queryir.Field]string{
			queryir.FieldIP:       "ip",
			queryir.FieldPort:     "port",
			queryir.FieldDomain:   "domain",
			queryir.FieldHost:     "host",
			queryir.FieldTitle:    "title",
			queryir.FieldBody:     "body",
			queryir.FieldProduct:  "product",
			queryir.FieldCountry:  "country",
			queryir.FieldOrg:      "org",
			queryir.FieldASN:      "asn",
			queryir.FieldProtocol: "protocol",
			queryir.FieldStatus:   "status_code",
			queryir.FieldCertCN:   "cert.subject",
			queryir.FieldOS:       "os",
		},
		format:       quotedFormat("=", "!="),
		and:          " && ",
		or:           " || ",
		fieldLevelNe: true,
		paren:        true,
	}
}

// hunter (qianxin): `field="value"`, `&&` / `||`, field-level `!=`.
func hunter() *dialect {
	return &dialect{
		name: "hunter",
		fields: map[queryir.Field]string{
			queryir.FieldIP:       "ip",
			queryir.FieldPort:     "ip.port",
			queryir.FieldDomain:   "domain",
			queryir.FieldTitle:    "web.title",
			queryir.FieldBody:     "web.body",
			queryir.FieldProduct:  "app.name",
			queryir.FieldCountry:  "ip.country",
			queryir.FieldProtocol: "protocol",
			// host/org/asn/status/cert/os intentionally omitted (uncertain) -> ErrUnsupported.
		},
		format:       quotedFormat("=", "!="),
		and:          " && ",
		or:           " || ",
		fieldLevelNe: true,
		paren:        true,
	}
}

// censys: `field: value`, `and` / `or`, unary `not (...)`, parenthesized groups.
func censys() *dialect {
	return &dialect{
		name: "censys",
		fields: map[queryir.Field]string{
			queryir.FieldIP:       "ip",
			queryir.FieldPort:     "services.port",
			queryir.FieldDomain:   "names",
			queryir.FieldHost:     "names",
			queryir.FieldTitle:    "services.http.response.html_title",
			queryir.FieldBody:     "services.http.response.body",
			queryir.FieldProduct:  "services.software.product",
			queryir.FieldCountry:  "location.country_code",
			queryir.FieldOrg:      "autonomous_system.name",
			queryir.FieldASN:      "autonomous_system.asn",
			queryir.FieldProtocol: "services.service_name",
			queryir.FieldStatus:   "services.http.response.status_code",
			queryir.FieldCertCN:   "services.tls.certificates.leaf_data.subject.common_name",
			queryir.FieldOS:       "operating_system.product",
		},
		format:   censysFormat,
		and:      " and ",
		or:       " or ",
		not:      "not ",
		notGroup: true,
		paren:    true,
	}
}

// zoomeye: `key:"value"`, space = implicit AND, `-` = token negation. No OR
// operator and no group negation in the classic dork syntax, so those yield
// ErrUnsupported.
func zoomeye() *dialect {
	return &dialect{
		name: "zoomeye",
		fields: map[queryir.Field]string{
			queryir.FieldIP:       "ip",
			queryir.FieldPort:     "port",
			queryir.FieldDomain:   "site",
			queryir.FieldHost:     "hostname",
			queryir.FieldTitle:    "title",
			queryir.FieldProduct:  "app",
			queryir.FieldCountry:  "country",
			queryir.FieldOrg:      "org",
			queryir.FieldASN:      "asn",
			queryir.FieldProtocol: "service",
			queryir.FieldCertCN:   "ssl",
			queryir.FieldOS:       "os",
			// body/status intentionally omitted (no classic dork field) -> ErrUnsupported.
		},
		format:   colonQuotedFormat,
		and:      " ",
		or:       "", // no OR in classic dork -> ErrUnsupported
		not:      "-",
		notGroup: false,
		paren:    false,
	}
}

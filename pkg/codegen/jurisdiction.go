// Package codegen — jurisdiction.go
//
// Phase 3: Jurisdiction-specific clause library.
//
// Each jurisdiction entry supplies the boilerplate legal text that varies by
// governing law: the governing law statement in the header, and jurisdiction-
// specific provisions in §5 and §8 (Dispute Resolution).
//
// Selecting a jurisdiction via --jurisdiction affects only boilerplate text.
// Core obligations, transitions, and termination clauses (§3–§4) are always
// derived deterministically from the AST (REQ-3.1 / REQ-3.2).
//
// Valid values: common (default), delaware, california, uk
package codegen

import "strings"

// JurisdictionData holds jurisdiction-specific legal text for a single governing law.
type JurisdictionData struct {
	// Code is the canonical lowercase identifier (e.g. "delaware").
	Code string

	// DisplayName is the human-readable name shown in the document (e.g. "State of Delaware").
	DisplayName string

	// LegalSystemName is used in the §2 Definitions catchall sentence,
	// e.g. "the laws of the State of Delaware" or "common law".
	LegalSystemName string

	// GoverningLawLine is the short line printed at the top of the document,
	// e.g. "State of Delaware".
	GoverningLawLine string

	// GoverningClause is the full "governed by …" sentence fragment used in
	// the preamble and §5.
	GoverningClause string

	// DisputeResolutionClause is the full §8 Dispute Resolution paragraph.
	DisputeResolutionClause string

	// SeverabilityClause allows jurisdictions to override the standard §5.2 text.
	SeverabilityClause string

	// AdditionalProvisions holds any jurisdiction-specific extra sections
	// (e.g. statutory disclosures) appended to §5.
	AdditionalProvisions []JurisdictionProvision
}

// JurisdictionProvision is a numbered sub-section within §5.
type JurisdictionProvision struct {
	Heading string
	Body    string
}

// validJurisdictions is the canonical set of supported jurisdiction codes.
var validJurisdictions = map[string]bool{
	"common":     true,
	"delaware":   true,
	"california": true,
	"uk":         true,
	"pakistan":   true,
}

// IsValidJurisdiction returns true if code is a supported jurisdiction.
func IsValidJurisdiction(code string) bool {
	return validJurisdictions[strings.ToLower(code)]
}

// GetJurisdiction returns the JurisdictionData for the given code.
// Unknown codes fall back to "common".
func GetJurisdiction(code string) JurisdictionData {
	switch strings.ToLower(code) {
	case "delaware":
		return jurisdictionDelaware
	case "california":
		return jurisdictionCalifornia
	case "uk":
		return jurisdictionUK
	case "pakistan":
		return jurisdictionPakistan
	default:
		return jurisdictionCommon
	}
}

// ---------------------------------------------------------------------------
// Jurisdiction definitions
// ---------------------------------------------------------------------------

var jurisdictionCommon = JurisdictionData{
	Code:             "common",
	DisplayName:      "Common Law",
	LegalSystemName:  "applicable common law",
	GoverningLawLine: "Common Law",
	GoverningClause: "governed by and construed in accordance with applicable common law " +
		"principles, without reference to any specific jurisdiction's conflict of laws rules",
	DisputeResolutionClause: "Any dispute, controversy, or claim arising out of or relating to this " +
		"Agreement, or the breach, termination, or invalidity thereof, shall be resolved " +
		"by binding arbitration in accordance with the rules of a mutually agreed arbitral " +
		"body, or, failing agreement, through proceedings before a court of competent " +
		"jurisdiction applying common law principles.",
	SeverabilityClause: "If any provision of this Agreement is held by a court of competent " +
		"jurisdiction to be invalid, illegal, or unenforceable, the remaining provisions " +
		"shall continue in full force and effect as if the invalid provision had never " +
		"been included.",
	AdditionalProvisions: nil,
}

var jurisdictionDelaware = JurisdictionData{
	Code:             "delaware",
	DisplayName:      "State of Delaware",
	LegalSystemName:  "the laws of the State of Delaware",
	GoverningLawLine: "State of Delaware",
	GoverningClause: "governed by and construed in accordance with the laws of the State of " +
		"Delaware, excluding its conflict of laws provisions",
	DisputeResolutionClause: "Any dispute, controversy, or claim arising out of or relating to this " +
		"Agreement shall be exclusively resolved in the Court of Chancery of the State of " +
		"Delaware, or, if the Court of Chancery lacks jurisdiction, in the Superior Court " +
		"of the State of Delaware or the United States District Court for the District of " +
		"Delaware. Each party hereby irrevocably consents to the personal jurisdiction of " +
		"such courts and waives any objection to venue therein.",
	SeverabilityClause: "If any provision of this Agreement is held by a court of competent " +
		"jurisdiction to be invalid, illegal, or unenforceable under the laws of the State " +
		"of Delaware, the remaining provisions shall continue in full force and effect. The " +
		"parties further agree that any such invalid provision shall be modified to the " +
		"minimum extent necessary to make it enforceable.",
	AdditionalProvisions: []JurisdictionProvision{
		{
			Heading: "Delaware Statutory Authority",
			Body: "The parties acknowledge that this Agreement is entered into in reliance " +
				"upon the laws of the State of Delaware, including the Delaware General " +
				"Corporation Law (8 Del. C. § 101 et seq.) where applicable, and that the " +
				"courts of Delaware shall have authority to interpret and enforce the terms " +
				"of this Agreement in accordance with such laws.",
		},
	},
}

var jurisdictionCalifornia = JurisdictionData{
	Code:             "california",
	DisplayName:      "State of California",
	LegalSystemName:  "the laws of the State of California",
	GoverningLawLine: "State of California",
	GoverningClause: "governed by and construed in accordance with the laws of the State of " +
		"California, excluding its conflict of laws provisions",
	DisputeResolutionClause: "Any dispute, controversy, or claim arising out of or relating to this " +
		"Agreement, or the breach thereof, shall first be submitted to non-binding mediation " +
		"before a mutually agreed mediator in the State of California. If mediation fails to " +
		"resolve the dispute within sixty (60) days, the parties agree to submit the dispute " +
		"to binding arbitration administered by JAMS in accordance with its Comprehensive " +
		"Arbitration Rules and Procedures, with the arbitration seated in California. " +
		"Judgment on the award rendered by the arbitrator may be entered in any court " +
		"having jurisdiction thereof.",
	SeverabilityClause: "If any provision of this Agreement is held by a court of competent " +
		"jurisdiction to be invalid, illegal, or unenforceable under California law, the " +
		"remaining provisions shall continue in full force and effect.",
	AdditionalProvisions: []JurisdictionProvision{
		{
			Heading: "Consumer Protection Notice",
			Body: "Nothing in this Agreement shall be construed to waive any non-waivable " +
				"rights afforded to any party under California law, including without limitation " +
				"any rights under the California Consumer Legal Remedies Act (Cal. Civ. Code " +
				"§ 1750 et seq.) or the California Unfair Competition Law (Cal. Bus. & Prof. " +
				"Code § 17200 et seq.).",
		},
		{
			Heading: "Waiver of Jury Trial",
			Body: "TO THE FULLEST EXTENT PERMITTED BY APPLICABLE CALIFORNIA LAW, THE PARTIES " +
				"HEREBY IRREVOCABLY WAIVE THEIR RESPECTIVE RIGHTS TO A JURY TRIAL OF ANY " +
				"CLAIM OR CAUSE OF ACTION ARISING OUT OF OR RELATING TO THIS AGREEMENT.",
		},
	},
}

var jurisdictionUK = JurisdictionData{
	Code:             "uk",
	DisplayName:      "England and Wales",
	LegalSystemName:  "the laws of England and Wales",
	GoverningLawLine: "England and Wales",
	GoverningClause:  "governed by and construed in accordance with the laws of England and Wales",
	DisputeResolutionClause: "Any dispute, controversy, or claim arising out of or relating to this " +
		"Agreement, including any question regarding its existence, validity, or termination, " +
		"shall be referred to and finally resolved by arbitration under the London Court of " +
		"International Arbitration (LCIA) Rules, which Rules are deemed incorporated by " +
		"reference into this clause. The number of arbitrators shall be one. The seat of " +
		"arbitration shall be London. The language of the arbitration shall be English. " +
		"Nothing in this clause shall prevent a party from seeking urgent interim relief " +
		"from the English courts.",
	SeverabilityClause: "If any provision of this Agreement is held by any court or other competent " +
		"authority to be void or unenforceable in whole or in part under the laws of England " +
		"and Wales, the other provisions of this Agreement and the remainder of the affected " +
		"provision shall continue to be valid.",
	AdditionalProvisions: []JurisdictionProvision{
		{
			Heading: "Contracts (Rights of Third Parties) Act 1999",
			Body: "A person who is not a party to this Agreement has no right under the Contracts " +
				"(Rights of Third Parties) Act 1999 to enforce any term of this Agreement. This " +
				"does not affect any right or remedy of a third party which exists or is available " +
				"apart from that Act.",
		},
		{
			Heading: "Entire Agreement (UK)",
			Body: "This Agreement constitutes the entire agreement between the parties relating " +
				"to its subject matter and supersedes all prior agreements, understandings, " +
				"negotiations, and discussions, whether oral or written, between the parties. " +
				"Each party acknowledges that, in entering into this Agreement, it has not " +
				"relied on any representation or warranty not expressly set out in this Agreement.",
		},
	},
}

var jurisdictionPakistan = JurisdictionData{
	Code:             "pakistan",
	DisplayName:      "Pakistan",
	LegalSystemName:  "the laws of Pakistan",
	GoverningLawLine: "Pakistan",
	GoverningClause: "governed by and construed in accordance with the laws of Pakistan, " +
		"including the Contract Act, 1872, and all other applicable legislation of Pakistan, " +
		"excluding any conflict of laws provisions",
	DisputeResolutionClause: "Any dispute, controversy, or claim arising out of or relating to this " +
		"Agreement, or the breach, termination, or invalidity thereof, shall first be " +
		"referred to good-faith negotiation between senior representatives of the parties. " +
		"If unresolved within thirty (30) days, the dispute shall be submitted to binding " +
		"arbitration in accordance with the Arbitration Act, 1940 (as amended) of Pakistan, " +
		"with the seat of arbitration in Karachi or such other city as the parties may " +
		"mutually agree. The language of the arbitration shall be English. Judgment on the " +
		"arbitral award may be enforced in any court of competent jurisdiction in Pakistan.",
	SeverabilityClause: "If any provision of this Agreement is held by a court of competent " +
		"jurisdiction to be invalid, illegal, or unenforceable under the laws of Pakistan, " +
		"the remaining provisions shall continue in full force and effect. The parties " +
		"further agree that any such provision shall be modified to the minimum extent " +
		"necessary to render it enforceable in accordance with Pakistani law.",
	AdditionalProvisions: []JurisdictionProvision{
		{
			Heading: "Contract Act, 1872",
			Body: "The parties acknowledge that this Agreement is subject to the provisions of " +
				"the Contract Act, 1872, as applicable in Pakistan, including the requirements " +
				"of free consent, lawful consideration, and lawful object. Any provision of " +
				"this Agreement found to be contrary to the mandatory provisions of the " +
				"Contract Act, 1872 shall be void to that extent only.",
		},
		{
			Heading: "Stamp Duty",
			Body: "The parties shall be responsible for paying all applicable stamp duty on " +
				"this Agreement as required under the Stamp Act, 1899 (as applicable in the " +
				"relevant province of Pakistan). Failure to pay stamp duty shall not affect " +
				"the validity of this Agreement between the parties but may affect its " +
				"admissibility in evidence before a Pakistani court.",
		},
	},
}

package cve_v52

type Reference struct {
	Name string `json:"name"`

	// Tags is an array of one or more tags that describe the resource referenced by 'url'.
	// ref: https://cveproject.github.io/cve-schema/schema/docs/#oneOf_i0_containers_cna_problemTypes_items_descriptions_items_references_items_tags
	// broken-link: The reference link is returning a 404 error, or the site is no longer online.
	// customer-entitlement: Similar to Privileges Required, but specific to references that require non-public/paid access for customers of the particular vendor.
	// exploit: Reference contains an in-depth/detailed description of steps to exploit a vulnerability OR the reference contains any legitimate Proof of Concept (PoC) code or exploit kit.
	// government-resource: All reference links that are from a government agency or organization should be given the Government Resource tag.
	// issue-tracking: The reference is a post from a bug tracking tool such as MantisBT, Bugzilla, JIRA, Github Issues, etc...
	// mailing-list: The reference is from a mailing list -- often specific to a product or vendor.
	// mitigation: The reference contains information on steps to mitigate against the vulnerability in the event a patch can't be applied or is unavailable or for EOL product situations.
	// not-applicable: The reference link is not applicable to the vulnerability and was likely associated by MITRE accidentally (should be used sparingly).
	// patch: The reference contains an update to the software that fixes the vulnerability.
	// permissions-required: The reference link provided is blocked by a logon page. If credentials are required to see any information this tag must be applied.
	// media-coverage: The reference is from a media outlet such as a newspaper, magazine, social media, or weblog. This tag is not intended to apply to any individual's personal social media account. It is strictly intended for public media entities.
	// product: A reference appropriate for describing a product for the purpose of CPE or SWID.
	// related: A reference that is for a related (but not the same) vulnerability.
	// release-notes: The reference is in the format of a vendor or open source project's release notes or change log.
	// signature: The reference contains a method to detect or prevent the presence or exploitation of the vulnerability.
	// technical-description: The reference contains in-depth technical information about a vulnerability and its exploitation process, typically in the form of a presentation or whitepaper.
	// third-party-advisory: Advisory is from an organization that is not the vulnerable product's vendor/publisher/maintainer.
	// vendor-advisory: Advisory is from the vendor/publisher/maintainer of the product or the parent organization.
	// vdb-entry: VDBs are loosely defined as sites that provide information about this vulnerability, such as advisories, with identifiers. Included VDBs are free to access, substantially public, and have broad scope and coverage (not limited to a single vendor or research organization). See: https://www.first.org/global/sigs/vrdx/vdb-catalog
	Tags []string `json:"tags" `

	Url string `json:"url" `
}

type ReferenceTag string

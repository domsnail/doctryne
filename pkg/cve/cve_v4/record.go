package cve_v40

type Record struct {
	DataType    string `json:"data_type"`
	DataFormat  string `json:"data_format"`
	DataVersion string `json:"data_version"`

	CVEDataMeta struct {
		ASSIGNER string `json:"ASSIGNER"`
		ID       string `json:"ID"`
		STATE    string `json:"STATE"`
	} `json:"CVE_data_meta"`

	Affects struct {
		Vendor struct {
			VendorData []struct {
				Product struct {
					ProductData []struct {
						ProductName string `json:"product_name"`
						Version     struct {
							VersionData []struct {
								VersionValue string `json:"version_value"`
							} `json:"version_data"`
						} `json:"version"`
					} `json:"product_data"`
				} `json:"product"`
				VendorName string `json:"vendor_name"`
			} `json:"vendor_data"`
		} `json:"vendor"`
	} `json:"affects"`

	Description struct {
		DescriptionData []struct {
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"description_data"`
	} `json:"description"`

	ProblemType struct {
		ProblemTypeData []struct {
			Description []struct {
				Lang  string `json:"lang"`
				Value string `json:"value"`
			} `json:"description"`
		} `json:"problemtype_data"`
	} `json:"problemtype"`

	References struct {
		ReferenceData []struct {
			Name      string `json:"name"`
			Refsource string `json:"refsource"`
			Url       string `json:"url"`
		} `json:"reference_data"`
	} `json:"references"`
}

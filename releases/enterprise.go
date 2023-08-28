package releases

type EnterpriseOptions struct {
	Enterprise bool
	Meta       string // optional; may be "hsm", "fips1402", "hsm.fips1402", etc.
}

func (eo *EnterpriseOptions) requiredMetadata() string {
	metadata := ""
	if eo.Enterprise {
		metadata += "ent"
	}
	if eo.Meta != "" {
		metadata += "." + eo.Meta
	}
	return metadata
}

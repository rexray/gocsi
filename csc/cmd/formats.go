package cmd

// volumeInfoFormat is the default Go template format for emitting a
// csi.VolumeInfo
const volumeInfoFormat = `{{printf "%q\t%d" .Id .CapacityBytes}}` +
	`{{if .Attributes}}{{"\t"}}` +
	`{{range $k, $v := .Attributes}}{{printf "%q=%q\t" $k $v}}{{end}}` +
	`{{end}}{{"\n"}}`

// listVolumesFormat is the default Go template format for emitting a
// ListVolumesResponse
const listVolumesFormat = `{{range $k, $v := .Entries}}` +
	`{{with $v.VolumeInfo}}` + volumeInfoFormat + `{{end}}` +
	`{{end}}` + // {{range $v .Entries}}
	`{{if .NextToken}}{{printf "token=%q\n" .NextToken}}{{end}}`

// supportedVersionsFormat is the default Go template for emitting a
// csi.GetSupportedVersionsResponse
const supportedVersionsFormat = `{{range $v := .SupportedVersions}}` +
	`{{printf "%d.%d.%d\n" $v.Major $v.Minor $v.Patch}}{{end}}`

// pluginInfoFormat is the default Go template for emitting a
// csi.GetPluginInfoResponse
const pluginInfoFormat = `{{printf "%q\t%q" .Name .VendorVersion}}` +
	`{{range $k, $v := .Manifest}}{{printf "\t%q=%q" $k $v}}{{end}}` +
	`{{"\n"}}`

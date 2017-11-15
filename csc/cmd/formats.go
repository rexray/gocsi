package cmd

// volumeInfoFormat is the default Go template format for emitting a
// csi.VolumeInfo
const volumeInfoFormat = `{{printf "%q" .Id}}` +
	`{{if .Attributes}}{{"\t"}}` +
	`{{range $k, $v := .Attributes}}{{printf "%q=%q\t" $k $v}}{{end}}` +
	`{{end}}{{"\n"}}`

// listVolumesFormat is the default Go template format for emitting a
// ListVolumesResponse
const listVolumesFormat = `{{range $k, $v := .Entries}}` +
	`{{with $v.VolumeInfo}}` + volumeInfoFormat + `{{end}}` +
	`{{end}}` + // {{range $v .Entries}}
	`{{if .NextToken}}{{printf "token=%q\n" .NextToken}}{{end}}`

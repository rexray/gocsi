package cmd

// volumeInfoFormat is the default Go template format for emitting a
// csi.VolumeInfo
const volumeInfoFormat = `{{printf "%q\t%d" .Id .CapacityBytes}}` +
	`{{if .Attributes}}{{"\t"}}` +
	`{{range $k, $v := .Attributes}}{{printf "%q=%q\t" $k $v}}{{end}}` +
	`{{end}}{{"\n"}}`

// volumeInfoFormat is the default Go template format for emitting a
// csi.SnapshotInfo
const snapshotInfoFormat = `{{printf "%q\t%d\t%s\t%d\t%s\n" ` +
	`.Id .SizeBytes .SourceVolumeId .CreatedAt .Status}}`

// listVolumesFormat is the default Go template format for emitting a
// ListVolumesResponse
const listVolumesFormat = `{{range $k, $v := .Entries}}` +
	`{{with $v.Volume}}` + volumeInfoFormat + `{{end}}` +
	`{{end}}` + // {{range $v .Entries}}
	`{{if .NextToken}}{{printf "token=%q\n" .NextToken}}{{end}}`

// listSnapshotsFormat is the default Go template format for emitting a
// ListSnapshotsResponse
const listSnapshotsFormat = `{{range $k, $s := .Entries}}` +
	`{{with $s.Snapshot}}` + snapshotInfoFormat + `{{end}}` +
	`{{end}}` + // {{range $s .Entries}}
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

// pluginCapsFormat is the default Go template for emitting a
// csi.GetPluginCapabilities
const pluginCapsFormat = `{{range $v := .Capabilities}}` +
	`{{with $t := .Type}}` +
	`{{if isa $t "*csi.PluginCapability_Service_"}}{{if $t.Service}}` +
	`{{printf "%s\n" $t.Service.Type}}` +
	`{{end}}{{end}}` +
	`{{end}}` +
	`{{end}}`

// probeFormat is the default Go template for emitting a
// csi.Probe
const probeFormat = `{{printf "%t\n" .Ready.Value}}`

// statsFormat is the default Go template for emitting a
// csi.NodeGetVolumeStats
const statsFormat = `{{printf "%s\t%s\t" .Name .Path}}` +
	`{{range .Resp.Usage}}` +
	`{{printf "%d\t%d\t%d\t%s\n" .Available .Total .Used .Unit}}` +
	`{{end}}`

const nodeInfoFormat = `{{printf "%s\t%d\t%#v\n" .NodeId .MaxVolumesPerNode .AccessibleTopology}}`

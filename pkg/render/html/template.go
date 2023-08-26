package html

import (
	_ "embed"

	"go.xrstf.de/kubernetes-apis/pkg/types"
)

//go:embed index.html
var Index string

type IndexData struct {
	Database *types.APIOverview
	Releases []string
}

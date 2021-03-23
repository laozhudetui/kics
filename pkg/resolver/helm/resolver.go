package helm

import (
	"path/filepath"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

// Resolver is an instance of the helm resolver
type Resolver struct {
}

// splitManifest keeps the information of the manifest splitted by source
type splitManifest struct {
	path     string
	content  []byte
	original []byte
	splitID  string
}

// Resolve will render the passed helm chart and return its content ready for parsing
func (r *Resolver) Resolve(filePath string) (model.ResolvedFiles, error) {
	var rfiles = model.ResolvedFiles{}
	splits, err := renderHelm(filePath)
	if err != nil { // return error to be logged
		return model.ResolvedFiles{}, errors.New("failed to render helm chart")
	}
	for _, split := range *splits {
		origpath := filepath.Join(filepath.Dir(filePath), split.path)
		rfiles.File = append(rfiles.File, model.ResolvedFile{
			FileName:     origpath,
			Content:      split.content,
			OriginalData: split.original,
			SplitID:      split.splitID,
		})
	}
	return rfiles, nil
}

// SupportedTypes returns the supported fileKinds for this resolver
func (r *Resolver) SupportedTypes() []model.FileKind {
	return []model.FileKind{model.KINDHELM}
}

// renderHelm will use helm library to render helm charts
func renderHelm(path string) (*[]splitManifest, error) {
	client := newClient()
	manifest, err := runInstall([]string{path}, client, &values.Options{})
	if err != nil {
		return nil, err
	}
	return splitManifestYAML(manifest)
}

// splitManifestYAML will split the rendered file and return its content by template as well as the template path
func splitManifestYAML(template *release.Release) (*[]splitManifest, error) {
	sources := make([]*chart.File, 0)
	sources = updateName(sources, template.Chart, template.Chart.Name())
	splitedManifest := []splitManifest{}
	splitedSource := strings.Split(template.Manifest, "---") // split manifest by '---'
	origData := toMap(sources)
	for _, splited := range splitedSource {
		var lineID string
		for _, line := range strings.Split(splited, "\n") {
			if strings.Contains(line, "# KICS_HELM_ID_") {
				lineID = line // get auxiliary line id
				break
			}
		}
		path := strings.Split(strings.TrimLeft(splited, "\n# Source:"), "\n") // get source of splitted yaml
		// ignore auxiliary files used to render chart
		if path[0] == "" || strings.Contains(path[0], "secrets.yaml") || strings.Contains(path[0], "secret.yaml") {
			continue
		}
		if origData[path[0]] == nil {
			continue
		}
		splitedManifest = append(splitedManifest, splitManifest{
			path:     path[0],
			content:  []byte(splited),
			original: origData[path[0]], // get original data from template
			splitID:  lineID,
		})
	}
	return &splitedManifest, nil
}

// toMap will convert to map original data having the path as it's key
func toMap(files []*chart.File) map[string][]byte {
	mapFiles := make(map[string][]byte)
	for _, file := range files {
		mapFiles[file.Name] = file.Data
	}
	return mapFiles
}

// updateName will update the templates name as well as its dependecies
func updateName(template []*chart.File, charts *chart.Chart, name string) []*chart.File {
	if name != charts.Name() {
		name = filepath.Join(name, charts.Name())
	}
	for _, temp := range charts.Templates {
		temp.Name = filepath.Join(name, temp.Name)
	}
	template = append(template, charts.Templates...)
	for _, dep := range charts.Dependencies() {
		template = updateName(template, dep, filepath.Join(name, "charts"))
	}
	return template
}

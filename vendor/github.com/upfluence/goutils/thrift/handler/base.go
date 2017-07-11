package handler

import (
	"strconv"
	"strings"

	"github.com/upfluence/base"
	"github.com/upfluence/base/base_service"
	"github.com/upfluence/base/version"
)

const (
	BASE_EXCHANGE_NAME = "monitoring-exchange"

	GitCommit = ""
	GitBranch = ""
	GitRemote = ""

	Version = "v0.0.0"
)

type Interface interface {
	Version() *version.Version
	Name() string
}

type Base struct {
	UnitName   string
	SpawnDate  int64
	Interfaces []Interface
}

func buildSemanticVersion() *version.SemanticVersion {
	splittedVersion := strings.Split(Version, ".")

	if len(splittedVersion) != 3 {
		return nil
	}

	major, _ := strconv.Atoi(splittedVersion[0][1:])
	minor, _ := strconv.Atoi(splittedVersion[1])
	patch, _ := strconv.Atoi(splittedVersion[2])

	return &version.SemanticVersion{int16(major), int16(minor), int16(patch)}
}

func (h *Base) GetName() (string, error) {
	return h.UnitName, nil
}

func (h *Base) GetVersion() (*version.Version, error) {
	var gitVersion *version.GitVersion

	if GitCommit != "" {
		gitVersion = &version.GitVersion{GitCommit, GitRemote, GitBranch}
	}

	return &version.Version{buildSemanticVersion(), gitVersion}, nil
}

func (h *Base) GetInterfaceVersions() (map[string]*version.Version, error) {
	versions := make(map[string]*version.Version)

	for _, i := range h.Interfaces {
		versions[i.Name()] = i.Version()
	}

	if _, ok := versions["base"]; !ok {
		versions["base"] = (&base.Base{}).Version()
	}

	return versions, nil
}

func (h *Base) GetStatus() (base_service.Status, error) {
	return base_service.Status_ALIVE, nil
}

func (h *Base) AliveSince() (int64, error) {
	return h.SpawnDate, nil
}

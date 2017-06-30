package check

import (
	"fmt"
	"encoding/json"
	"time"
	"strings"
	"github.com/mattn/go-shellwords"
	"github.com/mitchellh/hashstructure"

	"github.com/hootsuite/sens8/util"
)

var (
	Defaults = make(map[string]interface{})
	argParser *shellwords.Parser
)

type CheckConfig struct {
	Name        string                 `json:"name"`
	Command     string                 `json:"command"`
	Interval    int                    `json:"interval"`
	Handler     *string                `json:"handler,omitempty"`
	Handlers    *[]string              `json:"handlers,omitempty"`
	Source      *string                `json:"source,omitempty"`
	Deregister  *bool                  `json:"deregister,omitempty"`
	Id          string                 `json:"-"`
	Hash        uint64                 `json:"-"`
	Argv        []string               `json:"-"`
	ExtraFields map[string]interface{} `json:"-"`
}

// ParseCheckConfigs unmarshals config data (an array of checks) and
// initializes each configs as a check. It collects errors and returns
// any valid checks as we don't want a bad config in the config array
// invalidate the whole array.
func ParseCheckConfigs(jsonData string, checkSource string, resourceType string) ([]Check, []error) {
	errors := []error{}
	checks := []Check{}

	if (jsonData == "") {
		return checks, errors
	}

	// parse json into generic map
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &parsed)
	if err != nil {
		return checks, append(errors, err)
	}

	// create checks
	for _, item := range parsed {
		check, err := parseCheckConfig(item, checkSource, resourceType)
		if err != nil {
			errors = append(errors, err)
		} else {
			checks = append(checks, check)
		}
	}

	return checks, errors
}

// parseCheckConfig parses generic structure into `CheckConfig` then create `Check`
func parseCheckConfig(item map[string]interface{}, checkSource string, resourceType string) (Check, error) {
	config := CheckConfig{}
	var check Check
	for k, v := range Defaults {
		if _, exists := item[k]; !exists {
			item[k] = v
		}
	}

	// decode and collect extra fields
	extras, err := util.DecodeWithExtraFields(item, &config)
	if err != nil {
		return check, fmt.Errorf("error parsing check %s: %s", config.Name, err.Error())
	}
	config.ExtraFields = extras
	config.Source = &checkSource

	// validate required fields
	config.Name = strings.TrimSpace(config.Name)
	if config.Name == "" {
		return check, fmt.Errorf("check name must not be empty")
	}
	if config.Command == "" {
		return check, fmt.Errorf("check command must not be empty")
	}
	if config.Interval <= 0 {
		return check, fmt.Errorf("check interval must be non-empty and > 0")
	}

	// parse command -> argv
	argv, err := argParser.Parse(config.Command)
	if err != nil {
		return check, fmt.Errorf("error parsing command %s: %s", config.Name, err.Error())
	}
	if argv == nil || len(argv) == 0 {
		return check, fmt.Errorf("error parsing command %s: must not be empty", config.Name)
	}
	config.Argv = argv
	config.Id = argv[0]

	// compute hash
	// keep last as we want to hash the fully processed config
	hash, err := hashstructure.Hash(config, nil)
	if err != nil {
		return check, fmt.Errorf("error computing check hash %s: %s", config.Name, err.Error())
	}
	config.Hash = hash

	check, err = NewCheck(config, resourceType)
	if err != nil {
		return check, fmt.Errorf("error creating check %s: %s", config.Name, err.Error())
	}

	return check, nil
}

type CheckStatus int

const (
	OK CheckStatus = iota
	WARN
	CRITICAL
)

type CheckResponse struct {
	Client string      `json:"client"`
	Check  interface{} `json:"check"` // we have arbitrary fields, so can't type this
}

type CheckResult struct {
	CheckConfig
	Status   CheckStatus `json:"status"`
	Output   string      `json:"output"`
	Duration float64     `json:"duration,omitempty"`
	Issued   int64       `json:"issued,omitempty"`
	Executed int64       `json:"executed,omitempty"`
}

// NewCheckResultFromConfig creates a new check result from  config data and populates timestamps
func NewCheckResultFromConfig(conf CheckConfig) CheckResult {
	t := time.Now().Unix()
	return CheckResult{
		CheckConfig: conf,
		Executed: t,
		Issued: t,
	}
}

// JsonResponse wraps the result in a response and marshals it into json
// including all extra check config fields
func (c *CheckResult) JsonResponse(client string) ([]byte, error) {
	merged := c.ExtraFields
	err := util.JsonStructToMap(c, &merged)
	if err != nil {
		return []byte{}, err
	}

	// wrap in a response struct and marshal for real
	return json.Marshal(CheckResponse{
		Client: client,
		Check: merged,
	})
}

type Check interface {
	// GetHash returns the computed hash of the check config
	GetHash() uint64

	// GetConfig return its CheckConfig
	GetConfig() *CheckConfig

	// Update updates the check with the resource that the controller received
	Update(resource interface{})

	// Execute run the check
	Execute() (CheckResult, error)

	// Usage returns the help docs for the check
	Usage() CheckUsage
}

type BaseCheck struct {
	Config CheckConfig
}

func (c *BaseCheck) GetHash() uint64 {
	return c.Config.Hash
}

func (c *BaseCheck) GetConfig() *CheckConfig {
	return &c.Config
}

type CheckFactory func(config CheckConfig) (Check, error)

type checkFactoryItem struct {
	factory       CheckFactory
	resourceTypes []string
}

var checkFactories = make(map[string]checkFactoryItem)

// RegisterCheck registers checks in the check factory. Call this in init()
func RegisterCheck(id string, factory CheckFactory, resourceTypes []string) error {
	if factory == nil {
		panic(fmt.Sprintf("CheckFactory factory %s does not exist.", id))
	}
	_, registered := checkFactories[id]
	if registered {
		return fmt.Errorf("CheckFactory factory %s already registered. Ignoring.", id)
	}
	checkFactories[id] = checkFactoryItem{factory, resourceTypes}
	return nil
}

// NewCheck factory for checks. instantiates checks based on the config Id (first chunk of `command`)
func NewCheck(config CheckConfig, resourceType string) (Check, error) {
	id := config.Id
	item, ok := checkFactories[id]
	if !ok {
		return nil, fmt.Errorf("check (or CheckFactory) does not exist for %s", id)
	}

	// filter checks that are not compatible with the given resource type
	f := true
	for _, t := range item.resourceTypes {
		if t == resourceType {
			f = false
		}
	}
	if f {
		return nil, fmt.Errorf("%s type is not compatible with %s", config.Id, resourceType)
	}

	return item.factory(config)
}

// CheckFactoryIds gets the list of registered check ids
func CheckFactoryIds() []string {
	i := []string{}
	for k := range checkFactories {
		i = append(i, k)
	}
	return i
}

func init() {
	argParser = shellwords.NewParser()
	argParser.ParseBacktick = true
	argParser.ParseEnv = true
}

package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

// App is the interface exposed by application
type App interface {
	// Name returns name of application
	Name() string

	// DeploymentInfo returns app ready to deploy
	Deployment() Deployment
}

// Deployment is the app ready to deploy to the target
type Deployment interface {
	// Deploy deploys app to the target
	Deploy(ctx context.Context, target AppTarget, config Config) (DeploymentInfo, error)
}

// Mode is the list of applications to deploy
type Mode []App

// Deploy deploys app in environment to the target
func (m Mode) Deploy(ctx context.Context, t AppTarget, config Config, spec *Spec) error {
	for _, app := range m {
		if appSpec, exists := spec.Apps[app.Name()]; exists && appSpec.Status() == AppStatusRunning {
			continue
		}
		info, err := app.Deployment().Deploy(ctx, t, config)
		if err != nil {
			return err
		}
		appInfo := spec.Apps[app.Name()]
		appInfo.SetIP(info.IP)
		appInfo.SetPorts(info.Ports)
		appInfo.SetStatus(AppStatusRunning)
	}
	return spec.Save()
}

// DeploymentInfo contains info about deployed application
type DeploymentInfo struct {
	// IP is the IP address assigned to application
	IP net.IP

	// Ports stores the named ports exposed by the application
	Ports map[string]int
}

// Target represents target of deployment from the perspective of coreznet
type Target interface {
	// Deploy deploys environment to the target
	Deploy(ctx context.Context, mode Mode) error

	// Stop stops apps in the environment
	Stop(ctx context.Context) error

	// Remove removes apps in the environment
	Remove(ctx context.Context) error
}

// AppTarget represents target of deployment from the perspective of application
type AppTarget interface {
	// DeployBinary deploys binary to the target
	DeployBinary(ctx context.Context, app Binary) (DeploymentInfo, error)

	// DeployContainer deploys container to the target
	DeployContainer(ctx context.Context, app Container) (DeploymentInfo, error)
}

// Prerequisites specifies list of other apps which have to be healthy because app may be started.
type Prerequisites struct {
	// Timeout tells how long we should wait for prerequisite to become healthy
	Timeout time.Duration

	// Dependencies specifies a list of health checks this app depends on
	Dependencies []HealthCheckCapable
}

// AppBase contain properties common to all types of app
type AppBase struct {
	// Name of the application
	Name string

	// Info stores runtime information about the app
	Info *AppInfo

	// ArgsFunc is the function returning args passed to binary
	ArgsFunc func(ip net.IP, homeDir string) []string

	// Ports are the network ports exposed by the application
	Ports map[string]int

	// Requires is the list of health checks to be required before app can be deployed
	Requires Prerequisites

	// PreFunc is called to preprocess app
	PreFunc func(ctx context.Context) error

	// PostFunc is called after app is deployed
	PostFunc func(ctx context.Context, deployment DeploymentInfo) error
}

func (app AppBase) preprocess(ctx context.Context, config Config, target AppTarget) error {
	must.OK(os.MkdirAll(config.AppDir+"/"+app.Name, 0o700))

	if len(app.Requires.Dependencies) > 0 {
		waitCtx, waitCancel := context.WithTimeout(ctx, app.Requires.Timeout)
		defer waitCancel()
		if err := WaitUntilHealthy(waitCtx, app.Requires.Dependencies...); err != nil {
			return err
		}
	}

	if app.Info.Status() == AppStatusStopped {
		return nil
	}

	if app.PreFunc != nil {
		return app.PreFunc(ctx)
	}
	return nil
}

func (app AppBase) postprocess(ctx context.Context, info *DeploymentInfo) error {
	info.Ports = app.Ports
	if app.PostFunc != nil {
		return app.PostFunc(ctx, *info)
	}
	return nil
}

// Binary represents binary file to be deployed
type Binary struct {
	AppBase

	// BinPathFunc is the function returning path to binary file
	BinPathFunc func(targetOS string) string
}

// Deploy deploys binary to the target
func (app Binary) Deploy(ctx context.Context, target AppTarget, config Config) (DeploymentInfo, error) {
	if err := app.AppBase.preprocess(ctx, config, target); err != nil {
		return DeploymentInfo{}, err
	}

	info, err := target.DeployBinary(ctx, app)
	if err != nil {
		return DeploymentInfo{}, err
	}
	if err := app.AppBase.postprocess(ctx, &info); err != nil {
		return DeploymentInfo{}, err
	}
	return info, nil
}

// Container represents container to be deployed
type Container struct {
	AppBase

	// Image is the url of the container image
	Image string

	// Tag is the tag of the image
	Tag string
}

// Deploy deploys container to the target
func (app Container) Deploy(ctx context.Context, target AppTarget, config Config) (DeploymentInfo, error) {
	panic("not implemented")
}

// NewSpec returns new spec
func NewSpec(config Config) *Spec {
	specFile := config.HomeDir + "/spec.json"
	specRaw, err := ioutil.ReadFile(specFile)
	switch {
	case err == nil:
		spec := &Spec{
			specFile: specFile,
		}
		must.OK(json.Unmarshal(specRaw, spec))
		if spec.Target != config.Target {
			panic(fmt.Sprintf("target mismatch, spec: %s, config: %s", spec.Target, config.Target))
		}
		if spec.Env != config.EnvName {
			panic(fmt.Sprintf("env mismatch, spec: %s, config: %s", spec.Env, config.EnvName))
		}
		if spec.Mode != config.ModeName {
			panic(fmt.Sprintf("mode mismatch, spec: %s, config: %s", spec.Mode, config.ModeName))
		}
		return spec
	case errors.Is(err, os.ErrNotExist):
	default:
		panic(err)
	}

	spec := &Spec{
		specFile: specFile,
		Target:   config.Target,
		Mode:     config.ModeName,
		Env:      config.EnvName,
		Apps:     map[string]*AppInfo{},
	}
	if config.Target == "direct" {
		spec.PGID = os.Getpid()
	}
	return spec
}

// Spec describes running environment
type Spec struct {
	specFile string

	// PGID stores process group ID used to run apps - used only by direct target
	PGID int `json:"pgid,omitempty"`

	// Target is the name of target being used to run apps
	Target string `json:"target"`

	// Mode is the name of mode
	Mode string `json:"mode"`

	// Env is the name of env
	Env string `json:"env"`

	mu sync.Mutex

	// Apps is the description of running apps
	Apps map[string]*AppInfo `json:"apps"`
}

// DescribeApp adds description of running app
func (s *Spec) DescribeApp(appType string, name string) *AppInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app, exists := s.Apps[name]; exists {
		if app.data.Type != appType {
			panic(fmt.Sprintf("app type doesn't match for application existing in spec: %s, expected: %s, got: %s", name, app.data.Type, appType))
		}
		return app
	}

	appDesc := &AppInfo{
		data: appInfoData{
			Type: appType,
		},
	}
	s.Apps[name] = appDesc
	return appDesc
}

// String converts spec to json string
func (s *Spec) String() string {
	return string(must.Bytes(json.MarshalIndent(s, "", "  ")))
}

// Save saves spec into file
func (s *Spec) Save() error {
	return ioutil.WriteFile(s.specFile, []byte(s.String()), 0o600)
}

// AppStatus describes current status of an application
type AppStatus string

const (
	// AppStatusNotDeployed ,eans that app has been never deployed
	AppStatusNotDeployed AppStatus = ""

	// AppStatusRunning means that app is running
	AppStatusRunning AppStatus = "running"

	// AppStatusStopped means app was running but now is stopped
	AppStatusStopped AppStatus = "stopped"
)

type appInfoData struct {
	// appType is the type of app
	Type string `json:"type"`

	// IP is the IP reserved for this application
	IP net.IP `json:"ip,omitempty"`

	// Status indicates the status of the application
	Status AppStatus `json:"status"`

	// Ports describe network ports provided by the application
	Ports map[string]int `json:"ports,omitempty"`
}

// AppInfo describes app running in environment
type AppInfo struct {
	mu sync.RWMutex

	data appInfoData
}

// IP returns IP
func (ai *AppInfo) IP() net.IP {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	return ai.data.IP
}

// SetIP sets IP
func (ai *AppInfo) SetIP(ip net.IP) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.data.IP = ip
}

// Status returns status
func (ai *AppInfo) Status() AppStatus {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	return ai.data.Status
}

// SetStatus sets status
func (ai *AppInfo) SetStatus(status AppStatus) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.data.Status = status
}

// SetPorts sets ports
func (ai *AppInfo) SetPorts(ports map[string]int) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.data.Ports = ports
}

// MarshalJSON marshals data to JSON
func (ai *AppInfo) MarshalJSON() ([]byte, error) {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	return json.Marshal(ai.data)
}

// UnmarshalJSON unmarshals data from JSON
func (ai *AppInfo) UnmarshalJSON(data []byte) error {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	return json.Unmarshal(data, &ai.data)
}

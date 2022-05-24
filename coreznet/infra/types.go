package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"text/template"
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
	Deploy(ctx context.Context, target AppTarget, config Config) error
}

// Mode is the list of applications to deploy
type Mode []App

// Deploy deploys app in environment to the target
func (m Mode) Deploy(ctx context.Context, t AppTarget, config Config, spec *Spec) error {
	for _, app := range m {
		if appSpec, exists := spec.Apps[app.Name()]; exists && appSpec.Status == AppStatusRunning {
			continue
		}
		if err := app.Deployment().Deploy(ctx, t, config); err != nil {
			return err
		}
		spec.Apps[app.Name()].Status = AppStatusRunning
	}
	return spec.Save()
}

// DeploymentInfo contains info about deployed application
type DeploymentInfo struct {
	// IP is the IP address assigned to application
	IP net.IP
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

// TargetEnvironment stores properties of target required to prepare app for executing
type TargetEnvironment struct {
	// IP application should bind to
	IP net.IP

	// HomeDir is the path to home dir of the application
	HomeDir string
}

// AppTarget represents target of deployment from the perspective of application
type AppTarget interface {
	// Environment returns environment properties for the application deployed to target
	Environment(app AppBase) TargetEnvironment

	// DeployBinary deploys binary to the target
	DeployBinary(ctx context.Context, app Binary) (DeploymentInfo, error)

	// DeployContainer deploys container to the target
	DeployContainer(ctx context.Context, app Container) (DeploymentInfo, error)
}

// PreprocessFunc is the function called to preprocess app
type PreprocessFunc func(ctx context.Context) error

// PostprocessFunc is the function called after application is deployed
type PostprocessFunc func(ctx context.Context, deployment DeploymentInfo) error

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

	// Args are args passed to binary
	Args []string

	// Requires is the list of health checks to be required before app can be deployed
	Requires Prerequisites

	// PreFunc is called to preprocess app
	PreFunc PreprocessFunc

	// PostFunc is called after app is deployed
	PostFunc PostprocessFunc
}

func (app AppBase) preprocess(ctx context.Context, config Config, target AppTarget) error {
	must.OK(os.MkdirAll(config.AppDir+"/"+app.Name, 0o700))

	env := target.Environment(app)
	for i, arg := range app.Args {
		tpl := template.Must(template.New("").Parse(arg))
		buf := &bytes.Buffer{}
		must.OK(tpl.Execute(buf, env))
		app.Args[i] = buf.String()
	}

	if len(app.Requires.Dependencies) > 0 {
		waitCtx, waitCancel := context.WithTimeout(ctx, app.Requires.Timeout)
		defer waitCancel()
		if err := WaitUntilHealthy(waitCtx, app.Requires.Dependencies...); err != nil {
			return err
		}
	}

	if app.Info.Status == AppStatusStopped {
		return nil
	}

	if app.PreFunc != nil {
		return app.PreFunc(ctx)
	}
	return nil
}

func (app AppBase) postprocess(ctx context.Context, info DeploymentInfo) error {
	if app.PostFunc != nil {
		return app.PostFunc(ctx, info)
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
func (app Binary) Deploy(ctx context.Context, target AppTarget, config Config) error {
	if err := app.AppBase.preprocess(ctx, config, target); err != nil {
		return err
	}

	info, err := target.DeployBinary(ctx, app)
	if err != nil {
		return err
	}
	return app.AppBase.postprocess(ctx, info)
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
func (app Container) Deploy(ctx context.Context, target AppTarget, config Config) error {
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
		if app.Type != appType {
			panic(fmt.Sprintf("app type doesn't match for application existing in spec: %s, expected: %s, got: %s", name, app.Type, appType))
		}
		return app
	}

	appDesc := &AppInfo{
		Type: appType,
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

// AppInfo describes app running in environment
type AppInfo struct {
	// Type is the type of app
	Type string `json:"type"`

	// IP is the IP reserved for this application
	IP net.IP `json:"ip,omitempty"`

	// Status indicates the status of the application
	Status AppStatus `json:"status"`

	mu sync.Mutex

	// Ports describe network ports provided by the application
	Ports map[string]int `json:"ports,omitempty"`
}

// AddPort adds port to app description
func (a *AppInfo) AddPort(name string, port int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Ports == nil {
		a.Ports = map[string]int{}
	}

	a.Ports[name] = port
}

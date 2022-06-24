package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/crust/exec"
)

const binaryDockerImage = "alpine:3.16.0"

// AppType represents the type of application
type AppType string

// App is the interface exposed by application
type App interface {
	// Type returns type of application
	Type() AppType

	// Info returns deployment info
	Info() DeploymentInfo

	// Name returns name of application
	Name() string

	// Deployment returns app ready to deploy
	Deployment() Deployment
}

// Deployment is the app ready to deploy to the target
type Deployment interface {
	// Dependencies returns the list of applications which must be running before the current deployment may be started
	Dependencies() []HealthCheckCapable

	// DockerImage returns the docker image used by the deployment
	DockerImage() string

	// Deploy deploys app to the target
	Deploy(ctx context.Context, target AppTarget, config Config) (DeploymentInfo, error)
}

// Mode is the list of applications to deploy
type Mode []App

// Deploy deploys app in environment to the target
func (m Mode) Deploy(ctx context.Context, t AppTarget, config Config, spec *Spec) error {
	err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		deploymentSlots := make(chan struct{}, runtime.NumCPU())
		for i := 0; i < cap(deploymentSlots); i++ {
			deploymentSlots <- struct{}{}
		}
		imagePullSlots := make(chan struct{}, 3)
		for i := 0; i < cap(imagePullSlots); i++ {
			imagePullSlots <- struct{}{}
		}

		deployments := map[string]struct {
			Deployment   Deployment
			ImageReadyCh chan struct{}
			ReadyCh      chan struct{}
		}{}
		images := map[string]chan struct{}{}
		for _, app := range m {
			deployment := app.Deployment()
			if _, exists := images[deployment.DockerImage()]; !exists {
				ch := make(chan struct{}, 1)
				ch <- struct{}{}
				images[deployment.DockerImage()] = ch
			}
			deployments[app.Name()] = struct {
				Deployment   Deployment
				ImageReadyCh chan struct{}
				ReadyCh      chan struct{}
			}{
				Deployment:   deployment,
				ImageReadyCh: images[deployment.DockerImage()],
				ReadyCh:      make(chan struct{}),
			}
		}
		for name, toDeploy := range deployments {
			if appSpec, exists := spec.Apps[name]; exists && appSpec.Info().Status == AppStatusRunning {
				close(toDeploy.ReadyCh)
				continue
			}

			appInfo := spec.Apps[name]
			toDeploy := toDeploy
			spawn("deploy."+name, parallel.Continue, func(ctx context.Context) error {
				log := logger.Get(ctx)
				deployment := toDeploy.Deployment

				log.Info("Deployment initialized")

				if err := ensureDockerImage(ctx, deployment.DockerImage(), imagePullSlots, toDeploy.ImageReadyCh); err != nil {
					return err
				}

				if dependencies := deployment.Dependencies(); len(dependencies) > 0 {
					names := make([]string, 0, len(dependencies))
					for _, d := range dependencies {
						names = append(names, d.Name())
					}
					log.Info("Waiting for dependencies", zap.Strings("dependencies", names))
					for _, name := range names {
						select {
						case <-ctx.Done():
							return errors.WithStack(ctx.Err())
						case <-deployments[name].ReadyCh:
						}
					}
					log.Info("Dependencies are running now")
				}

				log.Info("Waiting for free slot for deploying the application")
				select {
				case <-ctx.Done():
					return errors.WithStack(ctx.Err())
				case <-deploymentSlots:
				}

				log.Info("Deployment started")

				info, err := deployment.Deploy(ctx, t, config)
				if err != nil {
					return err
				}
				appInfo.SetInfo(info)

				log.Info("Deployment succeeded")

				close(toDeploy.ReadyCh)
				deploymentSlots <- struct{}{}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	return spec.Save()
}

func ensureDockerImage(ctx context.Context, image string, slots chan struct{}, readyCh chan struct{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, ok := <-readyCh:
		// If <-imageReadyCh blocks it means another goroutine is pulling the image
		if !ok {
			// Channel is closed, image has been already pulled, we are ready to go
			return nil
		}
	}

	log := logger.Get(ctx).With(zap.String("image", image))

	imageBuf := &bytes.Buffer{}
	imageCmd := exec.Docker("images", "-q", image)
	imageCmd.Stdout = imageBuf
	if err := libexec.Exec(ctx, imageCmd); err != nil {
		return errors.Wrapf(err, "failed to list image '%s'", image)
	}
	if imageBuf.Len() > 0 {
		log.Info("Docker image exists")
		close(readyCh)
		return nil
	}

	log.Info("Waiting for free slot for pulling the docker image")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-slots:
	}

	log.Info("Pulling docker image")

	if err := libexec.Exec(ctx, exec.Docker("pull", image)); err != nil {
		return errors.Wrapf(err, "failed to pull docker image '%s'", image)
	}

	log.Info("Image pulled")
	close(readyCh)
	slots <- struct{}{}
	return nil
}

// DeploymentInfo contains info about deployed application
type DeploymentInfo struct {
	// Container stores the name of the docker container where app is running - present only for apps running in docker
	Container string `json:"container,omitempty"`

	// FromHostIP is the host's IP application binds to
	FromHostIP net.IP `json:"fromHostIP,omitempty"` // nolint:tagliatelle // it wants fromHostIp

	// FromContainerIP is the IP of the container application is running in
	FromContainerIP net.IP `json:"fromContainerIP,omitempty"` // nolint:tagliatelle // it wants fromContainerIp

	// Status indicates the status of the application
	Status AppStatus `json:"status"`

	// Ports describe network ports provided by the application
	Ports map[string]int `json:"ports,omitempty"`
}

// Target represents target of deployment from the perspective of crustznet
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

// Prerequisites specifies list of other apps which have to be healthy before app may be started.
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
	ArgsFunc func() []string

	// Ports are the network ports exposed by the application
	Ports map[string]int

	// Requires is the list of health checks to be required before app can be deployed
	Requires Prerequisites

	// PrepareFunc is the function called before application is deployed for the first time.
	// It is a good place to prepare configuration files and other things which must or might be done before application runs.
	PrepareFunc func() error

	// ConfigureFunc is the function called after application is deployed for the first time.
	// It is a good place to connect to the application to configure it because at this stage the app's IP address is known.
	ConfigureFunc func(ctx context.Context, deployment DeploymentInfo) error
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

	if app.Info.Info().Status == AppStatusStopped {
		return nil
	}

	if app.PrepareFunc != nil {
		return app.PrepareFunc()
	}
	return nil
}

func (app AppBase) postprocess(ctx context.Context, info DeploymentInfo) error {
	if app.Info.Info().Status == AppStatusStopped {
		return nil
	}
	if app.ConfigureFunc != nil {
		return app.ConfigureFunc(ctx, info)
	}
	return nil
}

// Binary represents binary file to be deployed
type Binary struct {
	AppBase

	// BinPath is the path to linux binary file
	BinPath string
}

// DockerImage returns the docker image used by the deployment
func (app Binary) DockerImage() string {
	return binaryDockerImage
}

// Dependencies returns the list of applications which must be running before the current deployment may be started
func (app Binary) Dependencies() []HealthCheckCapable {
	return app.Requires.Dependencies
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
	if err := app.AppBase.postprocess(ctx, info); err != nil {
		return DeploymentInfo{}, err
	}
	return info, nil
}

// EnvVar is used to define environment variable for docker container
type EnvVar struct {
	Name  string
	Value string
}

// Container represents container to be deployed
type Container struct {
	AppBase

	// Image is the url of the container image
	Image string

	// EnvVars define environment variables for docker container
	EnvVars []EnvVar
}

// DockerImage returns the docker image used by the deployment
func (app Container) DockerImage() string {
	return app.Image
}

// Dependencies returns the list of applications which must be running before the current deployment may be started
func (app Container) Dependencies() []HealthCheckCapable {
	return app.Requires.Dependencies
}

// Deploy deploys container to the target
func (app Container) Deploy(ctx context.Context, target AppTarget, config Config) (DeploymentInfo, error) {
	if err := app.AppBase.preprocess(ctx, config, target); err != nil {
		return DeploymentInfo{}, err
	}

	info, err := target.DeployContainer(ctx, app)
	if err != nil {
		return DeploymentInfo{}, err
	}
	if err := app.AppBase.postprocess(ctx, info); err != nil {
		return DeploymentInfo{}, err
	}
	return info, nil
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
	return spec
}

// Spec describes running environment
type Spec struct {
	specFile string

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
func (s *Spec) DescribeApp(appType AppType, name string) *AppInfo {
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
	// Type is the type of app
	Type AppType `json:"type"`

	// Info stores app deployment information
	Info DeploymentInfo `json:"info"`
}

// AppInfo describes app running in environment
type AppInfo struct {
	mu sync.RWMutex

	data appInfoData
}

// SetInfo sets fields based on deployment info
func (ai *AppInfo) SetInfo(info DeploymentInfo) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.data.Info = info
}

// Info returns deployment info
func (ai *AppInfo) Info() DeploymentInfo {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	return ai.data.Info
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

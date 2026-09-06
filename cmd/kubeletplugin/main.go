/*
 * Copyright 2023 The Kubernetes Authors.
 * Copyright 2026 Graeme Lawes.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/klog/v2"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
	"github.com/gclawes/rockchip-dra-driver/pkg/flags"
	"github.com/gclawes/rockchip-dra-driver/pkg/metrics"
)

const driverPluginCheckpointFile = "checkpoint.json"

type Flags struct {
	kubeClientConfig flags.KubeClientConfig
	loggingConfig    *flags.LoggingConfig

	nodeName                      string
	cdiRoot                       string
	kubeletRegistrarDirectoryPath string
	kubeletPluginsDirectoryPath   string
	healthcheckPort               int
	metricsPort                   int
	driverName                    string
	podUID                        string
	mockDevices                   bool
	npuEnabled                    bool
	gpuEnabled                    bool
	npuMaxAllocations             int
	gpuMaxAllocations             int
	sysfsRoot                     string
}

type Config struct {
	flags         *Flags
	coreclient    coreclientset.Interface
	cancelMainCtx func(error)
}

func (c Config) DriverPluginPath() string {
	return filepath.Join(c.flags.kubeletPluginsDirectoryPath, c.flags.driverName)
}

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	f := &Flags{
		loggingConfig: flags.NewLoggingConfig(),
	}
	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "node-name",
			Usage:       "The name of the node to be worked on.",
			Required:    true,
			Destination: &f.nodeName,
			EnvVars:     []string{"NODE_NAME"},
		},
		&cli.StringFlag{
			Name:        "cdi-root",
			Usage:       "Absolute path to the directory where CDI files will be generated.",
			Value:       "/etc/cdi",
			Destination: &f.cdiRoot,
			EnvVars:     []string{"CDI_ROOT"},
		},
		&cli.StringFlag{
			Name:        "kubelet-registrar-directory-path",
			Usage:       "Absolute path to the directory where kubelet stores plugin registrations.",
			Value:       kubeletplugin.KubeletRegistryDir,
			Destination: &f.kubeletRegistrarDirectoryPath,
			EnvVars:     []string{"KUBELET_REGISTRAR_DIRECTORY_PATH"},
		},
		&cli.StringFlag{
			Name:        "kubelet-plugins-directory-path",
			Usage:       "Absolute path to the directory where kubelet stores plugin data.",
			Value:       kubeletplugin.KubeletPluginsDir,
			Destination: &f.kubeletPluginsDirectoryPath,
			EnvVars:     []string{"KUBELET_PLUGINS_DIRECTORY_PATH"},
		},
		&cli.IntFlag{
			Name:        "healthcheck-port",
			Usage:       "Port to start a gRPC healthcheck service. Negative disables it.",
			Value:       -1,
			Destination: &f.healthcheckPort,
			EnvVars:     []string{"HEALTHCHECK_PORT"},
		},
		&cli.IntFlag{
			Name:        "metrics-port",
			Usage:       "Port to expose Prometheus metrics at /metrics. Negative disables it.",
			Value:       -1,
			Destination: &f.metricsPort,
			EnvVars:     []string{"METRICS_PORT"},
		},
		&cli.StringFlag{
			Name:        "driver-name",
			Usage:       "DRA driver name published on ResourceSlices.",
			Value:       consts.DriverName,
			Destination: &f.driverName,
			EnvVars:     []string{"DRIVER_NAME"},
		},
		&cli.StringFlag{
			Name:        "pod-uid",
			Usage:       "UID of the plugin pod (used for rolling updates).",
			Destination: &f.podUID,
			EnvVars:     []string{"POD_UID"},
		},
		&cli.BoolFlag{
			Name:        "mock-devices",
			Usage:       "Advertise a synthetic RK3588 NPU and GPU instead of reading sysfs. For kind e2e.",
			Destination: &f.mockDevices,
			EnvVars:     []string{"MOCK_DEVICES"},
		},
		&cli.BoolFlag{
			Name:        "npu-enabled",
			Usage:       "Discover and publish NPU devices.",
			Value:       true,
			Destination: &f.npuEnabled,
			EnvVars:     []string{"NPU_ENABLED"},
		},
		&cli.BoolFlag{
			Name:        "gpu-enabled",
			Usage:       "Discover and publish GPU devices.",
			Value:       true,
			Destination: &f.gpuEnabled,
			EnvVars:     []string{"GPU_ENABLED"},
		},
		&cli.IntFlag{
			Name:        "npu-max-allocations",
			Usage:       "Maximum concurrent NPU allocations. 0 uses a SoC-aware default.",
			Value:       0,
			Destination: &f.npuMaxAllocations,
			EnvVars:     []string{"NPU_MAX_ALLOCATIONS"},
		},
		&cli.IntFlag{
			Name:        "gpu-max-allocations",
			Usage:       "Maximum concurrent GPU allocations. 0 uses the default of 8.",
			Value:       0,
			Destination: &f.gpuMaxAllocations,
			EnvVars:     []string{"GPU_MAX_ALLOCATIONS"},
		},
		&cli.StringFlag{
			Name:        "sysfs-root",
			Usage:       "Sysfs root used for discovery. Tests pass a fixture tree.",
			Value:       "/sys",
			Destination: &f.sysfsRoot,
			EnvVars:     []string{"SYSFS_ROOT"},
		},
	}
	cliFlags = append(cliFlags, f.kubeClientConfig.Flags()...)
	cliFlags = append(cliFlags, f.loggingConfig.Flags()...)

	return &cli.App{
		Name:            "kubeletplugin",
		Usage:           "Rockchip DRA kubelet plugin.",
		ArgsUsage:       " ",
		HideHelpCommand: true,
		Flags:           cliFlags,
		Before: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				return fmt.Errorf("arguments not supported: %v", c.Args().Slice())
			}
			return f.loggingConfig.Apply()
		},
		Action: func(c *cli.Context) error {
			clientSets, err := f.kubeClientConfig.NewClientSets()
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}
			config := &Config{
				flags:      f,
				coreclient: clientSets.Core,
			}
			return RunPlugin(c.Context, config)
		},
	}
}

func RunPlugin(ctx context.Context, config *Config) error {
	logger := klog.FromContext(ctx)

	if err := os.MkdirAll(config.DriverPluginPath(), 0o750); err != nil {
		return err
	}

	info, err := os.Stat(config.flags.cdiRoot)
	switch {
	case err != nil && os.IsNotExist(err):
		if err := os.MkdirAll(config.flags.cdiRoot, 0o750); err != nil {
			return err
		}
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("path for cdi file generation is not a directory: %s", config.flags.cdiRoot)
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	ctx, cancel := context.WithCancelCause(ctx)
	config.cancelMainCtx = cancel

	metricsServer, err := metrics.StartServer(ctx, config.flags.metricsPort)
	if err != nil {
		return fmt.Errorf("start metrics server: %w", err)
	}
	defer func(ctx context.Context) {
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
		defer shutdownCancel()
		if err := metricsServer.Stop(shutdownCtx); err != nil {
			logger.Error(err, "failed to stop metrics server")
		}
	}(context.WithoutCancel(ctx))

	driver, err := NewDriver(ctx, config)
	if err != nil {
		return err
	}

	<-ctx.Done()
	stop()
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(err, "error from context")
	}

	if err := driver.Shutdown(logger); err != nil {
		logger.Error(err, "Unable to cleanly shutdown driver")
	}
	return nil
}

func (c *Config) discoveryConfig() discovery.Config {
	return discovery.Config{
		SysfsRoot:         c.flags.sysfsRoot,
		Mock:              c.flags.mockDevices,
		NPUEnabled:        c.flags.npuEnabled,
		GPUEnabled:        c.flags.gpuEnabled,
		NPUMaxAllocations: c.flags.npuMaxAllocations,
		GPUMaxAllocations: c.flags.gpuMaxAllocations,
	}
}

/*
 * Copyright The Kubernetes Authors.
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
	"time"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/klog/v2"

	"github.com/gclawes/rockchip-dra-driver/pkg/metrics"
)

type driver struct {
	client            coreclientset.Interface
	helper            *kubeletplugin.Helper
	state             *DeviceState
	cancelCtx         func(error)
	discoveryInterval time.Duration
	reconcileNow      chan struct{}
	healthChanged     chan struct{}
	rediscoveryDone   chan struct{}
}

func NewDriver(ctx context.Context, config *Config) (*driver, error) {
	interval := config.flags.discoveryInterval
	if interval <= 0 {
		interval = defaultDiscoveryInterval
	}
	d := &driver{
		client:            config.coreclient,
		cancelCtx:         config.cancelMainCtx,
		discoveryInterval: interval,
		reconcileNow:      make(chan struct{}, 1),
		healthChanged:     make(chan struct{}, 1),
	}

	state, err := NewDeviceState(ctx, config)
	if err != nil {
		return nil, err
	}
	d.state = state

	helper, err := kubeletplugin.Start(ctx, d,
		kubeletplugin.KubeClient(config.coreclient),
		kubeletplugin.NodeName(config.flags.nodeName),
		kubeletplugin.DriverName(config.flags.driverName),
		kubeletplugin.RegistrarDirectoryPath(config.flags.kubeletRegistrarDirectoryPath),
		kubeletplugin.PluginDataDirectoryPath(config.DriverPluginPath()),
		kubeletplugin.RollingUpdate(types.UID(config.flags.podUID)),
	)
	if err != nil {
		return nil, err
	}
	d.helper = helper

	if err := helper.PublishResources(ctx, state.Resources()); err != nil {
		return nil, err
	}

	d.rediscoveryDone = make(chan struct{})
	go d.runRediscovery(ctx)

	return d, nil
}

// runRediscovery polls sysfs and republishes the ResourceSlice when the
// device set or a device's health changes. inotify on sysfs is not reliable
// for driver bind and unbind, so this is a poll.
func (d *driver) runRediscovery(ctx context.Context) {
	defer close(d.rediscoveryDone)
	logger := klog.FromContext(ctx)
	if d.discoveryInterval >= healthLease {
		logger.Info("Discovery interval is at least the kubelet health lease; reported health can stay stale", "interval", d.discoveryInterval, "lease", healthLease)
	}
	ticker := time.NewTicker(d.discoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-d.reconcileNow:
		}
		if err := d.reconcileAndPublish(ctx); err != nil {
			logger.Error(err, "Rediscovery failed")
		}
	}
}

func (d *driver) reconcileAndPublish(ctx context.Context) error {
	result, err := d.state.Reconcile(ctx)
	if err != nil {
		return err
	}
	if result.HealthChanged {
		d.notifyHealth()
	}
	if !result.ResourcesChanged {
		return nil
	}
	klog.FromContext(ctx).Info("Publishing updated devices")
	return d.helper.PublishResources(ctx, result.Resources)
}

func (d *driver) kickReconcile() {
	if d == nil || d.reconcileNow == nil {
		return
	}
	select {
	case d.reconcileNow <- struct{}{}:
	default:
	}
}

func (d *driver) notifyHealth() {
	if d == nil || d.healthChanged == nil {
		return
	}
	select {
	case d.healthChanged <- struct{}{}:
	default:
	}
}

func (d *driver) Shutdown(logger klog.Logger) error {
	if d.rediscoveryDone != nil {
		<-d.rediscoveryDone
	}
	d.helper.Stop()
	return nil
}

func (d *driver) PrepareResourceClaims(ctx context.Context, claims []*resourceapi.ResourceClaim) (map[types.UID]kubeletplugin.PrepareResult, error) {
	logger := klog.FromContext(ctx)
	logger.Info("PrepareResourceClaims is called", "numClaims", len(claims))
	result := make(map[types.UID]kubeletplugin.PrepareResult)
	for _, claim := range claims {
		result[claim.UID] = d.prepareResourceClaim(ctx, claim)
	}
	return result, nil
}

func (d *driver) prepareResourceClaim(ctx context.Context, claim *resourceapi.ResourceClaim) (result kubeletplugin.PrepareResult) {
	logger := klog.FromContext(ctx)
	start := time.Now()
	defer func() {
		metrics.ObservePrepareClaim(result.Err, time.Since(start))
	}()

	preparedDevices, err := d.state.Prepare(ctx, claim)
	if err != nil {
		logger.Error(err, "Error preparing devices for claim", "uid", claim.UID)
		return kubeletplugin.PrepareResult{Err: fmt.Errorf("error preparing devices for claim %v: %w", claim.UID, err)}
	}

	var prepared []kubeletplugin.Device
	for _, pd := range preparedDevices {
		prepared = append(prepared, kubeletplugin.Device{
			Requests:     pd.RequestNames,
			PoolName:     pd.PoolName,
			DeviceName:   pd.DeviceName,
			CDIDeviceIDs: pd.CDIDeviceIDs,
			ShareID:      pd.ShareID,
		})
	}
	return kubeletplugin.PrepareResult{Devices: prepared}
}

func (d *driver) UnprepareResourceClaims(ctx context.Context, claims []kubeletplugin.NamespacedObject) (map[types.UID]error, error) {
	logger := klog.FromContext(ctx)
	logger.Info("UnprepareResourceClaims is called", "numClaims", len(claims))
	result := make(map[types.UID]error)
	for _, claim := range claims {
		result[claim.UID] = d.unprepareResourceClaim(ctx, claim)
	}
	return result, nil
}

func (d *driver) unprepareResourceClaim(_ context.Context, claim kubeletplugin.NamespacedObject) (err error) {
	start := time.Now()
	defer func() {
		metrics.ObserveUnprepareClaim(err, time.Since(start))
	}()
	if err = d.state.Unprepare(claim.UID); err != nil {
		return fmt.Errorf("error unpreparing devices for claim %v: %w", claim.UID, err)
	}
	// Drop a retained device as soon as its last claim is gone, instead of
	// waiting for the next poll.
	d.kickReconcile()
	return nil
}

// WatchHealthStatus sends the health of every published device, then again
// whenever rediscovery changes it and at least once per healthResendInterval.
// The kubelet treats a device as unknown if its report is not refreshed
// within HealthCheckTimeout.
func (d *driver) WatchHealthStatus(ctx context.Context, reports chan<- kubeletplugin.DeviceHealthReport) error {
	resend := time.NewTicker(healthResendInterval)
	defer resend.Stop()
	for {
		report := d.state.healthReport()
		select {
		case <-ctx.Done():
			return nil
		case reports <- report:
		}
		select {
		case <-ctx.Done():
			return nil
		case <-d.healthChanged:
		case <-resend.C:
		}
	}
}

func (d *driver) HandleError(ctx context.Context, err error, msg string) {
	utilruntime.HandleErrorWithContext(ctx, err, msg)
	if !errors.Is(err, kubeletplugin.ErrRecoverable) {
		metrics.FatalBackgroundErrorsTotal.Inc()
		if d.cancelCtx != nil {
			d.cancelCtx(fmt.Errorf("fatal background error: %w", err))
		}
	}
}

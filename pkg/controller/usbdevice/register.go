package usbdevice

import (
	"context"

	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/sirupsen/logrus"

	"github.com/harvester/pcidevices/pkg/config"
)

const (
	KubeVirtNamespace      = "harvester-system"
	KubeVirtResource       = "kubevirt"
	KubeVirtResourcePrefix = "kubevirt.io/"
)

func Register(ctx context.Context, management *config.FactoryManager) error {
	usbDeviceCtrl := management.DeviceFactory.Devices().V1beta1().USBDevice()
	usbDeviceClaimCtrl := management.DeviceFactory.Devices().V1beta1().USBDeviceClaim()
	virtClient := management.KubevirtFactory.Kubevirt().V1().KubeVirt()

	handler := NewHandler(usbDeviceCtrl, usbDeviceClaimCtrl, usbDeviceCtrl.Cache(), usbDeviceClaimCtrl.Cache())
	usbDeviceClaimController := NewClaimHandler(usbDeviceCtrl.Cache(), usbDeviceClaimCtrl, usbDeviceCtrl, virtClient, handler.reconcileSignal)

	// Initial reconcile
	if err := handler.reconcile(); err != nil {
		logrus.Errorf("error initial reconcile usb devices: %v", err)
		return err
	}

	if err := handler.WatchUSBDevices(ctx); err != nil {
		logrus.Errorf("error watching usb devices: %v", err)
		return err
	}

	usbDeviceClaimCtrl.OnChange(ctx, "usbClaimClient-device-claim", usbDeviceClaimController.OnUSBDeviceClaimChanged)
	usbDeviceClaimCtrl.OnRemove(ctx, "usbClaimClient-device-claim-remove", usbDeviceClaimController.OnRemove)
	relatedresource.WatchClusterScoped(ctx, "USBDeviceToClaimReconcile", handler.OnDeviceChange, usbDeviceClaimCtrl, usbDeviceCtrl)

	return nil
}

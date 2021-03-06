/*
 * Copyright 2020-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//Package core provides the utility for olt devices, flows and statistics
package core

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opencord/voltha-lib-go/v3/pkg/adapters/adapterif"
	"github.com/opencord/voltha-lib-go/v3/pkg/flows"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-lib-go/v3/pkg/pmmetrics"
	"github.com/opencord/voltha-openolt-adapter/internal/pkg/core/l2oam"
	"github.com/opencord/voltha-openolt-adapter/internal/pkg/olterrors"
	rsrcMgr "github.com/opencord/voltha-openolt-adapter/internal/pkg/resourcemanager"
	"github.com/opencord/voltha-openolt-adapter/pkg/mocks"
	"github.com/opencord/voltha-protos/v3/go/common"
	ic "github.com/opencord/voltha-protos/v3/go/inter_container"
	of "github.com/opencord/voltha-protos/v3/go/openflow_13"
	oop "github.com/opencord/voltha-protos/v3/go/openolt"
	"github.com/opencord/voltha-protos/v3/go/voltha"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Constants for number of retries and for timeout
const (
	MaxRetry       = 10
	MaxTimeOutInMs = 500
	InvalidPort    = 0xffffffff
)

// pendingFlowRemoveDataKey is key to pendingFlowRemoveDataPerSubscriber map
type pendingFlowRemoveDataKey struct {
	intfID uint32
	onuID  uint32
	uniID  uint32
}

// pendingFlowRemoveData is value stored in pendingFlowRemoveDataPerSubscriber map
// This holds the number of pending flow removes and also a signal channel to
// to indicate the receiver when all flow removes are handled
type pendingFlowRemoveData struct {
	pendingFlowRemoveCount uint32
	allFlowsRemoved        chan struct{}
}

//DeviceHandler will interact with the OLT device.
type DeviceHandler struct {
	device        *voltha.Device
	coreProxy     adapterif.CoreProxy
	AdapterProxy  adapterif.AdapterProxy
	EventProxy    adapterif.EventProxy
	openOLT       *OpenOLT
	exitChannel   chan int
	lockDevice    sync.RWMutex
	Client        oop.OpenoltClient
	transitionMap *TransitionMap
	clientCon     *grpc.ClientConn
	flowMgr       *OpenOltFlowMgr
	eventMgr      *OpenOltEventMgr
	resourceMgr   *rsrcMgr.OpenOltResourceMgr

	discOnus                      sync.Map
	onus                          sync.Map
	portStats                     *OpenOltStatisticsMgr
	metrics                       *pmmetrics.PmMetrics
	stopCollector                 chan bool
	stopHeartbeatCheck            chan bool
	activePorts                   sync.Map
	stopIndications               chan bool
	isReadIndicationRoutineActive bool

	pendingFlowRemoveDataPerSubscriber map[pendingFlowRemoveDataKey]pendingFlowRemoveData
}

//OnuDevice represents ONU related info
type OnuDevice struct {
	deviceID      string
	deviceType    string
	serialNumber  string
	onuID         uint32
	intfID        uint32
	proxyDeviceID string
	losRaised     bool
	rdiRaised     bool
}

var pmNames = []string{
	"rx_bytes",
	"rx_packets",
	"rx_mcast_packets",
	"rx_bcast_packets",
	"tx_bytes",
	"tx_packets",
	"tx_mcast_packets",
	"tx_bcast_packets",
}

// set true if you want to disable BBSim function
var disableBBSim bool = true

//NewOnuDevice creates a new Onu Device
func NewOnuDevice(devID, deviceTp, serialNum string, onuID, intfID uint32, proxyDevID string, losRaised bool) *OnuDevice {
	var device OnuDevice
	device.deviceID = devID
	device.deviceType = deviceTp
	device.serialNumber = serialNum
	device.onuID = onuID
	device.intfID = intfID
	device.proxyDeviceID = proxyDevID
	device.losRaised = losRaised
	return &device
}

//NewDeviceHandler creates a new device handler
func NewDeviceHandler(cp adapterif.CoreProxy, ap adapterif.AdapterProxy, ep adapterif.EventProxy, device *voltha.Device, adapter *OpenOLT) *DeviceHandler {
	var dh DeviceHandler
	dh.coreProxy = cp
	dh.AdapterProxy = ap
	dh.EventProxy = ep
	cloned := (proto.Clone(device)).(*voltha.Device)
	dh.device = cloned
	dh.openOLT = adapter
	dh.exitChannel = make(chan int, 1)
	dh.lockDevice = sync.RWMutex{}
	dh.stopCollector = make(chan bool, 2)
	dh.stopHeartbeatCheck = make(chan bool, 2)
	dh.metrics = pmmetrics.NewPmMetrics(cloned.Id, pmmetrics.Frequency(150), pmmetrics.FrequencyOverride(false), pmmetrics.Grouped(false), pmmetrics.Metrics(pmNames))
	dh.activePorts = sync.Map{}
	dh.stopIndications = make(chan bool, 1)
	dh.pendingFlowRemoveDataPerSubscriber = make(map[pendingFlowRemoveDataKey]pendingFlowRemoveData)

	if dh.openOLT.config != nil {
		NewL2oamHandle(context.Background(), dh.openOLT.config.InterfaceName, dh.openOLT.config.SrcMac)
	}

	return &dh
}

// start save the device to the data model
func (dh *DeviceHandler) start(ctx context.Context) {
	dh.lockDevice.Lock()
	defer dh.lockDevice.Unlock()
	logger.Debugw(ctx, "starting-device-agent", log.Fields{"device": dh.device})
	logger.Debug(ctx, "device-agent-started")
}

// stop stops the device dh.  Not much to do for now
func (dh *DeviceHandler) stop(ctx context.Context) {
	dh.lockDevice.Lock()
	defer dh.lockDevice.Unlock()
	logger.Debug(ctx, "stopping-device-agent")
	dh.exitChannel <- 1
	logger.Debug(ctx, "device-agent-stopped")
}

func macifyIP(ip net.IP) string {
	if len(ip) > 0 {
		oct1 := strconv.FormatInt(int64(ip[12]), 16)
		oct2 := strconv.FormatInt(int64(ip[13]), 16)
		oct3 := strconv.FormatInt(int64(ip[14]), 16)
		oct4 := strconv.FormatInt(int64(ip[15]), 16)
		return fmt.Sprintf("00:00:%02v:%02v:%02v:%02v", oct1, oct2, oct3, oct4)
	}
	return ""
}

func generateMacFromHost(ctx context.Context, host string) (string, error) {
	var genmac string
	var addr net.IP
	var ips []string
	var err error

	logger.Debugw(ctx, "generating-mac-from-host", log.Fields{"host": host})

	if addr = net.ParseIP(host); addr == nil {
		logger.Debugw(ctx, "looking-up-hostname", log.Fields{"host": host})

		if ips, err = net.LookupHost(host); err == nil {
			logger.Debugw(ctx, "dns-result-ips", log.Fields{"ips": ips})
			if addr = net.ParseIP(ips[0]); addr == nil {
				return "", olterrors.NewErrInvalidValue(log.Fields{"ip": ips[0]}, nil)
			}
			genmac = macifyIP(addr)
			logger.Debugw(ctx, "using-ip-as-mac",
				log.Fields{"host": ips[0],
					"mac": genmac})
			return genmac, nil
		}
		return "", olterrors.NewErrAdapter("cannot-resolve-hostname-to-ip", log.Fields{"host": host}, err)
	}

	genmac = macifyIP(addr)
	logger.Debugw(ctx, "using-ip-as-mac",
		log.Fields{"host": host,
			"mac": genmac})
	return genmac, nil
}

func macAddressToUint32Array(mac string) []uint32 {
	slist := strings.Split(mac, ":")
	result := make([]uint32, len(slist))
	var err error
	var tmp int64
	for index, val := range slist {
		if tmp, err = strconv.ParseInt(val, 16, 32); err != nil {
			return []uint32{1, 2, 3, 4, 5, 6}
		}
		result[index] = uint32(tmp)
	}
	return result
}

//GetportLabel returns the label for the NNI and the PON port based on port number and port type
func GetportLabel(portNum uint32, portType voltha.Port_PortType) (string, error) {

	switch portType {
	case voltha.Port_ETHERNET_NNI:
		return fmt.Sprintf("nni-%d", portNum), nil
	case voltha.Port_PON_OLT:
		return fmt.Sprintf("pon-%d", portNum), nil
	}

	return "", olterrors.NewErrInvalidValue(log.Fields{"port-type": portType}, nil)
}

func (dh *DeviceHandler) addPort(ctx context.Context, intfID uint32, portType voltha.Port_PortType, state string) error {
	var operStatus common.OperStatus_Types
	if state == "up" {
		operStatus = voltha.OperStatus_ACTIVE
		//populating the intfStatus map
		dh.activePorts.Store(intfID, true)
	} else {
		operStatus = voltha.OperStatus_DISCOVERED
		dh.activePorts.Store(intfID, false)
	}
	portNum := IntfIDToPortNo(intfID, portType)
	label, err := GetportLabel(portNum, portType)
	if err != nil {
		return olterrors.NewErrNotFound("port-label", log.Fields{"port-number": portNum, "port-type": portType}, err)
	}

	if port, err := dh.coreProxy.GetDevicePort(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, portNum); err == nil && port.Type == portType {
		logger.Debug(ctx, "port-already-exists-updating-oper-status-of-port")
		if err := dh.coreProxy.PortStateUpdate(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, portType, portNum, operStatus); err != nil {
			return olterrors.NewErrAdapter("failed-to-update-port-state", log.Fields{
				"device-id":   dh.device.Id,
				"port-type":   portType,
				"port-number": portNum,
				"oper-status": operStatus}, err).Log()
		}
		return nil
	}
	capacity := uint32(of.OfpPortFeatures_OFPPF_1GB_FD | of.OfpPortFeatures_OFPPF_FIBER)
	port := &voltha.Port{
		PortNo:     portNum,
		Label:      label,
		Type:       portType,
		OperStatus: operStatus,
		OfpPort: &of.OfpPort{
			HwAddr:     macAddressToUint32Array(dh.device.MacAddress),
			Config:     0,
			State:      uint32(of.OfpPortState_OFPPS_LIVE),
			Curr:       capacity,
			Advertised: capacity,
			Peer:       capacity,
			CurrSpeed:  uint32(of.OfpPortFeatures_OFPPF_1GB_FD),
			MaxSpeed:   uint32(of.OfpPortFeatures_OFPPF_1GB_FD),
		},
	}
	logger.Debugw(ctx, "sending-port-update-to-core", log.Fields{"port": port})
	if err := dh.coreProxy.PortCreated(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, port); err != nil {
		return olterrors.NewErrAdapter("error-creating-port", log.Fields{
			"device-id": dh.device.Id,
			"port-type": portType}, err)
	}
	go dh.updateLocalDevice(ctx)
	return nil
}

func (dh *DeviceHandler) updateLocalDevice(ctx context.Context) {
	dh.lockDevice.Lock()
	defer dh.lockDevice.Unlock()
	device, err := dh.coreProxy.GetDevice(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, dh.device.Id)
	if err != nil || device == nil {
		logger.Errorf(ctx, "device-not-found", log.Fields{"device-id": dh.device.Id}, err)
		return
	}
	dh.device = device
}

// nolint: gocyclo
// readIndications to read the indications from the OLT device
func (dh *DeviceHandler) readIndications(ctx context.Context) error {
	defer logger.Debugw(ctx, "indications-ended", log.Fields{"device-id": dh.device.Id})
	defer func() {
		dh.lockDevice.Lock()
		dh.isReadIndicationRoutineActive = false
		dh.lockDevice.Unlock()
	}()
	indications, err := dh.startOpenOltIndicationStream(ctx)
	if err != nil {
		return err
	}
	if disableBBSim {
		return nil
	}
	/* get device state */
	device, err := dh.coreProxy.GetDevice(ctx, dh.device.Id, dh.device.Id)
	if err != nil || device == nil {
		/*TODO: needs to handle error scenarios */
		return olterrors.NewErrNotFound("device", log.Fields{"device-id": dh.device.Id}, err)
	}

	indicationBackoff := backoff.NewExponentialBackOff()
	indicationBackoff.MaxElapsedTime = 0
	indicationBackoff.MaxInterval = 1 * time.Minute

	dh.lockDevice.Lock()
	dh.isReadIndicationRoutineActive = true
	dh.lockDevice.Unlock()

Loop:
	for {
		select {
		case <-dh.stopIndications:
			logger.Debugw(ctx, "stopping-collecting-indications-for-olt", log.Fields{"deviceID:": dh.device.Id})
			break Loop
		default:
			indication, err := indications.Recv()
			if err == io.EOF {
				logger.Infow(ctx, "eof-for-indications",
					log.Fields{"err": err,
						"device-id": dh.device.Id})
				// Use an exponential back off to prevent getting into a tight loop
				duration := indicationBackoff.NextBackOff()
				if duration == backoff.Stop {
					// If we reach a maximum then warn and reset the backoff
					// timer and keep attempting.
					logger.Warnw(ctx, "maximum-indication-backoff-reached--resetting-backoff-timer",
						log.Fields{"max-indication-backoff": indicationBackoff.MaxElapsedTime,
							"device-id": dh.device.Id})
					indicationBackoff.Reset()
				}

				// On failure process a backoff timer while watching for stopIndications
				// events
				backoff := time.NewTimer(indicationBackoff.NextBackOff())
				select {
				case <-dh.stopIndications:
					logger.Debugw(ctx, "stopping-collecting-indications-for-olt", log.Fields{"deviceID:": dh.device.Id})
					if !backoff.Stop() {
						<-backoff.C
					}
					break Loop
				case <-backoff.C:
					// backoff expired continue
				}
				if indications, err = dh.startOpenOltIndicationStream(ctx); err != nil {
					return err
				}
				continue
			}
			if err != nil {
				logger.Errorw(ctx, "read-indication-error",
					log.Fields{"err": err,
						"device-id": dh.device.Id})
				if device.AdminState == voltha.AdminState_DELETED {
					logger.Debug(ctx, "device-deleted--stopping-the-read-indication-thread")
					break Loop
				}
				// Close the stream, and re-initialize it
				if err = indications.CloseSend(); err != nil {
					// Ok to ignore here, because we landed here due to a problem on the stream
					// In all probability, the closeSend call may fail
					logger.Debugw(ctx, "error-closing-send stream--error-ignored",
						log.Fields{"err": err,
							"device-id": dh.device.Id})
				}
				if indications, err = dh.startOpenOltIndicationStream(ctx); err != nil {
					return err
				}
				// once we re-initialized the indication stream, continue to read indications
				continue
			}
			// Reset backoff if we have a successful receive
			indicationBackoff.Reset()
			// When OLT is admin down, ignore all indications.
			if device.AdminState == voltha.AdminState_DISABLED && !isIndicationAllowedDuringOltAdminDown(indication) {
				logger.Debugw(ctx, "olt-is-admin-down, ignore indication",
					log.Fields{"indication": indication,
						"device-id": dh.device.Id})
				continue
			}
			dh.handleIndication(ctx, indication)
		}
	}
	_ = indications.CloseSend() // Ok to ignore error, as we stopping the readIndication anyway

	return nil
}

func (dh *DeviceHandler) startOpenOltIndicationStream(ctx context.Context) (oop.Openolt_EnableIndicationClient, error) {
	L2oamEnableIndication(ctx, dh)

	indications, err := dh.Client.EnableIndication(ctx, new(oop.Empty))
	if err != nil {
		return nil, olterrors.NewErrCommunication("indication-read-failure", log.Fields{"device-id": dh.device.Id}, err).Log()
	}
	if indications == nil {
		return nil, olterrors.NewErrInvalidValue(log.Fields{"indications": nil, "device-id": dh.device.Id}, nil).Log()
	}

	return indications, nil
}

// isIndicationAllowedDuringOltAdminDown returns true if the indication is allowed during OLT Admin down, else false
func isIndicationAllowedDuringOltAdminDown(indication *oop.Indication) bool {
	switch indication.Data.(type) {
	case *oop.Indication_OltInd, *oop.Indication_IntfInd, *oop.Indication_IntfOperInd:
		return true

	default:
		return false
	}
}

func (dh *DeviceHandler) handleOltIndication(ctx context.Context, oltIndication *oop.OltIndication) error {
	raisedTs := time.Now().UnixNano()
	if oltIndication.OperState == "up" && dh.transitionMap.currentDeviceState != deviceStateUp {
		dh.transitionMap.Handle(ctx, DeviceUpInd)
	} else if oltIndication.OperState == "down" {
		dh.transitionMap.Handle(ctx, DeviceDownInd)
	}
	if err := dh.eventMgr.oltUpDownIndication(ctx, oltIndication, dh.device.Id, raisedTs); err != nil {
		return olterrors.NewErrAdapter("failed-indication", log.Fields{
			"device_id":  dh.device.Id,
			"indication": oltIndication,
			"timestamp":  raisedTs}, err)
	}
	return nil
}

// nolint: gocyclo
func (dh *DeviceHandler) handleIndication(ctx context.Context, indication *oop.Indication) {
	raisedTs := time.Now().UnixNano()
	switch indication.Data.(type) {
	case *oop.Indication_OltInd:
		span, ctx := log.CreateChildSpan(ctx, "olt-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		if err := dh.handleOltIndication(ctx, indication.GetOltInd()); err != nil {
			_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "olt", "device-id": dh.device.Id}, err).Log()
		}
	case *oop.Indication_IntfInd:
		span, ctx := log.CreateChildSpan(ctx, "interface-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		intfInd := indication.GetIntfInd()
		go func() {
			if err := dh.addPort(ctx, intfInd.GetIntfId(), voltha.Port_PON_OLT, intfInd.GetOperState()); err != nil {
				_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "interface", "device-id": dh.device.Id}, err).Log()
			}
		}()
		logger.Infow(ctx, "received-interface-indication", log.Fields{"InterfaceInd": intfInd, "device-id": dh.device.Id})
	case *oop.Indication_IntfOperInd:
		span, ctx := log.CreateChildSpan(ctx, "interface-oper-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		intfOperInd := indication.GetIntfOperInd()
		if intfOperInd.GetType() == "nni" {
			go func() {
				if err := dh.addPort(ctx, intfOperInd.GetIntfId(), voltha.Port_ETHERNET_NNI, intfOperInd.GetOperState()); err != nil {
					_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "interface-oper-nni", "device-id": dh.device.Id}, err).Log()
				}
			}()
			if err := dh.resourceMgr.AddNNIToKVStore(ctx, intfOperInd.GetIntfId()); err != nil {
				logger.Warn(ctx, err)
			}
		} else if intfOperInd.GetType() == "pon" {
			// TODO: Check what needs to be handled here for When PON PORT down, ONU will be down
			// Handle pon port update
			go func() {
				if err := dh.addPort(ctx, intfOperInd.GetIntfId(), voltha.Port_PON_OLT, intfOperInd.GetOperState()); err != nil {
					_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "interface-oper-pon", "device-id": dh.device.Id}, err).Log()
				}
			}()
			go dh.eventMgr.oltIntfOperIndication(ctx, indication.GetIntfOperInd(), dh.device.Id, raisedTs)
		}
		logger.Infow(ctx, "received-interface-oper-indication",
			log.Fields{"interfaceOperInd": intfOperInd,
				"device-id": dh.device.Id})
	case *oop.Indication_OnuDiscInd:
		span, ctx := log.CreateChildSpan(ctx, "onu-discovery-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		onuDiscInd := indication.GetOnuDiscInd()
		logger.Infow(ctx, "received-onu-discovery-indication", log.Fields{"OnuDiscInd": onuDiscInd, "device-id": dh.device.Id})
		sn := dh.stringifySerialNumber(onuDiscInd.SerialNumber)
		go func() {
			if err := dh.onuDiscIndication(ctx, onuDiscInd, sn); err != nil {
				_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "onu-discovery", "device-id": dh.device.Id}, err).Log()
			}
		}()
	case *oop.Indication_OnuInd:
		span, ctx := log.CreateChildSpan(ctx, "onu-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		onuInd := indication.GetOnuInd()
		logger.Infow(ctx, "received-onu-indication", log.Fields{"OnuInd": onuInd, "device-id": dh.device.Id})
		go func() {
			if err := dh.onuIndication(ctx, onuInd); err != nil {
				_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "onu", "device-id": dh.device.Id}, err).Log()
			}
		}()
	case *oop.Indication_OmciInd:
		span, ctx := log.CreateChildSpan(ctx, "omci-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		omciInd := indication.GetOmciInd()
		logger.Debugw(ctx, "received-omci-indication", log.Fields{"intf-id": omciInd.IntfId, "onu-id": omciInd.OnuId, "device-id": dh.device.Id})
		go func() {
			if err := dh.omciIndication(ctx, omciInd); err != nil {
				_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "omci", "device-id": dh.device.Id}, err).Log()
			}
		}()
	case *oop.Indication_PktInd:
		span, ctx := log.CreateChildSpan(ctx, "packet-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		pktInd := indication.GetPktInd()
		logger.Debugw(ctx, "received-packet-indication", log.Fields{
			"intf-type":   pktInd.IntfId,
			"intf-id":     pktInd.IntfId,
			"gem-port-id": pktInd.GemportId,
			"port-no":     pktInd.PortNo,
			"device-id":   dh.device.Id,
		})

		if logger.V(log.DebugLevel) {
			logger.Debugw(ctx, "received-packet-indication-packet", log.Fields{
				"intf-type":   pktInd.IntfId,
				"intf-id":     pktInd.IntfId,
				"gem-port-id": pktInd.GemportId,
				"port-no":     pktInd.PortNo,
				"packet":      hex.EncodeToString(pktInd.Pkt),
				"device-id":   dh.device.Id,
			})
		}

		go func() {
			if err := dh.handlePacketIndication(ctx, pktInd); err != nil {
				_ = olterrors.NewErrAdapter("handle-indication-error", log.Fields{"type": "packet", "device-id": dh.device.Id}, err).Log()
			}
		}()
	case *oop.Indication_PortStats:
		span, ctx := log.CreateChildSpan(ctx, "port-statistics-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		portStats := indication.GetPortStats()
		go dh.portStats.PortStatisticsIndication(ctx, portStats, dh.resourceMgr.DevInfo.GetPonPorts())
	case *oop.Indication_FlowStats:
		span, ctx := log.CreateChildSpan(ctx, "flow-stats-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		flowStats := indication.GetFlowStats()
		logger.Infow(ctx, "received-flow-stats", log.Fields{"FlowStats": flowStats, "device-id": dh.device.Id})
	case *oop.Indication_AlarmInd:
		span, ctx := log.CreateChildSpan(ctx, "alarm-indication", log.Fields{"device-id": dh.device.Id})
		defer span.Finish()

		alarmInd := indication.GetAlarmInd()
		logger.Infow(ctx, "received-alarm-indication", log.Fields{"AlarmInd": alarmInd, "device-id": dh.device.Id})
		go dh.eventMgr.ProcessEvents(ctx, alarmInd, dh.device.Id, raisedTs)
	}
}

// doStateUp handle the olt up indication and update to voltha core
func (dh *DeviceHandler) doStateUp(ctx context.Context) error {
	go startCollector(ctx, dh)

	if err := dh.coreProxy.DeviceStateUpdate(ctx, dh.device.Id, voltha.ConnectStatus_REACHABLE,
		voltha.OperStatus_ACTIVE); err != nil {
		return olterrors.NewErrAdapter("device-update-failed", log.Fields{"device-id": dh.device.Id}, err)
	}
	return nil
}

// doStateDown handle the olt down indication
func (dh *DeviceHandler) doStateDown(ctx context.Context) error {
	dh.lockDevice.Lock()
	defer dh.lockDevice.Unlock()
	logger.Debugw(ctx, "do-state-down-start", log.Fields{"device-id": dh.device.Id})

	device, err := dh.coreProxy.GetDevice(ctx, dh.device.Id, dh.device.Id)
	if err != nil || device == nil {
		/*TODO: needs to handle error scenarios */
		return olterrors.NewErrNotFound("device", log.Fields{"device-id": dh.device.Id}, err)
	}

	cloned := proto.Clone(device).(*voltha.Device)

	cloned.OperStatus = voltha.OperStatus_UNKNOWN
	dh.device = cloned

	if err = dh.coreProxy.DeviceStateUpdate(ctx, cloned.Id, cloned.ConnectStatus, cloned.OperStatus); err != nil {
		return olterrors.NewErrAdapter("state-update-failed", log.Fields{"device-id": device.Id}, err)
	}

	onuDevices, err := dh.coreProxy.GetChildDevices(ctx, dh.device.Id)
	if err != nil {
		return olterrors.NewErrAdapter("child-device-fetch-failed", log.Fields{"device-id": dh.device.Id}, err)
	}
	for _, onuDevice := range onuDevices.Items {

		// Update onu state as down in onu adapter
		onuInd := oop.OnuIndication{}
		onuInd.OperState = "down"
		err := dh.AdapterProxy.SendInterAdapterMessage(ctx, &onuInd, ic.InterAdapterMessageType_ONU_IND_REQUEST,
			"openolt", onuDevice.Type, onuDevice.Id, onuDevice.ProxyAddress.DeviceId, "")
		if err != nil {
			_ = olterrors.NewErrCommunication("inter-adapter-send-failed", log.Fields{
				"source":        "openolt",
				"onu-indicator": onuInd,
				"device-type":   onuDevice.Type,
				"device-id":     onuDevice.Id}, err).LogAt(log.ErrorLevel)
			//Do not return here and continue to process other ONUs
		}
	}
	/* Discovered ONUs entries need to be cleared , since after OLT
	   is up, it starts sending discovery indications again*/
	dh.discOnus = sync.Map{}
	logger.Debugw(ctx, "do-state-down-end", log.Fields{"device-id": device.Id})
	return nil
}

// doStateInit dial the grpc before going to init state
func (dh *DeviceHandler) doStateInit(ctx context.Context) error {
	if disableBBSim {
		return nil
	}
	var err error
	dh.clientCon, err = grpc.Dial(dh.device.GetHostAndPort(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			grpc_opentracing.StreamClientInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
		)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
		)))

	if err != nil {
		return olterrors.NewErrCommunication("dial-failure", log.Fields{
			"device-id":     dh.device.Id,
			"host-and-port": dh.device.GetHostAndPort()}, err)
	}
	return nil
}

// postInit create olt client instance to invoke RPC on the olt device
func (dh *DeviceHandler) postInit(ctx context.Context) error {
	if disableBBSim {
		dh.Client = &mocks.MockOpenoltClient{}
	} else {
		dh.Client = oop.NewOpenoltClient(dh.clientCon)
	}
	dh.transitionMap.Handle(ctx, GrpcConnected)
	return nil
}

// doStateConnected get the device info and update to voltha core
func (dh *DeviceHandler) doStateConnected(ctx context.Context) error {
	var err error
	logger.Debugw(ctx, "olt-device-connected", log.Fields{"device-id": dh.device.Id})

	device, err := dh.coreProxy.GetDevice(ctx, dh.device.Id, dh.device.Id)
	if err != nil || device == nil {
		/*TODO: needs to handle error scenarios */
		return olterrors.NewErrAdapter("device-fetch-failed", log.Fields{"device-id": dh.device.Id}, err).LogAt(log.ErrorLevel)
	}
	if device.AdminState == voltha.AdminState_DISABLED {
		logger.Debugln(ctx, "do-state-connected--device-admin-state-down")

		cloned := proto.Clone(device).(*voltha.Device)
		cloned.ConnectStatus = voltha.ConnectStatus_REACHABLE
		cloned.OperStatus = voltha.OperStatus_UNKNOWN
		dh.device = cloned
		if err = dh.coreProxy.DeviceStateUpdate(ctx, cloned.Id, cloned.ConnectStatus, cloned.OperStatus); err != nil {
			return olterrors.NewErrAdapter("device-state-update-failed", log.Fields{"device-id": dh.device.Id}, err).LogAt(log.ErrorLevel)
		}

		// Since the device was disabled before the OLT was rebooted, enforce the OLT to be Disabled after re-connection.
		_, err = dh.Client.DisableOlt(ctx, new(oop.Empty))
		if err != nil {
			return olterrors.NewErrAdapter("olt-disable-failed", log.Fields{"device-id": dh.device.Id}, err).LogAt(log.ErrorLevel)
		}
		// We should still go ahead an initialize various device handler modules so that when OLT is re-enabled, we have
		// all the modules initialized and ready to handle incoming ONUs.

		err = dh.initializeDeviceHandlerModules(ctx)
		if err != nil {
			return olterrors.NewErrAdapter("device-handler-initialization-failed", log.Fields{"device-id": dh.device.Id}, err).LogAt(log.ErrorLevel)
		}

		// Start reading indications
		go func() {
			if err = dh.readIndications(ctx); err != nil {
				_ = olterrors.NewErrAdapter("indication-read-failure", log.Fields{"device-id": dh.device.Id}, err).LogAt(log.ErrorLevel)
			}
		}()
		return nil
	}

	ports, err := dh.coreProxy.ListDevicePorts(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id)
	if err != nil {
		/*TODO: needs to handle error scenarios */
		return olterrors.NewErrAdapter("fetch-ports-failed", log.Fields{"device-id": dh.device.Id}, err)
	}
	dh.populateActivePorts(ctx, ports)
	if err := dh.disableAdminDownPorts(ctx, ports); err != nil {
		return olterrors.NewErrAdapter("port-status-update-failed", log.Fields{"ports": ports}, err)
	}

	if err := dh.initializeDeviceHandlerModules(ctx); err != nil {
		return olterrors.NewErrAdapter("device-handler-initialization-failed", log.Fields{"device-id": dh.device.Id}, err).LogAt(log.ErrorLevel)
	}

	go func() {
		if err := dh.readIndications(ctx); err != nil {
			_ = olterrors.NewErrAdapter("read-indications-failure", log.Fields{"device-id": dh.device.Id}, err).Log()
		}
	}()
	go dh.updateLocalDevice(ctx)

	if device.PmConfigs != nil {
		dh.UpdatePmConfig(ctx, device.PmConfigs)
	}
	return nil
}

func (dh *DeviceHandler) initializeDeviceHandlerModules(ctx context.Context) error {
	deviceInfo, err := dh.populateDeviceInfo(ctx)

	if err != nil {
		return olterrors.NewErrAdapter("populate-device-info-failed", log.Fields{"device-id": dh.device.Id}, err)
	}
	if dh.resourceMgr = rsrcMgr.NewResourceMgr(ctx, dh.device.Id, dh.openOLT.KVStoreAddress, dh.openOLT.KVStoreType, dh.device.Type, deviceInfo); dh.resourceMgr == nil {
		return olterrors.ErrResourceManagerInstantiating
	}

	if dh.flowMgr = NewFlowManager(ctx, dh, dh.resourceMgr); dh.flowMgr == nil {
		return olterrors.ErrResourceManagerInstantiating

	}
	/* TODO: Instantiate Alarm , stats , BW managers */
	/* Instantiating Event Manager to handle Alarms and KPIs */
	dh.eventMgr = NewEventMgr(dh.EventProxy, dh)

	dh.portStats = NewOpenOltStatsMgr(ctx, dh)

	return nil

}

func (dh *DeviceHandler) populateDeviceInfo(ctx context.Context) (*oop.DeviceInfo, error) {
	var err error
	var deviceInfo *oop.DeviceInfo

	if disableBBSim {
		deviceInfo, err = L2oamGetDeviceInfo(log.WithSpanFromContext(context.Background(), ctx), dh)
	} else {
		deviceInfo, err = dh.Client.GetDeviceInfo(log.WithSpanFromContext(context.Background(), ctx), new(oop.Empty))
	}

	if err != nil {
		return nil, olterrors.NewErrPersistence("get", "device", 0, nil, err)
	}
	if deviceInfo == nil {
		return nil, olterrors.NewErrInvalidValue(log.Fields{"device": nil}, nil)
	}

	logger.Debugw(ctx, "fetched-device-info", log.Fields{"deviceInfo": deviceInfo, "device-id": dh.device.Id})
	dh.device.Root = true
	dh.device.Vendor = deviceInfo.Vendor
	dh.device.Model = deviceInfo.Model
	dh.device.SerialNumber = deviceInfo.DeviceSerialNumber
	dh.device.HardwareVersion = deviceInfo.HardwareVersion
	dh.device.FirmwareVersion = deviceInfo.FirmwareVersion

	if deviceInfo.DeviceId == "" {
		logger.Warnw(ctx, "no-device-id-provided-using-host", log.Fields{"hostport": dh.device.GetHostAndPort()})
		host := strings.Split(dh.device.GetHostAndPort(), ":")[0]
		genmac, err := generateMacFromHost(ctx, host)
		if err != nil {
			return nil, olterrors.NewErrAdapter("failed-to-generate-mac-host", log.Fields{"host": host}, err)
		}
		logger.Debugw(ctx, "using-host-for-mac-address", log.Fields{"host": host, "mac": genmac})
		dh.device.MacAddress = genmac
	} else {
		dh.device.MacAddress = deviceInfo.DeviceId
	}

	if err := dh.coreProxy.DeviceUpdate(log.WithSpanFromContext(context.TODO(), ctx), dh.device); err != nil {
		return nil, olterrors.NewErrAdapter("device-update-failed", log.Fields{"device-id": dh.device.Id}, err)
	}

	return deviceInfo, nil
}

func startCollector(ctx context.Context, dh *DeviceHandler) {
	logger.Debugf(ctx, "starting-collector")
	for {
		select {
		case <-dh.stopCollector:
			logger.Debugw(ctx, "stopping-collector-for-olt", log.Fields{"deviceID:": dh.device.Id})
			return
		case <-time.After(time.Duration(dh.metrics.ToPmConfigs().DefaultFreq) * time.Second):

			ports, err := dh.coreProxy.ListDevicePorts(log.WithSpanFromContext(context.Background(), ctx), dh.device.Id)
			if err != nil {
				logger.Warnw(ctx, "failed-to-list-ports", log.Fields{"device-id": dh.device.Id, "error": err})
				continue
			}
			for _, port := range ports {
				// NNI Stats
				if port.Type == voltha.Port_ETHERNET_NNI {
					intfID := PortNoToIntfID(port.PortNo, voltha.Port_ETHERNET_NNI)
					cmnni := dh.portStats.collectNNIMetrics(intfID)
					logger.Debugw(ctx, "collect-nni-metrics", log.Fields{"metrics": cmnni})
					go dh.portStats.publishMetrics(ctx, cmnni, port, dh.device.Id, dh.device.Type)
					logger.Debugw(ctx, "publish-nni-metrics", log.Fields{"nni-port": port.Label})
				}
				// PON Stats
				if port.Type == voltha.Port_PON_OLT {
					intfID := PortNoToIntfID(port.PortNo, voltha.Port_PON_OLT)
					if val, ok := dh.activePorts.Load(intfID); ok && val == true {
						cmpon := dh.portStats.collectPONMetrics(intfID)
						logger.Debugw(ctx, "collect-pon-metrics", log.Fields{"metrics": cmpon})
						go dh.portStats.publishMetrics(ctx, cmpon, port, dh.device.Id, dh.device.Type)
					}
					logger.Debugw(ctx, "publish-pon-metrics", log.Fields{"pon-port": port.Label})
				}
			}
		}
	}
}

//AdoptDevice adopts the OLT device
func (dh *DeviceHandler) AdoptDevice(ctx context.Context, device *voltha.Device) {
	dh.transitionMap = NewTransitionMap(dh)
	logger.Infow(ctx, "adopt-device", log.Fields{"device-id": device.Id, "Address": device.GetHostAndPort()})
	dh.transitionMap.Handle(ctx, DeviceInit)

	if err := dh.coreProxy.DevicePMConfigUpdate(ctx, dh.metrics.ToPmConfigs()); err != nil {
		_ = olterrors.NewErrAdapter("error-updating-performance-metrics", log.Fields{"device-id": device.Id}, err).LogAt(log.ErrorLevel)
	}

	go startHeartbeatCheck(ctx, dh)
}

//GetOfpDeviceInfo Gets the Ofp information of the given device
func (dh *DeviceHandler) GetOfpDeviceInfo(device *voltha.Device) (*ic.SwitchCapability, error) {
	return &ic.SwitchCapability{
		Desc: &of.OfpDesc{
			MfrDesc:   "VOLTHA Project",
			HwDesc:    "open_pon",
			SwDesc:    "open_pon",
			SerialNum: dh.device.SerialNumber,
		},
		SwitchFeatures: &of.OfpSwitchFeatures{
			NBuffers: 256,
			NTables:  2,
			Capabilities: uint32(of.OfpCapabilities_OFPC_FLOW_STATS |
				of.OfpCapabilities_OFPC_TABLE_STATS |
				of.OfpCapabilities_OFPC_PORT_STATS |
				of.OfpCapabilities_OFPC_GROUP_STATS),
		},
	}, nil
}

func (dh *DeviceHandler) omciIndication(ctx context.Context, omciInd *oop.OmciIndication) error {
	logger.Debugw(ctx, "omci-indication", log.Fields{"intf-id": omciInd.IntfId, "onu-id": omciInd.OnuId, "device-id": dh.device.Id})
	var deviceType string
	var deviceID string
	var proxyDeviceID string

	transid := extractOmciTransactionID(omciInd.Pkt)
	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "recv-omci-msg", log.Fields{"intf-id": omciInd.IntfId, "onu-id": omciInd.OnuId, "device-id": dh.device.Id,
			"omci-transaction-id": transid, "omci-msg": hex.EncodeToString(omciInd.Pkt)})
	}

	onuKey := dh.formOnuKey(omciInd.IntfId, omciInd.OnuId)

	if onuInCache, ok := dh.onus.Load(onuKey); !ok {

		logger.Debugw(ctx, "omci-indication-for-a-device-not-in-cache.", log.Fields{"intf-id": omciInd.IntfId, "onu-id": omciInd.OnuId, "device-id": dh.device.Id})
		ponPort := IntfIDToPortNo(omciInd.GetIntfId(), voltha.Port_PON_OLT)
		kwargs := make(map[string]interface{})
		kwargs["onu_id"] = omciInd.OnuId
		kwargs["parent_port_no"] = ponPort

		onuDevice, err := dh.coreProxy.GetChildDevice(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, kwargs)
		if err != nil {
			return olterrors.NewErrNotFound("onu", log.Fields{
				"intf-id": omciInd.IntfId,
				"onu-id":  omciInd.OnuId}, err)
		}
		deviceType = onuDevice.Type
		deviceID = onuDevice.Id
		proxyDeviceID = onuDevice.ProxyAddress.DeviceId
		//if not exist in cache, then add to cache.
		dh.onus.Store(onuKey, NewOnuDevice(deviceID, deviceType, onuDevice.SerialNumber, omciInd.OnuId, omciInd.IntfId, proxyDeviceID, false))
	} else {
		//found in cache
		logger.Debugw(ctx, "omci-indication-for-a-device-in-cache.", log.Fields{"intf-id": omciInd.IntfId, "onu-id": omciInd.OnuId, "device-id": dh.device.Id})
		deviceType = onuInCache.(*OnuDevice).deviceType
		deviceID = onuInCache.(*OnuDevice).deviceID
		proxyDeviceID = onuInCache.(*OnuDevice).proxyDeviceID
	}

	omciMsg := &ic.InterAdapterOmciMessage{Message: omciInd.Pkt}
	if err := dh.AdapterProxy.SendInterAdapterMessage(log.WithSpanFromContext(context.Background(), ctx), omciMsg,
		ic.InterAdapterMessageType_OMCI_REQUEST, dh.device.Type, deviceType,
		deviceID, proxyDeviceID, ""); err != nil {
		return olterrors.NewErrCommunication("omci-request", log.Fields{
			"source":          dh.device.Type,
			"destination":     deviceType,
			"onu-id":          deviceID,
			"proxy-device-id": proxyDeviceID}, err)
	}
	return nil
}

//ProcessInterAdapterMessage sends the proxied messages to the target device
// If the proxy address is not found in the unmarshalled message, it first fetches the onu device for which the message
// is meant, and then send the unmarshalled omci message to this onu
func (dh *DeviceHandler) ProcessInterAdapterMessage(ctx context.Context, msg *ic.InterAdapterMessage) error {
	logger.Debugw(ctx, "process-inter-adapter-message", log.Fields{"msgID": msg.Header.Id})
	if msg.Header.Type == ic.InterAdapterMessageType_OMCI_REQUEST {
		msgID := msg.Header.Id
		fromTopic := msg.Header.FromTopic
		toTopic := msg.Header.ToTopic
		toDeviceID := msg.Header.ToDeviceId
		proxyDeviceID := msg.Header.ProxyDeviceId

		logger.Debugw(ctx, "omci-request-message-header", log.Fields{"msgID": msgID, "fromTopic": fromTopic, "toTopic": toTopic, "toDeviceID": toDeviceID, "proxyDeviceID": proxyDeviceID})

		msgBody := msg.GetBody()

		omciMsg := &ic.InterAdapterOmciMessage{}
		if err := ptypes.UnmarshalAny(msgBody, omciMsg); err != nil {
			return olterrors.NewErrAdapter("cannot-unmarshal-omci-msg-body", log.Fields{"msgbody": msgBody}, err)
		}

		if omciMsg.GetProxyAddress() == nil {
			onuDevice, err := dh.coreProxy.GetDevice(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, toDeviceID)
			if err != nil {
				return olterrors.NewErrNotFound("onu", log.Fields{
					"device-id":     dh.device.Id,
					"onu-device-id": toDeviceID}, err)
			}
			logger.Debugw(ctx, "device-retrieved-from-core", log.Fields{"msgID": msgID, "fromTopic": fromTopic, "toTopic": toTopic, "toDeviceID": toDeviceID, "proxyDeviceID": proxyDeviceID})
			if err := dh.sendProxiedMessage(ctx, onuDevice, omciMsg); err != nil {
				return olterrors.NewErrCommunication("send-failed", log.Fields{
					"device-id":     dh.device.Id,
					"onu-device-id": toDeviceID}, err)
			}
		} else {
			logger.Debugw(ctx, "proxy-address-found-in-omci-message", log.Fields{"msgID": msgID, "fromTopic": fromTopic, "toTopic": toTopic, "toDeviceID": toDeviceID, "proxyDeviceID": proxyDeviceID})
			if err := dh.sendProxiedMessage(ctx, nil, omciMsg); err != nil {
				return olterrors.NewErrCommunication("send-failed", log.Fields{
					"device-id":     dh.device.Id,
					"onu-device-id": toDeviceID}, err)
			}
		}
	} else if msg.Header.Type == ic.InterAdapterMessageType_ONU_IND_REQUEST {
		logger.Debugw(ctx, "got-message-from-onu", log.Fields{"message": msg.Header.Id})
		if err := dh.receivedMsgFromOnu(ctx, msg); err != nil {
			return err
		}
	} else {
		return olterrors.NewErrInvalidValue(log.Fields{"inter-adapter-message-type": msg.Header.Type}, nil)
	}
	return nil
}

func (dh *DeviceHandler) sendProxiedMessage(ctx context.Context, onuDevice *voltha.Device, omciMsg *ic.InterAdapterOmciMessage) error {
	var intfID uint32
	var onuID uint32
	var connectStatus common.ConnectStatus_Types
	if onuDevice != nil {
		intfID = onuDevice.ProxyAddress.GetChannelId()
		onuID = onuDevice.ProxyAddress.GetOnuId()
		connectStatus = onuDevice.ConnectStatus
	} else {
		intfID = omciMsg.GetProxyAddress().GetChannelId()
		onuID = omciMsg.GetProxyAddress().GetOnuId()
		connectStatus = omciMsg.GetConnectStatus()
	}
	if connectStatus != voltha.ConnectStatus_REACHABLE {
		logger.Debugw(ctx, "onu-not-reachable--cannot-send-omci", log.Fields{"intf-id": intfID, "onu-id": onuID})

		return olterrors.NewErrCommunication("unreachable", log.Fields{
			"intf-id": intfID,
			"onu-id":  onuID}, nil)
	}

	var omciMessage *oop.OmciMsg
	hexPkt := make([]byte, hex.EncodedLen(len(omciMsg.Message)))
	hex.Encode(hexPkt, omciMsg.Message)
	omciMessage = &oop.OmciMsg{IntfId: intfID, OnuId: onuID, Pkt: hexPkt}

	transid := extractOmciTransactionID(omciMsg.Message)
	logger.Debugw(ctx, "sent-omci-msg", log.Fields{"intf-id": intfID, "onu-id": onuID,
		"omciTransactionID": transid, "omciMsg": string(omciMessage.Pkt)})

	_, err := dh.Client.OmciMsgOut(log.WithSpanFromContext(context.Background(), ctx), omciMessage)
	if err != nil {
		return olterrors.NewErrCommunication("omci-send-failed", log.Fields{
			"intf-id": intfID,
			"onu-id":  onuID,
			"message": omciMessage}, err)
	}

	if disableBBSim {
		if err = dh.fakeOmciIndication(ctx, intfID, onuID, omciMsg.Message); err != nil {
			return nil
		}
	}
	return nil
}

func (dh *DeviceHandler) activateONU(ctx context.Context, intfID uint32, onuID int64, serialNum *oop.SerialNumber, serialNumber string) error {
	logger.Debugw(ctx, "activate-onu", log.Fields{"intf-id": intfID, "onu-id": onuID, "serialNum": serialNum, "serialNumber": serialNumber, "device-id": dh.device.Id})
	if err := dh.flowMgr.UpdateOnuInfo(ctx, intfID, uint32(onuID), serialNumber); err != nil {
		return olterrors.NewErrAdapter("onu-activate-failed", log.Fields{"onu": onuID, "intf-id": intfID}, err)
	}
	var pir uint32 = 1000000
	Onu := oop.Onu{IntfId: intfID, OnuId: uint32(onuID), SerialNumber: serialNum, Pir: pir}
	if _, err := dh.Client.ActivateOnu(ctx, &Onu); err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.AlreadyExists {
			logger.Debugw(ctx, "onu-activation-in-progress", log.Fields{"SerialNumber": serialNumber, "onu-id": onuID, "device-id": dh.device.Id})

		} else {
			return olterrors.NewErrAdapter("onu-activate-failed", log.Fields{"onu": Onu, "device-id": dh.device.Id}, err)
		}
	} else {
		logger.Infow(ctx, "activated-onu", log.Fields{"SerialNumber": serialNumber, "device-id": dh.device.Id})
	}
	return nil
}

func (dh *DeviceHandler) onuDiscIndication(ctx context.Context, onuDiscInd *oop.OnuDiscIndication, sn string) error {
	channelID := onuDiscInd.GetIntfId()
	parentPortNo := IntfIDToPortNo(onuDiscInd.GetIntfId(), voltha.Port_PON_OLT)

	logger.Infow(ctx, "new-discovery-indication", log.Fields{"sn": sn})

	kwargs := make(map[string]interface{})
	if sn != "" {
		kwargs["serial_number"] = sn
	} else {
		return olterrors.NewErrInvalidValue(log.Fields{"serial-number": sn}, nil)
	}

	var alarmInd oop.OnuAlarmIndication
	raisedTs := time.Now().UnixNano()
	if _, loaded := dh.discOnus.LoadOrStore(sn, true); loaded {

		/* When PON cable disconnected and connected back from OLT, it was expected OnuAlarmIndication
		   with "los_status: off" should be raised but BAL does not raise this Alarm hence manually sending
		   OnuLosClear event on receiving OnuDiscoveryIndication for the Onu after checking whether
		   OnuLosRaise event sent for it */
		dh.onus.Range(func(Onukey interface{}, onuInCache interface{}) bool {
			if onuInCache.(*OnuDevice).serialNumber == sn && onuInCache.(*OnuDevice).losRaised {
				if onuDiscInd.GetIntfId() != onuInCache.(*OnuDevice).intfID {
					logger.Warnw(ctx, "onu-is-on-a-different-intf-id-now", log.Fields{
						"previousIntfId": onuInCache.(*OnuDevice).intfID,
						"currentIntfId":  onuDiscInd.GetIntfId()})
					// TODO:: Should we need to ignore raising OnuLosClear event
					// when onu connected to different PON?
				}
				alarmInd.IntfId = onuInCache.(*OnuDevice).intfID
				alarmInd.OnuId = onuInCache.(*OnuDevice).onuID
				alarmInd.LosStatus = statusCheckOff
				go func() {
					if err := dh.eventMgr.onuAlarmIndication(ctx, &alarmInd, onuInCache.(*OnuDevice).deviceID, raisedTs); err != nil {
						logger.Debugw(ctx, "indication-failed", log.Fields{"error": err})
					}
				}()
			}
			return true
		})

		logger.Warnw(ctx, "onu-sn-is-already-being-processed", log.Fields{"sn": sn})
		return nil
	}

	var onuID uint32

	onuDevice, err := dh.coreProxy.GetChildDevice(ctx, dh.device.Id, kwargs)

	if err != nil {
		logger.Debugw(ctx, "core-proxy-get-child-device-failed", log.Fields{"parentDevice": dh.device.Id, "err": err, "sn": sn})
		if e, ok := status.FromError(err); ok {
			logger.Debugw(ctx, "core-proxy-get-child-device-failed-with-code", log.Fields{"errCode": e.Code(), "sn": sn})
			switch e.Code() {
			case codes.Internal:
				// this probably means NOT FOUND, so just create a new device
				onuDevice = nil
			case codes.DeadlineExceeded:
				// if the call times out, cleanup and exit
				dh.discOnus.Delete(sn)
				return olterrors.NewErrTimeout("get-child-device", log.Fields{"device-id": dh.device.Id}, err)
			}
		}
	}

	if onuDevice == nil {
		// NOTE this should happen a single time, and only if GetChildDevice returns NotFound
		logger.Debugw(ctx, "creating-new-onu", log.Fields{"sn": sn})
		// we need to create a new ChildDevice
		ponintfid := onuDiscInd.GetIntfId()
		dh.lockDevice.Lock()
		onuID, err = dh.resourceMgr.GetONUID(ctx, ponintfid)
		dh.lockDevice.Unlock()

		logger.Infow(ctx, "creating-new-onu-got-onu-id", log.Fields{"sn": sn, "onuId": onuID})

		if err != nil {
			// if we can't create an ID in resource manager,
			// cleanup and exit
			dh.discOnus.Delete(sn)
			return olterrors.NewErrAdapter("resource-manager-get-onu-id-failed", log.Fields{
				"pon-intf-id":   ponintfid,
				"serial-number": sn}, err)
		}

		if onuDevice, err = dh.coreProxy.ChildDeviceDetected(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, int(parentPortNo),
			"", int(channelID), string(onuDiscInd.SerialNumber.GetVendorId()), sn, int64(onuID)); err != nil {
			dh.discOnus.Delete(sn)
			dh.resourceMgr.FreeonuID(ctx, ponintfid, []uint32{onuID}) // NOTE I'm not sure this method is actually cleaning up the right thing
			return olterrors.NewErrAdapter("core-proxy-child-device-detected-failed", log.Fields{
				"pon-intf-id":   ponintfid,
				"serial-number": sn}, err)
		}
		if err := dh.eventMgr.OnuDiscoveryIndication(ctx, onuDiscInd, dh.device.Id, onuDevice.Id, onuID, sn, time.Now().UnixNano()); err != nil {
			logger.Warnw(ctx, "discovery-indication-failed", log.Fields{"error": err})
		}
		logger.Infow(ctx, "onu-child-device-added",
			log.Fields{"onuDevice": onuDevice,
				"sn":        sn,
				"onu-id":    onuID,
				"device-id": dh.device.Id})
	}

	onuID = onuDevice.ProxyAddress.OnuId
	logger.Debugw(ctx, "onu-discovery-indication-key-create",
		log.Fields{"onu-id": onuID,
			"intfId": onuDiscInd.GetIntfId(),
			"sn":     sn})
	onuKey := dh.formOnuKey(onuDiscInd.GetIntfId(), onuID)

	onuDev := NewOnuDevice(onuDevice.Id, onuDevice.Type, onuDevice.SerialNumber, onuID, onuDiscInd.GetIntfId(), onuDevice.ProxyAddress.DeviceId, false)
	dh.onus.Store(onuKey, onuDev)
	logger.Debugw(ctx, "new-onu-device-discovered",
		log.Fields{"onu": onuDev,
			"sn": sn})

	if err := dh.coreProxy.DeviceStateUpdate(ctx, onuDevice.Id, common.ConnectStatus_REACHABLE, common.OperStatus_DISCOVERED); err != nil {
		return olterrors.NewErrAdapter("failed-to-update-device-state", log.Fields{
			"device-id":     onuDevice.Id,
			"serial-number": sn}, err)
	}
	logger.Infow(ctx, "onu-discovered-reachable", log.Fields{"device-id": onuDevice.Id, "sn": sn})
	if err := dh.activateONU(ctx, onuDiscInd.IntfId, int64(onuID), onuDiscInd.SerialNumber, sn); err != nil {
		return olterrors.NewErrAdapter("onu-activation-failed", log.Fields{
			"device-id":     onuDevice.Id,
			"serial-number": sn}, err)
	}
	return nil
}
func (dh *DeviceHandler) onuDiscIndication2(ctx context.Context, onuDiscInd *oop.OnuDiscIndication, sn string, macAddress string) error {
	channelID := onuDiscInd.GetIntfId()
	parentPortNo := IntfIDToPortNo(onuDiscInd.GetIntfId(), voltha.Port_PON_OLT)

	logger.Infow(ctx, "new-discovery-indication", log.Fields{"sn": sn})

	kwargs := make(map[string]interface{})
	if sn != "" {
		kwargs["serial_number"] = sn
	} else {
		return olterrors.NewErrInvalidValue(log.Fields{"serial-number": sn}, nil)
	}

	var alarmInd oop.OnuAlarmIndication
	raisedTs := time.Now().UnixNano()
	if _, loaded := dh.discOnus.LoadOrStore(sn, true); loaded {

		/* When PON cable disconnected and connected back from OLT, it was expected OnuAlarmIndication
		   with "los_status: off" should be raised but BAL does not raise this Alarm hence manually sending
		   OnuLosClear event on receiving OnuDiscoveryIndication for the Onu after checking whether
		   OnuLosRaise event sent for it */
		dh.onus.Range(func(Onukey interface{}, onuInCache interface{}) bool {
			if onuInCache.(*OnuDevice).serialNumber == sn && onuInCache.(*OnuDevice).losRaised {
				if onuDiscInd.GetIntfId() != onuInCache.(*OnuDevice).intfID {
					logger.Warnw(ctx, "onu-is-on-a-different-intf-id-now", log.Fields{
						"previousIntfId": onuInCache.(*OnuDevice).intfID,
						"currentIntfId":  onuDiscInd.GetIntfId()})
					// TODO:: Should we need to ignore raising OnuLosClear event
					// when onu connected to different PON?
				}
				alarmInd.IntfId = onuInCache.(*OnuDevice).intfID
				alarmInd.OnuId = onuInCache.(*OnuDevice).onuID
				alarmInd.LosStatus = statusCheckOff
				go func() {
					if err := dh.eventMgr.onuAlarmIndication(ctx, &alarmInd, onuInCache.(*OnuDevice).deviceID, raisedTs); err != nil {
						logger.Debugw(ctx, "indication-failed", log.Fields{"error": err})
					}
				}()
			}
			return true
		})

		logger.Warnw(ctx, "onu-sn-is-already-being-processed", log.Fields{"sn": sn})
		return nil
	}

	var onuID uint32

	onuDevice, err := dh.coreProxy.GetChildDevice(ctx, dh.device.Id, kwargs)

	if err != nil {
		logger.Debugw(ctx, "core-proxy-get-child-device-failed", log.Fields{"parentDevice": dh.device.Id, "err": err, "sn": sn})
		if e, ok := status.FromError(err); ok {
			logger.Debugw(ctx, "core-proxy-get-child-device-failed-with-code", log.Fields{"errCode": e.Code(), "sn": sn})
			switch e.Code() {
			case codes.Internal:
				// this probably means NOT FOUND, so just create a new device
				onuDevice = nil
			case codes.DeadlineExceeded:
				// if the call times out, cleanup and exit
				dh.discOnus.Delete(sn)
				return olterrors.NewErrTimeout("get-child-device", log.Fields{"device-id": dh.device.Id}, err)
			}
		}
	}

	if onuDevice == nil {
		// NOTE this should happen a single time, and only if GetChildDevice returns NotFound
		logger.Debugw(ctx, "creating-new-onu", log.Fields{"sn": sn})
		// we need to create a new ChildDevice
		ponintfid := onuDiscInd.GetIntfId()
		dh.lockDevice.Lock()
		onuID, err = dh.resourceMgr.GetONUID(ctx, ponintfid)
		dh.lockDevice.Unlock()

		logger.Infow(ctx, "creating-new-onu-got-onu-id", log.Fields{"sn": sn, "onuId": onuID})

		if err != nil {
			// if we can't create an ID in resource manager,
			// cleanup and exit
			dh.discOnus.Delete(sn)
			return olterrors.NewErrAdapter("resource-manager-get-onu-id-failed", log.Fields{
				"pon-intf-id":   ponintfid,
				"serial-number": sn}, err)
		}

		if onuDevice, err = dh.coreProxy.ChildDeviceDetected(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, int(parentPortNo),
			"", int(channelID), string(onuDiscInd.SerialNumber.GetVendorId()), sn, int64(onuID)); err != nil {
			dh.discOnus.Delete(sn)
			dh.resourceMgr.FreeonuID(ctx, ponintfid, []uint32{onuID}) // NOTE I'm not sure this method is actually cleaning up the right thing
			return olterrors.NewErrAdapter("core-proxy-child-device-detected-failed", log.Fields{
				"pon-intf-id":   ponintfid,
				"serial-number": sn}, err)
		}
		if err := dh.eventMgr.OnuDiscoveryIndication(ctx, onuDiscInd, dh.device.Id, onuDevice.Id, onuID, sn, time.Now().UnixNano()); err != nil {
			logger.Warnw(ctx, "discovery-indication-failed", log.Fields{"error": err})
		}
		logger.Infow(ctx, "onu-child-device-added",
			log.Fields{"onuDevice": onuDevice,
				"sn":        sn,
				"onu-id":    onuID,
				"device-id": dh.device.Id})
	}

	onuID = onuDevice.ProxyAddress.OnuId
	logger.Debugw(ctx, "onu-discovery-indication-key-create",
		log.Fields{"onu-id": onuID,
			"intfId": onuDiscInd.GetIntfId(),
			"sn":     sn})
	onuKey := dh.formOnuKey(onuDiscInd.GetIntfId(), onuID)

	onuDev := NewOnuDevice(onuDevice.Id, onuDevice.Type, onuDevice.SerialNumber, onuID, onuDiscInd.GetIntfId(), onuDevice.ProxyAddress.DeviceId, false)
	dh.onus.Store(onuKey, onuDev)
	logger.Debugw(ctx, "new-onu-device-discovered",
		log.Fields{"onu": onuDev,
			"sn": sn})

	if err := dh.coreProxy.DeviceStateUpdate(ctx, onuDevice.Id, common.ConnectStatus_REACHABLE, common.OperStatus_DISCOVERED); err != nil {
		return olterrors.NewErrAdapter("failed-to-update-device-state", log.Fields{
			"device-id":     onuDevice.Id,
			"serial-number": sn}, err)
	}
	logger.Infow(ctx, "onu-discovered-reachable", log.Fields{"device-id": onuDevice.Id, "sn": sn})
	if err := dh.activateONU(ctx, onuDiscInd.IntfId, int64(onuID), onuDiscInd.SerialNumber, sn); err != nil {
		return olterrors.NewErrAdapter("onu-activation-failed", log.Fields{
			"device-id":     onuDevice.Id,
			"serial-number": sn}, err)
	}

	device := FindL2oamDevice(macAddress)
	if device != nil {
		logger.Debug(ctx, fmt.Sprintf("onuDiscIndication2() onu.initialize() called. deviceId=%s, onuId=%d", onuDevice.Id, onuID))
		onu := device.(*L2oamOnuDevice)
		onu.update(macAddress, onuDevice.Id, onuID)
		onu.setActiveState(ctx, true)
	}
	onuDevice.MacAddress = macAddress
	onuDevice.Vendor = string(onuDiscInd.SerialNumber.VendorSpecific)
	if err := dh.coreProxy.DeviceUpdate(log.WithSpanFromContext(context.TODO(), ctx), onuDevice); err != nil {
		logger.Error(ctx, fmt.Sprintf("onuDiscIndication2() DeviceUpdate() failed. macAddress=%s, deviceId=%s", macAddress, onuDevice.Id))
	}

	return nil
}
func (dh *DeviceHandler) onuIndication(ctx context.Context, onuInd *oop.OnuIndication) error {
	serialNumber := dh.stringifySerialNumber(onuInd.SerialNumber)

	kwargs := make(map[string]interface{})
	ponPort := IntfIDToPortNo(onuInd.GetIntfId(), voltha.Port_PON_OLT)
	var onuDevice *voltha.Device
	var err error
	foundInCache := false
	logger.Debugw(ctx, "onu-indication-key-create",
		log.Fields{"onuId": onuInd.OnuId,
			"intfId":    onuInd.GetIntfId(),
			"device-id": dh.device.Id})
	onuKey := dh.formOnuKey(onuInd.GetIntfId(), onuInd.OnuId)

	errFields := log.Fields{"device-id": dh.device.Id}

	if onuInCache, ok := dh.onus.Load(onuKey); ok {

		//If ONU id is discovered before then use GetDevice to get onuDevice because it is cheaper.
		foundInCache = true
		errFields["onu-id"] = onuInCache.(*OnuDevice).deviceID
		onuDevice, err = dh.coreProxy.GetDevice(ctx, dh.device.Id, onuInCache.(*OnuDevice).deviceID)
	} else {
		//If ONU not found in adapter cache then we have to use GetChildDevice to get onuDevice
		if serialNumber != "" {
			kwargs["serial_number"] = serialNumber
			errFields["serial-number"] = serialNumber
		} else {
			kwargs["onu_id"] = onuInd.OnuId
			kwargs["parent_port_no"] = ponPort
			errFields["onu-id"] = onuInd.OnuId
			errFields["parent-port-no"] = ponPort
		}
		onuDevice, err = dh.coreProxy.GetChildDevice(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, kwargs)
	}

	if err != nil || onuDevice == nil {
		return olterrors.NewErrNotFound("onu-device", errFields, err)
	}

	if onuDevice.ParentPortNo != ponPort {
		logger.Warnw(ctx, "onu-is-on-a-different-intf-id-now", log.Fields{
			"previousIntfId": onuDevice.ParentPortNo,
			"currentIntfId":  ponPort})
	}

	if onuDevice.ProxyAddress.OnuId != onuInd.OnuId {
		logger.Warnw(ctx, "onu-id-mismatch-possible-if-voltha-and-olt-rebooted", log.Fields{
			"expected-onu-id": onuDevice.ProxyAddress.OnuId,
			"received-onu-id": onuInd.OnuId,
			"device-id":       dh.device.Id})
	}
	if !foundInCache {
		onuKey := dh.formOnuKey(onuInd.GetIntfId(), onuInd.GetOnuId())

		dh.onus.Store(onuKey, NewOnuDevice(onuDevice.Id, onuDevice.Type, onuDevice.SerialNumber, onuInd.GetOnuId(), onuInd.GetIntfId(), onuDevice.ProxyAddress.DeviceId, false))

	}
	if err := dh.updateOnuStates(ctx, onuDevice, onuInd); err != nil {
		return olterrors.NewErrCommunication("state-update-failed", errFields, err)
	}
	return nil
}

func (dh *DeviceHandler) updateOnuStates(ctx context.Context, onuDevice *voltha.Device, onuInd *oop.OnuIndication) error {
	logger.Debugw(ctx, "onu-indication-for-state", log.Fields{"onuIndication": onuInd, "device-id": onuDevice.Id, "operStatus": onuDevice.OperStatus, "adminStatus": onuDevice.AdminState})
	if onuInd.AdminState == "down" || onuInd.OperState == "down" {
		// The ONU has gone admin_state "down" or oper_state "down" - we expect the ONU to send discovery again
		// The ONU admin_state is "up" while "oper_state" is down in cases where ONU activation fails. In this case
		// the ONU sends Discovery again.
		dh.discOnus.Delete(onuDevice.SerialNumber)
		// Tests have shown that we sometimes get OperState as NOT down even if AdminState is down, forcing it
		if onuInd.OperState != "down" {
			logger.Warnw(ctx, "onu-admin-state-down", log.Fields{"operState": onuInd.OperState})
			onuInd.OperState = "down"
		}
	}

	switch onuInd.OperState {
	case "down":
		logger.Debugw(ctx, "sending-interadapter-onu-indication", log.Fields{"onuIndication": onuInd, "device-id": onuDevice.Id, "operStatus": onuDevice.OperStatus, "adminStatus": onuDevice.AdminState})
		// TODO NEW CORE do not hardcode adapter name. Handler needs Adapter reference
		err := dh.AdapterProxy.SendInterAdapterMessage(ctx, onuInd, ic.InterAdapterMessageType_ONU_IND_REQUEST,
			"openolt", onuDevice.Type, onuDevice.Id, onuDevice.ProxyAddress.DeviceId, "")
		if err != nil {
			return olterrors.NewErrCommunication("inter-adapter-send-failed", log.Fields{
				"onu-indicator": onuInd,
				"source":        "openolt",
				"device-type":   onuDevice.Type,
				"device-id":     onuDevice.Id}, err)
		}
	case "up":
		logger.Debugw(ctx, "sending-interadapter-onu-indication", log.Fields{"onuIndication": onuInd, "device-id": onuDevice.Id, "operStatus": onuDevice.OperStatus, "adminStatus": onuDevice.AdminState})
		// TODO NEW CORE do not hardcode adapter name. Handler needs Adapter reference
		err := dh.AdapterProxy.SendInterAdapterMessage(ctx, onuInd, ic.InterAdapterMessageType_ONU_IND_REQUEST,
			"openolt", onuDevice.Type, onuDevice.Id, onuDevice.ProxyAddress.DeviceId, "")
		if err != nil {
			return olterrors.NewErrCommunication("inter-adapter-send-failed", log.Fields{
				"onu-indicator": onuInd,
				"source":        "openolt",
				"device-type":   onuDevice.Type,
				"device-id":     onuDevice.Id}, err)
		}
	default:
		return olterrors.NewErrInvalidValue(log.Fields{"oper-state": onuInd.OperState}, nil)
	}
	return nil
}

func (dh *DeviceHandler) stringifySerialNumber(serialNum *oop.SerialNumber) string {
	if serialNum != nil {
		return string(serialNum.VendorId) + dh.stringifyVendorSpecific(serialNum.VendorSpecific)
	}
	return ""
}
func (dh *DeviceHandler) deStringifySerialNumber(serialNum string) (*oop.SerialNumber, error) {
	decodedStr, err := hex.DecodeString(serialNum[4:])
	if err != nil {
		return nil, err
	}
	return &oop.SerialNumber{
		VendorId:       []byte(serialNum[:4]),
		VendorSpecific: []byte(decodedStr),
	}, nil
}

func (dh *DeviceHandler) stringifyVendorSpecific(vendorSpecific []byte) string {
	tmp := fmt.Sprintf("%x", (uint32(vendorSpecific[0])>>4)&0x0f) +
		fmt.Sprintf("%x", uint32(vendorSpecific[0]&0x0f)) +
		fmt.Sprintf("%x", (uint32(vendorSpecific[1])>>4)&0x0f) +
		fmt.Sprintf("%x", (uint32(vendorSpecific[1]))&0x0f) +
		fmt.Sprintf("%x", (uint32(vendorSpecific[2])>>4)&0x0f) +
		fmt.Sprintf("%x", (uint32(vendorSpecific[2]))&0x0f) +
		fmt.Sprintf("%x", (uint32(vendorSpecific[3])>>4)&0x0f) +
		fmt.Sprintf("%x", (uint32(vendorSpecific[3]))&0x0f)
	return tmp
}

//UpdateFlowsBulk upates the bulk flow
func (dh *DeviceHandler) UpdateFlowsBulk() error {
	return olterrors.ErrNotImplemented
}

//GetChildDevice returns the child device for given parent port and onu id
func (dh *DeviceHandler) GetChildDevice(ctx context.Context, parentPort, onuID uint32) (*voltha.Device, error) {
	logger.Debugw(ctx, "getchilddevice",
		log.Fields{"pon-port": parentPort,
			"onu-id":    onuID,
			"device-id": dh.device.Id})
	kwargs := make(map[string]interface{})
	kwargs["onu_id"] = onuID
	kwargs["parent_port_no"] = parentPort
	onuDevice, err := dh.coreProxy.GetChildDevice(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, kwargs)
	if err != nil {
		return nil, olterrors.NewErrNotFound("onu-device", log.Fields{
			"intf-id": parentPort,
			"onu-id":  onuID}, err)
	}
	logger.Debugw(ctx, "successfully-received-child-device-from-core", log.Fields{"child-device-id": onuDevice.Id, "child-device-sn": onuDevice.SerialNumber})
	return onuDevice, nil
}

// SendPacketInToCore sends packet-in to core
// For this, it calls SendPacketIn of the core-proxy which uses a device specific topic to send the request.
// The adapter handling the device creates a device specific topic
func (dh *DeviceHandler) SendPacketInToCore(ctx context.Context, logicalPort uint32, packetPayload []byte) error {
	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "send-packet-in-to-core", log.Fields{
			"port":      logicalPort,
			"packet":    hex.EncodeToString(packetPayload),
			"device-id": dh.device.Id,
		})
	}
	if err := dh.coreProxy.SendPacketIn(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, logicalPort, packetPayload); err != nil {
		return olterrors.NewErrCommunication("packet-send-failed", log.Fields{
			"source":       "adapter",
			"destination":  "core",
			"device-id":    dh.device.Id,
			"logical-port": logicalPort,
			"packet":       hex.EncodeToString(packetPayload)}, err)
	}
	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "sent-packet-in-to-core-successfully", log.Fields{
			"packet":    hex.EncodeToString(packetPayload),
			"device-id": dh.device.Id,
		})
	}
	return nil
}

// UpdatePmConfig updates the pm metrics.
func (dh *DeviceHandler) UpdatePmConfig(ctx context.Context, pmConfigs *voltha.PmConfigs) {
	logger.Infow(ctx, "update-pm-configs", log.Fields{"device-id": dh.device.Id, "pm-configs": pmConfigs})

	if pmConfigs.DefaultFreq != dh.metrics.ToPmConfigs().DefaultFreq {
		dh.metrics.UpdateFrequency(pmConfigs.DefaultFreq)
		logger.Debugf(ctx, "frequency-updated")
	}

	if !pmConfigs.Grouped {
		metrics := dh.metrics.GetSubscriberMetrics()
		for _, m := range pmConfigs.Metrics {
			metrics[m.Name].Enabled = m.Enabled

		}
	}
}

//UpdateFlowsIncrementally updates the device flow
func (dh *DeviceHandler) UpdateFlowsIncrementally(ctx context.Context, device *voltha.Device, flows *of.FlowChanges, groups *of.FlowGroupChanges, flowMetadata *voltha.FlowMetadata) error {
	logger.Debugw(ctx, "received-incremental-flowupdate-in-device-handler", log.Fields{"device-id": device.Id, "flows": flows, "groups": groups, "flowMetadata": flowMetadata})

	var errorsList []error

	if flows != nil {
		for _, flow := range flows.ToRemove.Items {
			dh.incrementActiveFlowRemoveCount(ctx, flow)

			logger.Debugw(ctx, "removing-flow",
				log.Fields{"device-id": device.Id,
					"flowToRemove": flow})
			err := dh.flowMgr.RemoveFlow(ctx, flow)
			if err != nil {
				errorsList = append(errorsList, err)
			}

			dh.decrementActiveFlowRemoveCount(ctx, flow)
		}

		for _, flow := range flows.ToAdd.Items {
			logger.Debugw(ctx, "adding-flow",
				log.Fields{"device-id": device.Id,
					"flowToAdd": flow})
			// If there are active Flow Remove in progress for a given subscriber, wait until it completes
			dh.waitForFlowRemoveToFinish(ctx, flow)
			err := dh.flowMgr.AddFlow(ctx, flow, flowMetadata)
			if err != nil {
				errorsList = append(errorsList, err)
			}
		}
	}

	if groups != nil {
		for _, group := range groups.ToAdd.Items {
			err := dh.flowMgr.AddGroup(ctx, group)
			if err != nil {
				errorsList = append(errorsList, err)
			}
		}
		for _, group := range groups.ToUpdate.Items {
			err := dh.flowMgr.ModifyGroup(ctx, group)
			if err != nil {
				errorsList = append(errorsList, err)
			}
		}
		for _, group := range groups.ToRemove.Items {
			err := dh.flowMgr.DeleteGroup(ctx, group)
			if err != nil {
				errorsList = append(errorsList, err)
			}
		}
	}
	if len(errorsList) > 0 {
		return fmt.Errorf("errors-installing-flows-groups, errors:%v", errorsList)
	}
	logger.Debugw(ctx, "updated-flows-incrementally-successfully", log.Fields{"device-id": dh.device.Id})
	return nil
}

//DisableDevice disables the given device
//It marks the following for the given device:
//Device-Handler Admin-State : down
//Device Port-State: UNKNOWN
//Device Oper-State: UNKNOWN
func (dh *DeviceHandler) DisableDevice(ctx context.Context, device *voltha.Device) error {
	/* On device disable ,admin state update has to be done prior sending request to agent since
	   the indication thread may processes invalid  indications of ONU and OLT*/
	if dh.Client != nil {
		if _, err := dh.Client.DisableOlt(log.WithSpanFromContext(context.Background(), ctx), new(oop.Empty)); err != nil {
			if e, ok := status.FromError(err); ok && e.Code() == codes.Internal {
				return olterrors.NewErrAdapter("olt-disable-failed", log.Fields{"device-id": device.Id}, err)
			}
		}
	}
	L2oamDisableOlt(ctx, dh)
	logger.Debugw(ctx, "olt-disabled", log.Fields{"device-id": device.Id})
	/* Discovered ONUs entries need to be cleared , since on device disable the child devices goes to
	UNREACHABLE state which needs to be configured again*/

	dh.discOnus = sync.Map{}
	dh.onus = sync.Map{}

	dh.stopCollector <- true

	go dh.notifyChildDevices(ctx, "unreachable")
	cloned := proto.Clone(device).(*voltha.Device)
	dh.device = cloned
	if err := dh.coreProxy.PortsStateUpdate(log.WithSpanFromContext(context.TODO(), ctx), cloned.Id, ^uint32(1<<voltha.Port_PON_OLT), voltha.OperStatus_UNKNOWN); err != nil {
		return olterrors.NewErrAdapter("ports-state-update-failed", log.Fields{"device-id": device.Id}, err)
	}
	logger.Debugw(ctx, "disable-device-end", log.Fields{"device-id": device.Id})
	return nil
}

func (dh *DeviceHandler) notifyChildDevices(ctx context.Context, state string) {
	onuInd := oop.OnuIndication{}
	onuInd.OperState = state
	onuDevices, err := dh.coreProxy.GetChildDevices(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id)
	if err != nil {
		logger.Errorw(ctx, "failed-to-get-child-devices-information", log.Fields{"device-id": dh.device.Id, "error": err})
	}
	if onuDevices != nil {
		for _, onuDevice := range onuDevices.Items {
			err := dh.AdapterProxy.SendInterAdapterMessage(log.WithSpanFromContext(context.TODO(), ctx), &onuInd, ic.InterAdapterMessageType_ONU_IND_REQUEST,
				"openolt", onuDevice.Type, onuDevice.Id, onuDevice.ProxyAddress.DeviceId, "")
			if err != nil {
				logger.Errorw(ctx, "failed-to-send-inter-adapter-message", log.Fields{"OnuInd": onuInd,
					"From Adapter": "openolt", "DeviceType": onuDevice.Type, "device-id": onuDevice.Id})
			}

		}
	}

}

//ReenableDevice re-enables the olt device after disable
//It marks the following for the given device:
//Device-Handler Admin-State : up
//Device Port-State: ACTIVE
//Device Oper-State: ACTIVE
func (dh *DeviceHandler) ReenableDevice(ctx context.Context, device *voltha.Device) error {
	if _, err := dh.Client.ReenableOlt(log.WithSpanFromContext(context.Background(), ctx), new(oop.Empty)); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.Internal {
			return olterrors.NewErrAdapter("olt-reenable-failed", log.Fields{"device-id": dh.device.Id}, err)
		}
	}
	logger.Debug(ctx, "olt-reenabled")


	ports, err := dh.coreProxy.ListDevicePorts(ctx, device.Id)
	if err != nil {
		return olterrors.NewErrAdapter("list-ports-failed", log.Fields{"device": device.Id}, err)
	}
	if err := dh.disableAdminDownPorts(ctx, ports); err != nil {
		return olterrors.NewErrAdapter("port-status-update-failed-after-olt-reenable", log.Fields{"device": device}, err)
	}
	device.OperStatus = voltha.OperStatus_ACTIVE
	dh.device = device

	if err := dh.coreProxy.DeviceStateUpdate(log.WithSpanFromContext(context.TODO(), ctx), device.Id, device.ConnectStatus, device.OperStatus); err != nil {
		return olterrors.NewErrAdapter("state-update-failed", log.Fields{
			"device-id":      device.Id,
			"connect-status": device.ConnectStatus,
			"oper-status":    device.OperStatus}, err)
	}

	logger.Debugw(ctx, "reenabledevice-end", log.Fields{"device-id": device.Id})

	return nil
}

func (dh *DeviceHandler) clearUNIData(ctx context.Context, onu *rsrcMgr.OnuGemInfo) error {
	var uniID uint32
	var err error
	for _, port := range onu.UniPorts {
		uniID = UniIDFromPortNum(uint32(port))
		logger.Debugw(ctx, "clearing-resource-data-for-uni-port", log.Fields{"port": port, "uni-id": uniID})
		/* Delete tech-profile instance from the KV store */
		if err = dh.flowMgr.DeleteTechProfileInstances(ctx, onu.IntfID, onu.OnuID, uniID, onu.SerialNumber); err != nil {
			logger.Debugw(ctx, "failed-to-remove-tech-profile-instance-for-onu", log.Fields{"onu-id": onu.OnuID})
		}
		logger.Debugw(ctx, "deleted-tech-profile-instance-for-onu", log.Fields{"onu-id": onu.OnuID})
		flowIDs := dh.resourceMgr.GetCurrentFlowIDsForOnu(ctx, onu.IntfID, int32(onu.OnuID), int32(uniID))
		for _, flowID := range flowIDs {
			dh.resourceMgr.FreeFlowID(ctx, onu.IntfID, int32(onu.OnuID), int32(uniID), flowID)
		}
		tpIDList := dh.resourceMgr.GetTechProfileIDForOnu(ctx, onu.IntfID, onu.OnuID, uniID)
		for _, tpID := range tpIDList {
			if err = dh.resourceMgr.RemoveMeterIDForOnu(ctx, "upstream", onu.IntfID, onu.OnuID, uniID, tpID); err != nil {
				logger.Debugw(ctx, "failed-to-remove-meter-id-for-onu-upstream", log.Fields{"onu-id": onu.OnuID})
			}
			logger.Debugw(ctx, "removed-meter-id-for-onu-upstream", log.Fields{"onu-id": onu.OnuID})
			if err = dh.resourceMgr.RemoveMeterIDForOnu(ctx, "downstream", onu.IntfID, onu.OnuID, uniID, tpID); err != nil {
				logger.Debugw(ctx, "failed-to-remove-meter-id-for-onu-downstream", log.Fields{"onu-id": onu.OnuID})
			}
			logger.Debugw(ctx, "removed-meter-id-for-onu-downstream", log.Fields{"onu-id": onu.OnuID})
		}
		dh.resourceMgr.FreePONResourcesForONU(ctx, onu.IntfID, onu.OnuID, uniID)
		if err = dh.resourceMgr.RemoveTechProfileIDsForOnu(ctx, onu.IntfID, onu.OnuID, uniID); err != nil {
			logger.Debugw(ctx, "failed-to-remove-tech-profile-id-for-onu", log.Fields{"onu-id": onu.OnuID})
		}
		logger.Debugw(ctx, "removed-tech-profile-id-for-onu", log.Fields{"onu-id": onu.OnuID})
		if err = dh.resourceMgr.DelGemPortPktInOfAllServices(ctx, onu.IntfID, onu.OnuID, uint32(port)); err != nil {
			logger.Debugw(ctx, "failed-to-remove-gemport-pkt-in", log.Fields{"intfid": onu.IntfID, "onuid": onu.OnuID, "uniId": uniID})
		}
	}
	return nil
}

func (dh *DeviceHandler) clearNNIData(ctx context.Context) error {
	nniUniID := -1
	nniOnuID := -1

	if dh.resourceMgr == nil {
		return olterrors.NewErrNotFound("resource-manager", log.Fields{"device-id": dh.device.Id}, nil)
	}
	nni, err := dh.resourceMgr.GetNNIFromKVStore(ctx)
	if err != nil {
		return olterrors.NewErrPersistence("get", "nni", 0, nil, err)
	}
	logger.Debugw(ctx, "nni-", log.Fields{"nni": nni})
	for _, nniIntfID := range nni {
		flowIDs := dh.resourceMgr.GetCurrentFlowIDsForOnu(ctx, uint32(nniIntfID), int32(nniOnuID), int32(nniUniID))
		logger.Debugw(ctx, "current-flow-ids-for-nni", log.Fields{"flow-ids": flowIDs})
		for _, flowID := range flowIDs {
			dh.resourceMgr.FreeFlowID(ctx, uint32(nniIntfID), -1, -1, uint32(flowID))
		}
		dh.resourceMgr.RemoveResourceMap(ctx, nniIntfID, int32(nniOnuID), int32(nniUniID))
	}
	if err = dh.resourceMgr.DelNNiFromKVStore(ctx); err != nil {
		return olterrors.NewErrPersistence("clear", "nni", 0, nil, err)
	}
	return nil
}

// DeleteDevice deletes the device instance from openolt handler array.  Also clears allocated resource manager resources.  Also reboots the OLT hardware!
func (dh *DeviceHandler) DeleteDevice(ctx context.Context, device *voltha.Device) error {
	logger.Debug(ctx, "function-entry-delete-device")
	/* Clear the KV store data associated with the all the UNI ports
	   This clears up flow data and also resource map data for various
	   other pon resources like alloc_id and gemport_id
	*/
	L2oamDeleteOlt(ctx, dh)
	go dh.cleanupDeviceResources(ctx)
	logger.Debug(ctx, "removed-device-from-Resource-manager-KV-store")
	dh.stopCollector <- true
	dh.stopHeartbeatCheck <- true
	if dh.Client != nil {
		if _, err := dh.Client.Reboot(ctx, new(oop.Empty)); err != nil {
			return olterrors.NewErrAdapter("olt-reboot-failed", log.Fields{"device-id": dh.device.Id}, err).Log()
		}
	}
	cloned := proto.Clone(device).(*voltha.Device)
	cloned.OperStatus = voltha.OperStatus_UNKNOWN
	cloned.ConnectStatus = voltha.ConnectStatus_UNREACHABLE
	if err := dh.coreProxy.DeviceStateUpdate(ctx, cloned.Id, cloned.ConnectStatus, cloned.OperStatus); err != nil {
		return olterrors.NewErrAdapter("device-state-update-failed", log.Fields{
			"device-id":      device.Id,
			"connect-status": cloned.ConnectStatus,
			"oper-status":    cloned.OperStatus}, err).Log()
	}
	return nil
}
func (dh *DeviceHandler) cleanupDeviceResources(ctx context.Context) {

	if dh.resourceMgr != nil {
		noOfPonPorts := dh.resourceMgr.DevInfo.GetPonPorts()
		var ponPort uint32
		for ponPort = 0; ponPort < noOfPonPorts; ponPort++ {
			var onuGemData []rsrcMgr.OnuGemInfo
			err := dh.resourceMgr.ResourceMgrs[ponPort].GetOnuGemInfo(ctx, ponPort, &onuGemData)
			if err != nil {
				_ = olterrors.NewErrNotFound("onu", log.Fields{
					"device-id": dh.device.Id,
					"pon-port":  ponPort}, err).Log()
			}
			for _, onu := range onuGemData {
				onuID := make([]uint32, 1)
				logger.Debugw(ctx, "onu-data", log.Fields{"onu": onu})
				if err = dh.clearUNIData(ctx, &onu); err != nil {
					logger.Errorw(ctx, "failed-to-clear-data-for-onu", log.Fields{"onu-device": onu})
				}
				// Clear flowids for gem cache.
				for _, gem := range onu.GemPorts {
					dh.resourceMgr.DeleteFlowIDsForGem(ctx, ponPort, gem)
				}
				onuID[0] = onu.OnuID
				dh.resourceMgr.FreeonuID(ctx, ponPort, onuID)
			}
			dh.resourceMgr.DeleteIntfIDGempMapPath(ctx, ponPort)
			onuGemData = nil
			err = dh.resourceMgr.DelOnuGemInfoForIntf(ctx, ponPort)
			if err != nil {
				logger.Errorw(ctx, "failed-to-update-onugem-info", log.Fields{"intfid": ponPort, "onugeminfo": onuGemData})
			}
		}
		/* Clear the flows from KV store associated with NNI port.
		   There are mostly trap rules from NNI port (like LLDP)
		*/
		if err := dh.clearNNIData(ctx); err != nil {
			logger.Errorw(ctx, "failed-to-clear-data-for-NNI-port", log.Fields{"device-id": dh.device.Id})
		}

		/* Clear the resource pool for each PON port in the background */
		go func() {
			if err := dh.resourceMgr.Delete(ctx); err != nil {
				logger.Debug(ctx, err)
			}
		}()
	}

	/*Delete ONU map for the device*/
	dh.onus.Range(func(key interface{}, value interface{}) bool {
		dh.onus.Delete(key)
		return true
	})

	/*Delete discovered ONU map for the device*/
	dh.discOnus.Range(func(key interface{}, value interface{}) bool {
		dh.discOnus.Delete(key)
		return true
	})
}

//RebootDevice reboots the given device
func (dh *DeviceHandler) RebootDevice(ctx context.Context, device *voltha.Device) error {
	L2oamRebootDevice(ctx, dh, device)
	logger.Debugw(ctx, "rebooted-device-successfully", log.Fields{"device-id": device.Id})
	return nil
}

func (dh *DeviceHandler) handlePacketIndication(ctx context.Context, packetIn *oop.PacketIndication) error {
	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "received-packet-in", log.Fields{
			"packet-indication": *packetIn,
			"device-id":         dh.device.Id,
			"packet":            hex.EncodeToString(packetIn.Pkt),
		})
	}
	logicalPortNum, err := dh.flowMgr.GetLogicalPortFromPacketIn(ctx, packetIn)
	if err != nil {
		return olterrors.NewErrNotFound("logical-port", log.Fields{"packet": hex.EncodeToString(packetIn.Pkt)}, err)
	}
	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "sending-packet-in-to-core", log.Fields{
			"logical-port-num": logicalPortNum,
			"device-id":        dh.device.Id,
			"packet":           hex.EncodeToString(packetIn.Pkt),
		})
	}

	if err := dh.coreProxy.SendPacketIn(log.WithSpanFromContext(context.TODO(), ctx), dh.device.Id, logicalPortNum, packetIn.Pkt); err != nil {
		return olterrors.NewErrCommunication("send-packet-in", log.Fields{
			"destination": "core",
			"source":      dh.device.Type,
			"device-id":   dh.device.Id,
			"packet":      hex.EncodeToString(packetIn.Pkt),
		}, err)
	}

	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "success-sending-packet-in-to-core!", log.Fields{
			"packet":    hex.EncodeToString(packetIn.Pkt),
			"device-id": dh.device.Id,
		})
	}
	return nil
}

// PacketOut sends packet-out from VOLTHA to OLT on the egress port provided
func (dh *DeviceHandler) PacketOut(ctx context.Context, egressPortNo int, packet *of.OfpPacketOut) error {
	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "incoming-packet-out", log.Fields{
			"device-id":      dh.device.Id,
			"egress-port-no": egressPortNo,
			"pkt-length":     len(packet.Data),
			"packet":         hex.EncodeToString(packet.Data),
		})
	}

	if disableBBSim {
		if vlanEnable {
			byteString := string(packet.Data)
			searchByteString := string([]byte{0x88, 0x8e, 0x03, 0x00, 0x00, 0x04, 0x03})
			//if strings.Index(byteString, searchByteString) != -1 {
			if strings.Contains(byteString, searchByteString) {
				device := FindL2oamDevice(hex.EncodeToString(packet.Data[:6]))
				if onu, ok := device.(*L2oamOnuDevice); ok {
					onu.Base.EapFlag = true
				}
			}

			// add s-vlan tag
			sendPacket := make([]byte, len(packet.Data)+4)
			copy(sendPacket, packet.Data[:12])
			copy(sendPacket[12:], svlanTagAuth)
			copy(sendPacket[16:], packet.Data[12:])
			GetL2oamHandle().send(sendPacket)
		} else {
			GetL2oamHandle().send(packet.Data)
		}
		return nil
	}

	egressPortType := IntfIDToPortTypeName(uint32(egressPortNo))
	if egressPortType == voltha.Port_ETHERNET_UNI {
		outerEthType := (uint16(packet.Data[12]) << 8) | uint16(packet.Data[13])
		innerEthType := (uint16(packet.Data[16]) << 8) | uint16(packet.Data[17])
		if outerEthType == 0x8942 || outerEthType == 0x88cc {
			// Do not packet-out lldp packets on uni port.
			// ONOS has no clue about uni/nni ports, it just packets out on all
			// available ports on the Logical Switch. It should not be interested
			// in the UNI links.
			logger.Debugw(ctx, "dropping-lldp-packet-out-on-uni", log.Fields{
				"device-id": dh.device.Id,
			})
			return nil
		}
		if outerEthType == 0x88a8 || outerEthType == 0x8100 {
			if innerEthType == 0x8100 {
				// q-in-q 802.1ad or 802.1q double tagged packet.
				// slice out the outer tag.
				packet.Data = append(packet.Data[:12], packet.Data[16:]...)
				if logger.V(log.DebugLevel) {
					logger.Debugw(ctx, "packet-now-single-tagged", log.Fields{
						"packet-data": hex.EncodeToString(packet.Data),
						"device-id":   dh.device.Id,
					})
				}
			}
		}
		intfID := IntfIDFromUniPortNum(uint32(egressPortNo))
		onuID := OnuIDFromPortNum(uint32(egressPortNo))
		uniID := UniIDFromPortNum(uint32(egressPortNo))

		gemPortID, err := dh.flowMgr.GetPacketOutGemPortID(ctx, intfID, onuID, uint32(egressPortNo), packet.Data)
		if err != nil {
			// In this case the openolt agent will receive the gemPortID as 0.
			// The agent tries to retrieve the gemPortID in this case.
			// This may not always succeed at the agent and packetOut may fail.
			logger.Errorw(ctx, "failed-to-retrieve-gemport-id-for-packet-out", log.Fields{
				"intf-id":   intfID,
				"onu-id":    onuID,
				"uni-id":    uniID,
				"packet":    hex.EncodeToString(packet.Data),
				"device-id": dh.device.Id,
			})
		}

		onuPkt := oop.OnuPacket{IntfId: intfID, OnuId: onuID, PortNo: uint32(egressPortNo), GemportId: gemPortID, Pkt: packet.Data}

		if logger.V(log.DebugLevel) {
			logger.Debugw(ctx, "sending-packet-to-onu", log.Fields{
				"egress-port-no": egressPortNo,
				"intf-id":        intfID,
				"onu-id":         onuID,
				"uni-id":         uniID,
				"gem-port-id":    gemPortID,
				"packet":         hex.EncodeToString(packet.Data),
				"device-id":      dh.device.Id,
			})
		}

		if _, err := dh.Client.OnuPacketOut(ctx, &onuPkt); err != nil {
			return olterrors.NewErrCommunication("packet-out-send", log.Fields{
				"source":             "adapter",
				"destination":        "onu",
				"egress-port-number": egressPortNo,
				"intf-id":            intfID,
				"oni-id":             onuID,
				"uni-id":             uniID,
				"gem-port-id":        gemPortID,
				"packet":             hex.EncodeToString(packet.Data),
				"device-id":          dh.device.Id,
			}, err)
		}
	} else if egressPortType == voltha.Port_ETHERNET_NNI {
		nniIntfID, err := IntfIDFromNniPortNum(ctx, uint32(egressPortNo))
		if err != nil {
			return olterrors.NewErrInvalidValue(log.Fields{
				"egress-nni-port": egressPortNo,
				"device-id":       dh.device.Id,
			}, err)
		}
		uplinkPkt := oop.UplinkPacket{IntfId: nniIntfID, Pkt: packet.Data}

		if logger.V(log.DebugLevel) {
			logger.Debugw(ctx, "sending-packet-to-nni", log.Fields{
				"uplink-pkt": uplinkPkt,
				"packet":     hex.EncodeToString(packet.Data),
				"device-id":  dh.device.Id,
			})
		}

		if _, err := dh.Client.UplinkPacketOut(ctx, &uplinkPkt); err != nil {
			return olterrors.NewErrCommunication("packet-out-to-nni", log.Fields{
				"packet":    hex.EncodeToString(packet.Data),
				"device-id": dh.device.Id,
			}, err)
		}
	} else {
		logger.Warnw(ctx, "packet-out-to-this-interface-type-not-implemented", log.Fields{
			"egress-port-no": egressPortNo,
			"egressPortType": egressPortType,
			"packet":         hex.EncodeToString(packet.Data),
			"device-id":      dh.device.Id,
		})
	}
	return nil
}

func (dh *DeviceHandler) formOnuKey(intfID, onuID uint32) string {
	return "" + strconv.Itoa(int(intfID)) + "." + strconv.Itoa(int(onuID))
}

func startHeartbeatCheck(ctx context.Context, dh *DeviceHandler) {

	var timerCheck *time.Timer

	for {
		heartbeatTimer := time.NewTimer(dh.openOLT.HeartbeatCheckInterval)
		select {
		case <-heartbeatTimer.C:
			ctxWithTimeout, cancel := context.WithTimeout(log.WithSpanFromContext(context.Background(), ctx), dh.openOLT.GrpcTimeoutInterval)
			if heartBeat, err := dh.Client.HeartbeatCheck(ctxWithTimeout, new(oop.Empty)); err != nil {
				logger.Warnw(ctx, "hearbeat-failed", log.Fields{"device-id": dh.device.Id})
				if timerCheck == nil {
					// start a after func, when expired will update the state to the core
					timerCheck = time.AfterFunc(dh.openOLT.HeartbeatFailReportInterval, func() { dh.updateStateUnreachable(ctx) })
				}
			} else {
				if timerCheck != nil {
					if timerCheck.Stop() {
						logger.Debugw(ctx, "got-hearbeat-within-timeout", log.Fields{"device-id": dh.device.Id})
					}
					timerCheck = nil
				}
				logger.Debugw(ctx, "hearbeat",
					log.Fields{"signature": heartBeat,
						"device-id": dh.device.Id})
			}
			cancel()
		case <-dh.stopHeartbeatCheck:
			logger.Debugw(ctx, "stopping-heart-beat-check", log.Fields{"device-id": dh.device.Id})
			return
		}
	}
}

func (dh *DeviceHandler) updateStateUnreachable(ctx context.Context) {
	device, err := dh.coreProxy.GetDevice(ctx, dh.device.Id, dh.device.Id)
	if err != nil || device == nil {
		_ = olterrors.NewErrNotFound("device", log.Fields{"device-id": dh.device.Id}, err).Log()
	}

	if device.ConnectStatus == voltha.ConnectStatus_REACHABLE {
		if err = dh.coreProxy.DeviceStateUpdate(ctx, dh.device.Id, voltha.ConnectStatus_UNREACHABLE, voltha.OperStatus_UNKNOWN); err != nil {
			_ = olterrors.NewErrAdapter("device-state-update-failed", log.Fields{"device-id": dh.device.Id}, err).LogAt(log.ErrorLevel)
		}
		if err = dh.coreProxy.PortsStateUpdate(ctx, dh.device.Id, 0, voltha.OperStatus_UNKNOWN); err != nil {
			_ = olterrors.NewErrAdapter("port-update-failed", log.Fields{"device-id": dh.device.Id}, err).Log()
		}
		go dh.cleanupDeviceResources(ctx)

		dh.lockDevice.RLock()
		// Stop the read indication only if it the routine is active
		// The read indication would have already stopped due to failure on the gRPC stream following OLT going unreachable
		// Sending message on the 'stopIndication' channel again will cause the readIndication routine to immediately stop
		// on next execution of the readIndication routine.
		if dh.isReadIndicationRoutineActive {
			dh.stopIndications <- true
		}
		dh.lockDevice.RUnlock()

		dh.transitionMap.Handle(ctx, DeviceInit)

	}
}

// EnablePort to enable Pon interface
func (dh *DeviceHandler) EnablePort(ctx context.Context, port *voltha.Port) error {
	logger.Debugw(ctx, "enable-port", log.Fields{"Device": dh.device, "port": port})
	return dh.modifyPhyPort(ctx, port, true)
}

// DisablePort to disable pon interface
func (dh *DeviceHandler) DisablePort(ctx context.Context, port *voltha.Port) error {
	logger.Debugw(ctx, "disable-port", log.Fields{"Device": dh.device, "port": port})
	return dh.modifyPhyPort(ctx, port, false)
}

//modifyPhyPort is common function to enable and disable the port. parm :enablePort, true to enablePort and false to disablePort.
func (dh *DeviceHandler) modifyPhyPort(ctx context.Context, port *voltha.Port, enablePort bool) error {
	logger.Infow(ctx, "modifyPhyPort", log.Fields{"port": port, "Enable": enablePort, "device-id": dh.device.Id})
	if port.GetType() == voltha.Port_ETHERNET_NNI {
		// Bug is opened for VOL-2505 to support NNI disable feature.
		logger.Infow(ctx, "voltha-supports-single-nni-hence-disable-of-nni-not-allowed",
			log.Fields{"device": dh.device, "port": port})
		return olterrors.NewErrAdapter("illegal-port-request", log.Fields{
			"port-type":    port.GetType,
			"enable-state": enablePort}, nil)
	}
	ponID := PortNoToIntfID(port.GetPortNo(), voltha.Port_PON_OLT)
	ponIntf := &oop.Interface{IntfId: ponID}
	var operStatus voltha.OperStatus_Types
	if enablePort {
		operStatus = voltha.OperStatus_ACTIVE
		out, err := dh.Client.EnablePonIf(ctx, ponIntf)

		if err != nil {
			return olterrors.NewErrAdapter("pon-port-enable-failed", log.Fields{
				"device-id": dh.device.Id,
				"port":      port}, err)
		}
		// updating interface local cache for collecting stats
		dh.activePorts.Store(ponID, true)
		logger.Infow(ctx, "enabled-pon-port", log.Fields{"out": out, "device-id": dh.device, "Port": port})
	} else {
		operStatus = voltha.OperStatus_UNKNOWN
		out, err := dh.Client.DisablePonIf(ctx, ponIntf)
		if err != nil {
			return olterrors.NewErrAdapter("pon-port-disable-failed", log.Fields{
				"device-id": dh.device.Id,
				"port":      port}, err)
		}
		// updating interface local cache for collecting stats
		dh.activePorts.Store(ponID, false)
		logger.Infow(ctx, "disabled-pon-port", log.Fields{"out": out, "device-id": dh.device, "Port": port})
	}
	if err := dh.coreProxy.PortStateUpdate(ctx, dh.device.Id, voltha.Port_PON_OLT, port.PortNo, operStatus); err != nil {
		return olterrors.NewErrAdapter("port-state-update-failed", log.Fields{
			"device-id": dh.device.Id,
			"port":      port.PortNo}, err)
	}
	return nil
}

//disableAdminDownPorts disables the ports, if the corresponding port Adminstate is disabled on reboot and Renable device.
func (dh *DeviceHandler) disableAdminDownPorts(ctx context.Context, ports []*voltha.Port) error {
	for _, port := range ports {
		if port.AdminState == common.AdminState_DISABLED {
			if err := dh.DisablePort(ctx, port); err != nil {
				return olterrors.NewErrAdapter("port-disable-failed", log.Fields{
					"device-id": dh.device.Id,
					"port":      port}, err)
			}
		}
	}
	return nil
}

//populateActivePorts to populate activePorts map
func (dh *DeviceHandler) populateActivePorts(ctx context.Context, ports []*voltha.Port) {
	logger.Infow(ctx, "populateActivePorts", log.Fields{"device-id": dh.device.Id})
	for _, port := range ports {
		if port.Type == voltha.Port_ETHERNET_NNI {
			if port.OperStatus == voltha.OperStatus_ACTIVE {
				dh.activePorts.Store(PortNoToIntfID(port.PortNo, voltha.Port_ETHERNET_NNI), true)
			} else {
				dh.activePorts.Store(PortNoToIntfID(port.PortNo, voltha.Port_ETHERNET_NNI), false)
			}
		}
		if port.Type == voltha.Port_PON_OLT {
			if port.OperStatus == voltha.OperStatus_ACTIVE {
				dh.activePorts.Store(PortNoToIntfID(port.PortNo, voltha.Port_PON_OLT), true)
			} else {
				dh.activePorts.Store(PortNoToIntfID(port.PortNo, voltha.Port_PON_OLT), false)
			}
		}
	}
}

// ChildDeviceLost deletes ONU and clears pon resources related to it.
func (dh *DeviceHandler) ChildDeviceLost(ctx context.Context, pPortNo uint32, onuID uint32) error {
	logger.Debugw(ctx, "child-device-lost", log.Fields{"pdeviceID": dh.device.Id})
	intfID := PortNoToIntfID(pPortNo, voltha.Port_PON_OLT)
	onuKey := dh.formOnuKey(intfID, onuID)
	onuDevice, ok := dh.onus.Load(onuKey)
	if !ok {
		return olterrors.NewErrAdapter("failed-to-load-onu-details",
			log.Fields{
				"device-id": dh.device.Id,
				"onu-id":    onuID,
				"intf-id":   intfID}, nil).Log()
	}
	var sn *oop.SerialNumber
	var err error
	if sn, err = dh.deStringifySerialNumber(onuDevice.(*OnuDevice).serialNumber); err != nil {
		return olterrors.NewErrAdapter("failed-to-destringify-serial-number",
			log.Fields{
				"devicer-id":    dh.device.Id,
				"serial-number": onuDevice.(*OnuDevice).serialNumber}, err).Log()
	}

	for uniID := 0; uniID < MaxUnisPerOnu; uniID++ {
		var flowRemoveData pendingFlowRemoveData
		key := pendingFlowRemoveDataKey{intfID: intfID, onuID: onuID, uniID: uint32(uniID)}
		dh.lockDevice.RLock()
		if flowRemoveData, ok = dh.pendingFlowRemoveDataPerSubscriber[key]; !ok {
			dh.lockDevice.RUnlock()
			continue
		}
		dh.lockDevice.RUnlock()

		logger.Debugw(ctx, "wait-for-flow-remove-complete-before-processing-child-device-lost",
			log.Fields{"int-id": intfID, "onu-id": onuID, "uni-id": uniID})
		// Wait for all flow removes to finish first
		<-flowRemoveData.allFlowsRemoved
		logger.Debugw(ctx, "flow-removes-complete-for-subscriber",
			log.Fields{"int-id": intfID, "onu-id": onuID, "uni-id": uniID})
	}

	onu := &oop.Onu{IntfId: intfID, OnuId: onuID, SerialNumber: sn}
	if _, err := dh.Client.DeleteOnu(log.WithSpanFromContext(context.Background(), ctx), onu); err != nil {
		return olterrors.NewErrAdapter("failed-to-delete-onu", log.Fields{
			"device-id": dh.device.Id,
			"onu-id":    onuID}, err).Log()
	}
	var onuGemData []rsrcMgr.OnuGemInfo
	if onuMgr, ok := dh.resourceMgr.ResourceMgrs[intfID]; !ok {
		logger.Warnw(ctx, "failed-to-get-resource-manager-for-interface-Id", log.Fields{
			"device-id": dh.device.Id,
			"intf-id":   intfID})
	} else {
		if err := onuMgr.GetOnuGemInfo(ctx, intfID, &onuGemData); err != nil {
			logger.Warnw(ctx, "failed-to-get-onu-info-for-pon-port", log.Fields{
				"device-id": dh.device.Id,
				"intf-id":   intfID,
				"error":     err})
		} else {
			for i, onu := range onuGemData {
				if onu.OnuID == onuID && onu.SerialNumber == onuDevice.(*OnuDevice).serialNumber {
					logger.Debugw(ctx, "onu-data", log.Fields{"onu": onu})
					if err := dh.clearUNIData(ctx, &onu); err != nil {
						logger.Warnw(ctx, "failed-to-clear-uni-data-for-onu", log.Fields{
							"device-id":  dh.device.Id,
							"onu-device": onu,
							"error":      err})
					}
					// Clear flowids for gem cache.
					for _, gem := range onu.GemPorts {
						dh.resourceMgr.DeleteFlowIDsForGem(ctx, intfID, gem)
					}
					onuGemData = append(onuGemData[:i], onuGemData[i+1:]...)
					err := onuMgr.AddOnuGemInfo(ctx, intfID, onuGemData)
					if err != nil {
						logger.Warnw(ctx, "persistence-update-onu-gem-info-failed", log.Fields{
							"intf-id":    intfID,
							"onu-device": onu,
							"onu-gem":    onuGemData,
							"error":      err})
						//Not returning error on cleanup.
					}
					logger.Debugw(ctx, "removed-onu-gem-info", log.Fields{"intf": intfID, "onu-device": onu, "onugem": onuGemData})
					dh.resourceMgr.FreeonuID(ctx, intfID, []uint32{onu.OnuID})
					break
				}
			}
		}
	}
	dh.onus.Delete(onuKey)
	dh.discOnus.Delete(onuDevice.(*OnuDevice).serialNumber)
	return nil
}

func getInPortFromFlow(flow *of.OfpFlowStats) uint32 {
	for _, field := range flows.GetOfbFields(flow) {
		if field.Type == flows.IN_PORT {
			return field.GetPort()
		}
	}
	return InvalidPort
}

func getOutPortFromFlow(flow *of.OfpFlowStats) uint32 {
	for _, action := range flows.GetActions(flow) {
		if action.Type == flows.OUTPUT {
			if out := action.GetOutput(); out != nil {
				return out.GetPort()
			}
		}
	}
	return InvalidPort
}

func (dh *DeviceHandler) incrementActiveFlowRemoveCount(ctx context.Context, flow *of.OfpFlowStats) {
	inPort, outPort := getPorts(flow)
	logger.Debugw(ctx, "increment-flow-remove-count-for-inPort-out-port", log.Fields{"inPort": inPort, "out-port": outPort})
	if inPort != InvalidPort && outPort != InvalidPort {
		_, intfID, onuID, uniID := ExtractAccessFromFlow(inPort, outPort)
		key := pendingFlowRemoveDataKey{intfID: intfID, onuID: onuID, uniID: uniID}
		logger.Debugw(ctx, "increment-flow-remove-count-for-subscriber", log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID})

		dh.lockDevice.Lock()
		defer dh.lockDevice.Unlock()
		flowRemoveData, ok := dh.pendingFlowRemoveDataPerSubscriber[key]
		if !ok {
			flowRemoveData = pendingFlowRemoveData{
				pendingFlowRemoveCount: 0,
				allFlowsRemoved:        make(chan struct{}),
			}
		}
		flowRemoveData.pendingFlowRemoveCount++
		dh.pendingFlowRemoveDataPerSubscriber[key] = flowRemoveData

		logger.Debugw(ctx, "current-flow-remove-count窶妬ncrement",
			log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID,
				"currCnt": dh.pendingFlowRemoveDataPerSubscriber[key].pendingFlowRemoveCount})
	}
}

func (dh *DeviceHandler) decrementActiveFlowRemoveCount(ctx context.Context, flow *of.OfpFlowStats) {
	inPort, outPort := getPorts(flow)
	logger.Debugw(ctx, "decrement-flow-remove-count-for-inPort-out-port", log.Fields{"inPort": inPort, "out-port": outPort})
	if inPort != InvalidPort && outPort != InvalidPort {
		_, intfID, onuID, uniID := ExtractAccessFromFlow(uint32(inPort), uint32(outPort))
		key := pendingFlowRemoveDataKey{intfID: intfID, onuID: onuID, uniID: uniID}
		logger.Debugw(ctx, "decrement-flow-remove-count-for-subscriber", log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID})

		dh.lockDevice.Lock()
		defer dh.lockDevice.Unlock()
		if val, ok := dh.pendingFlowRemoveDataPerSubscriber[key]; !ok {
			logger.Fatalf(ctx, "flow-remove-key-not-found", log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID})
		} else {
			if val.pendingFlowRemoveCount > 0 {
				val.pendingFlowRemoveCount--
			}
			logger.Debugw(ctx, "current-flow-remove-count-after-decrement",
				log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID,
					"currCnt": dh.pendingFlowRemoveDataPerSubscriber[key].pendingFlowRemoveCount})
			// If all flow removes have finished, then close the channel to signal the receiver
			// to go ahead with flow adds.
			if val.pendingFlowRemoveCount == 0 {
				close(val.allFlowsRemoved)
				delete(dh.pendingFlowRemoveDataPerSubscriber, key)
				return
			}
			dh.pendingFlowRemoveDataPerSubscriber[key] = val
		}
	}
}

func (dh *DeviceHandler) waitForFlowRemoveToFinish(ctx context.Context, flow *of.OfpFlowStats) {
	var flowRemoveData pendingFlowRemoveData
	var ok bool
	inPort, outPort := getPorts(flow)
	logger.Debugw(ctx, "wait-for-flow-remove-to-finish-for-inPort-out-port", log.Fields{"inPort": inPort, "out-port": outPort})
	if inPort != InvalidPort && outPort != InvalidPort {
		_, intfID, onuID, uniID := ExtractAccessFromFlow(inPort, outPort)
		key := pendingFlowRemoveDataKey{intfID: intfID, onuID: onuID, uniID: uniID}
		logger.Debugw(ctx, "wait-for-flow-remove-to-finish-for-subscriber", log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID})

		dh.lockDevice.RLock()
		if flowRemoveData, ok = dh.pendingFlowRemoveDataPerSubscriber[key]; !ok {
			logger.Debugw(ctx, "no-pending-flow-to-remove", log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID})
			dh.lockDevice.RUnlock()
			return
		}
		dh.lockDevice.RUnlock()

		// Wait for all flow removes to finish first
		<-flowRemoveData.allFlowsRemoved

		logger.Debugw(ctx, "all-flows-cleared--handling-flow-add-now", log.Fields{"intf-id": intfID, "onu-id": onuID, "uni-id": uniID})
	}
}

func getPorts(flow *of.OfpFlowStats) (uint32, uint32) {
	inPort := getInPortFromFlow(flow)
	outPort := getOutPortFromFlow(flow)

	if inPort == InvalidPort || outPort == InvalidPort {
		return inPort, outPort
	}

	if isControllerFlow := IsControllerBoundFlow(outPort); isControllerFlow {
		/* Get UNI port/ IN Port from tunnel ID field for upstream controller bound flows  */
		if portType := IntfIDToPortTypeName(inPort); portType == voltha.Port_PON_OLT {
			if uniPort := flows.GetChildPortFromTunnelId(flow); uniPort != 0 {
				return uniPort, outPort
			}
		}
	} else {
		// Downstream flow from NNI to PON port , Use tunnel ID as new OUT port / UNI port
		if portType := IntfIDToPortTypeName(outPort); portType == voltha.Port_PON_OLT {
			if uniPort := flows.GetChildPortFromTunnelId(flow); uniPort != 0 {
				return inPort, uniPort
			}
			// Upstream flow from PON to NNI port , Use tunnel ID as new IN port / UNI port
		} else if portType := IntfIDToPortTypeName(inPort); portType == voltha.Port_PON_OLT {
			if uniPort := flows.GetChildPortFromTunnelId(flow); uniPort != 0 {
				return uniPort, outPort
			}
		}
	}

	return InvalidPort, InvalidPort
}

func extractOmciTransactionID(omciPkt []byte) uint16 {
	if len(omciPkt) > 3 {
		d := omciPkt[0:2]
		transid := binary.BigEndian.Uint16(d)
		return transid
	}
	return 0
}

// StoreOnuDevice stores the onu parameters to the local cache.
func (dh *DeviceHandler) StoreOnuDevice(onuDevice *OnuDevice) {
	onuKey := dh.formOnuKey(onuDevice.intfID, onuDevice.onuID)
	dh.onus.Store(onuKey, onuDevice)
}

func (dh *DeviceHandler) getExtValue(ctx context.Context, device *voltha.Device, value voltha.ValueType_Type) (*voltha.ReturnValues, error) {
	var err error
	var sn *oop.SerialNumber
	var ID uint32
	resp := new(voltha.ReturnValues)
	valueparam := new(oop.ValueParam)
	ctx = log.WithSpanFromContext(context.Background(), ctx)
	logger.Infow(ctx, "getExtValue", log.Fields{"onu-id": device.Id, "pon-intf": device.ParentPortNo})
	if sn, err = dh.deStringifySerialNumber(device.SerialNumber); err != nil {
		return nil, err
	}
	ID = device.ProxyAddress.GetOnuId()
	Onu := oop.Onu{IntfId: device.ParentPortNo, OnuId: ID, SerialNumber: sn}
	valueparam.Onu = &Onu
	valueparam.Value = value

	resp.Unsupported = uint32(value)
	_ = ctx

	/*
		resp, err = dh.Client.GetValue(ctx, valueparam)
		if err != nil {
			logger.Errorw("error-while-getValue", log.Fields{"DeviceID": dh.device, "onu-id": onuid, "error": err})
			return nil, err
		}
	*/

	logger.Infow(ctx, "get-ext-value", log.Fields{"resp": resp, "device-id": dh.device, "onu-id": device.Id, "pon-intf": device.ParentPortNo})
	return resp, nil
}

// func (dh *DeviceHandler) L2oamCmdRequest2(ctx context.Context, device *voltha.Device, request *voltha.SimulateAlarmRequest) error {
// 	logger.Infow(ctx, "L2oamRequest", log.Fields{"device-id": device.Id})

// 	switch request.Drift {
// 	case 0: // add-flow
// 		L2oamAfterKeepAlive(ctx, dh)
// 		L2oamAddFlow(ctx, dh)

// 	case 1: // add-flow-dev
// 		tomiObjContext := L2oamAddFlowToDeviceDS(ctx, dh)
// 		if tomiObjContext != nil {
// 			L2oamAddFlowToDeviceUS(ctx, dh, tomiObjContext)
// 		}
// 	}

// 	return nil
// }

// L2oamCmdRequest executes some commands for L2OAM
func (dh *DeviceHandler) L2oamCmdRequest(ctx context.Context, device *voltha.Device, request *voltha.SimulateAlarmRequest) error {
	logger.Infow(ctx, "L2oamCmdRequest()", log.Fields{"device-id": device.Id})

	cmd := parseL2oamCmd(request)
	logger.Info(ctx, "parseL2oamCmd() request=%v, cmd=%v", request, cmd)

	switch cmd.Type {
	case "add-flow":
		//L2oamAfterKeepAlive(ctx, dh)
		L2oamAddFlow(ctx, dh, cmd)

	case "add-flow-dev":
		err := L2oamAddFlowToDeviceAll(ctx, dh, cmd)
		if err == nil {
			onu := FindL2oamDeviceByDeviceID(cmd.OnuDeviceID)
			if onu == nil {
				logger.Debug(ctx, fmt.Sprintf("L2oamCmdRequest() FindL2oamDevice() onu not found. deviceId=%s", cmd.OnuDeviceID))
			} else {
				// start ONU mount sequence
				onu.startMountSequence(context.Background(), l2oam.OnuPkgType, cmd)
			}
		}
	default:
		logger.Error(ctx, "L2oamCmdRequest() cmd.Type error. request=%v, cmd=%v", request, cmd)
	}

	return nil
}

// L2oamCmd contains command arguments for L2OAM command
type L2oamCmd struct {
	Type        string
	Cir         []byte
	Pir         []byte
	Tpid        []byte
	Vid         []byte
	Itpid       []byte
	Ivid        []byte
	OnuDeviceID string
}

func parseL2oamCmd(request *voltha.SimulateAlarmRequest) *L2oamCmd {
	l2oamCmd := &L2oamCmd{
		Type:        "unknown",
		Cir:         []byte{0x00, 0x00, 0x03, 0xe8},
		Pir:         []byte{0x00, 0x98, 0x96, 0x80},
		Tpid:        []byte{0x88, 0xa8},
		Vid:         []byte{0x00, 0x64},
		Itpid:       []byte{0x81, 0x00},
		Ivid:        []byte{0x00, 0x0a},
		OnuDeviceID: "",
	}

	if request.Indicator == "" {
		return l2oamCmd
	}
	cmdStr := strings.Replace(request.Indicator, "\"", "", -1)
	if request.Drift == 0 {
		l2oamCmd.Type = "add-flow"
	} else {
		l2oamCmd.Type = "add-flow-dev"
	}

	cmds := strings.Split(cmdStr, ",")
	for _, cmd := range cmds {
		kv := strings.Split(cmd, "=")
		key := kv[0]
		value := kv[1]
		switch key {
		case "cir":
			bytes, _ := hex.DecodeString(value)
			l2oamCmd.Cir = bytes
		case "pir":
			bytes, _ := hex.DecodeString(value)
			l2oamCmd.Pir = bytes
		case "tpid":
			bytes, _ := hex.DecodeString(value)
			l2oamCmd.Tpid = bytes
		case "vid":
			bytes, _ := hex.DecodeString(value)
			l2oamCmd.Vid = bytes
		case "itpid":
			bytes, _ := hex.DecodeString(value)
			l2oamCmd.Itpid = bytes
		case "ivid":
			bytes, _ := hex.DecodeString(value)
			l2oamCmd.Ivid = bytes
		case "onudeviceid":
			l2oamCmd.OnuDeviceID = value
		default:
		}
	}
	return l2oamCmd
}

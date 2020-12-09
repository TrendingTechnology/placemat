package types

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/yaml"
)

// Cluster represents a set of resources for a virtual data center.
type Cluster struct {
	Networks []*NetworkSpec
	NetNSs   []*NetNSSpec
	Nodes    []*NodeSpec
	Images   []*ImageSpec
}

const (
	maxNetworkNameLen = 15
)

// Network types.
const (
	NetworkInternal = "internal"
	NetworkExternal = "external"
	NetworkBMC      = "bmc"
)

// NetworkSpec represents a Network specification in YAML
type NetworkSpec struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	UseNAT  bool   `json:"use-nat"`
	Address string `json:"address,omitempty"`
}

func (n *NetworkSpec) validate() error {
	if len(n.Name) > maxNetworkNameLen {
		return errors.New("too long name: " + n.Name)
	}

	switch n.Type {
	case NetworkInternal:
		if n.UseNAT {
			return errors.New("useNAT must be false for internal network")
		}
		if len(n.Address) > 0 {
			return errors.New("address cannot be specified for internal network")
		}
	case NetworkExternal:
		if len(n.Address) == 0 {
			return errors.New("address must be specified for external network")
		}
	case NetworkBMC:
		if n.UseNAT {
			return errors.New("useNAT must be false for BMC network")
		}
		if len(n.Address) == 0 {
			return errors.New("address must be specified for BMC network")
		}
	default:
		return errors.New("unknown type: " + n.Type)
	}

	return nil
}

// NetNSSpec represents a NetworkNamespace specification in YAML
type NetNSSpec struct {
	Kind        string                `json:"kind"`
	Name        string                `json:"name"`
	Interfaces  []*NetNSInterfaceSpec `json:"interfaces"`
	Apps        []*NetNSAppSpec       `json:"apps,omitempty"`
	InitScripts []string              `json:"init-scripts,omitempty"`
}

func (n *NetNSSpec) validate() error {
	if len(n.Name) == 0 {
		return errors.New("network namespace is empty")
	}

	if len(n.Interfaces) == 0 {
		return fmt.Errorf("no interface for Network Namespace %s", n.Name)
	}

	for _, app := range n.Apps {
		if len(app.Command) == 0 {
			return fmt.Errorf("no command for app %s", app.Name)
		}
	}
	return nil
}

// NetNSInterfaceSpec represents a NetworkNamespace's Interface definition in YAML
type NetNSInterfaceSpec struct {
	Network   string   `json:"network"`
	Addresses []string `json:"addresses,omitempty"`
}

// NetNSAppSpec represents a NetworkNamespace's App definition in YAML
type NetNSAppSpec struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
}

const (
	NodeVolumeCacheWriteback    = "writeback"
	NodeVolumeCacheNone         = "none"
	NodeVolumeCacheWritethrough = "writethrough"
	NodeVolumeCacheDirectSync   = "directsync"
	NodeVolumeCacheUnsafe       = "unsafe"

	NodeVolumeKindImage   = "image"
	NodeVolumeKindLocalds = "localds"
	NodeVolumeKindRaw     = "raw"
	NodeVolumeKindLv      = "lv"
	NodeVolumeKind9p      = "9p"

	NodeVolumeFormatQcow2 = "qcow2"
	NodeVolumeFormatRaw   = "raw"
)

// NodeSpec represents a Node specification in YAML
type NodeSpec struct {
	Kind         string           `json:"kind"`
	Name         string           `json:"name"`
	Interfaces   []string         `json:"interfaces,omitempty"`
	Volumes      []NodeVolumeSpec `json:"volumes,omitempty"`
	IgnitionFile string           `json:"ignition,omitempty"`
	CPU          int              `json:"cpu,omitempty"`
	Memory       string           `json:"memory,omitempty"`
	UEFI         bool             `json:"uefi,omitempty"`
	TPM          bool             `json:"tpm,omitempty"`
	SMBIOS       SMBIOSConfigSpec `json:"smbios,omitempty"`
}

func (n *NodeSpec) validate() error {
	if n.Name == "" {
		return errors.New("node name is empty")
	}
	return nil
}

// SMBIOSConfigSpec represents a Node's SMBIOS definition in YAML
type SMBIOSConfigSpec struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Product      string `json:"product,omitempty"`
	Serial       string `json:"serial,omitempty"`
}

// NodeVolumeSpec represents a Node's Volume specification in YAML
type NodeVolumeSpec struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Image         string `json:"image,omitempty"`
	UserData      string `json:"user-data,omitempty"`
	NetworkConfig string `json:"network-config,omitempty"`
	Size          string `json:"size,omitempty"`
	Folder        string `json:"folder,omitempty"`
	CopyOnWrite   bool   `json:"copy-on-write,omitempty"`
	Cache         string `json:"cache,omitempty"`
	Format        string `json:"format,omitempty"`
	VG            string `json:"vg,omitempty"`
	Writable      bool   `json:"writable,omitempty"`
}

func (n *NodeVolumeSpec) validate() error {
	switch n.Cache {
	case "":
		n.Cache = NodeVolumeCacheNone
	case NodeVolumeCacheWriteback, NodeVolumeCacheNone, NodeVolumeCacheWritethrough, NodeVolumeCacheDirectSync, NodeVolumeCacheUnsafe:
	default:
		return errors.New("invalid cache type for volume")
	}

	switch n.Kind {
	case NodeVolumeKindImage:
		if n.Image == "" {
			return errors.New("image volume must specify an image name")
		}
	case NodeVolumeKindLocalds:
		if n.UserData == "" {
			return errors.New("localds volume must specify user-data")
		}
	case NodeVolumeKindRaw:
		if n.Size == "" {
			return errors.New("raw volume must specify size")
		}
		switch n.Format {
		case "":
			n.Format = NodeVolumeFormatQcow2
		case NodeVolumeFormatQcow2, NodeVolumeFormatRaw:
		default:
			return errors.New("invalid format for raw volume")
		}
	case NodeVolumeKindLv:
		if n.Size == "" {
			return errors.New("lv volume must specify size")
		}
		if n.VG == "" {
			return errors.New("lv volume must specify vg")
		}
	case NodeVolumeKind9p:
		if n.Folder == "" {
			return errors.New("9p volume must specify a folder name")
		}
	default:
		return errors.New("unknown volume kind: " + n.Kind)
	}

	return nil
}

// ImageSpec represents an Image specification in YAML.
type ImageSpec struct {
	Kind              string `json:"kind"`
	Name              string `json:"name"`
	URL               string `json:"url,omitempty"`
	File              string `json:"file,omitempty"`
	CompressionMethod string `json:"compression,omitempty"`
}

func (i *ImageSpec) validate() error {
	if len(i.Name) == 0 {
		return errors.New("invalid image spec: " + i.Name)
	}

	if len(i.URL) == 0 && len(i.File) == 0 {
		return errors.New("invalid image spec: " + i.Name)
	}

	if len(i.URL) > 0 {
		if len(i.File) > 0 {
			return errors.New("invalid image spec: " + i.Name)
		}
	}

	return nil
}

type baseConfig struct {
	Kind string `json:"kind"`
}

// Parse reads a yaml document and create Cluster
func Parse(r io.Reader) (*Cluster, error) {
	cluster := &Cluster{}
	f := json.YAMLFramer.NewFrameReader(ioutil.NopCloser(r))
	for {
		y, err := readSingleYamlDoc(f)
		if err == io.EOF {
			break
		}
		b := &baseConfig{}
		if err := yaml.Unmarshal(y, b); err != nil {
			return nil, fmt.Errorf("failed to unmarshal the yaml document %s: %w", y, err)
		}

		switch b.Kind {
		case "Network":
			n := &NetworkSpec{}
			if err := yaml.Unmarshal(y, n); err != nil {
				return nil, fmt.Errorf("failed to unmarshal the Network yaml document %s: %w", y, err)
			}
			if err := n.validate(); err != nil {
				return nil, fmt.Errorf("invalid Network resource: %w", err)
			}
			cluster.Networks = append(cluster.Networks, n)
		case "NetworkNamespace":
			n := &NetNSSpec{}
			if err := yaml.Unmarshal(y, n); err != nil {
				return nil, fmt.Errorf("failed to unmarshal the NetworkNamespace yaml document %s: %w", y, err)
			}
			if err := n.validate(); err != nil {
				return nil, fmt.Errorf("invalid NetworkNamespace resource: %w", err)
			}
			cluster.NetNSs = append(cluster.NetNSs, n)
		case "Node":
			n := &NodeSpec{}
			if err := yaml.Unmarshal(y, n); err != nil {
				return nil, fmt.Errorf("failed to unmarshal the Node yaml document %s: %w", y, err)
			}
			if err := n.validate(); err != nil {
				return nil, fmt.Errorf("invalid Node resource: %w", err)
			}
			for _, v := range n.Volumes {
				if err := v.validate(); err != nil {
					return nil, fmt.Errorf("invalid NodeVolume resource: %w", err)
				}
			}
			cluster.Nodes = append(cluster.Nodes, n)
		case "Image":
			i := &ImageSpec{}
			if err := yaml.Unmarshal(y, i); err != nil {
				return nil, fmt.Errorf("failed to unmarshal the Image yaml document %s: %w", y, err)
			}
			if err := i.validate(); err != nil {
				return nil, fmt.Errorf("invalid Image resource: %w", err)
			}
			cluster.Images = append(cluster.Images, i)
		default:
			return nil, errors.New("unknown resource: " + b.Kind)
		}
	}
	return cluster, nil
}

func readSingleYamlDoc(reader io.Reader) ([]byte, error) {
	buf := make([]byte, 1024)
	maxBytes := 16 * 1024 * 1024
	base := 0
	for {
		n, err := reader.Read(buf[base:])
		if err == io.ErrShortBuffer {
			if n == 0 {
				return nil, fmt.Errorf("got short buffer with n=0, base=%d, cap=%d", base, cap(buf))
			}
			if len(buf) < maxBytes {
				base += n
				buf = append(buf, make([]byte, len(buf))...)
				continue
			}
			return nil, errors.New("yaml document is too large")
		}
		if err != nil {
			return nil, err
		}
		base += n
		return buf[:base], nil
	}
}

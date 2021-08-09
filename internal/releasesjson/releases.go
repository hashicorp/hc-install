package releasesjson

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/httpclient"
)

const defaultBaseURL = "https://releases.hashicorp.com"

// Product is a top-level product like "Consul" or "Nomad". A Product may have
// one or more versions.
type Product struct {
	Name     string                     `json:"name"`
	Versions map[string]*ProductVersion `json:"versions"`
}

// ProductVersion is a wrapper around a particular product version like
// "consul 0.5.1". A ProductVersion may have one or more builds.
type ProductVersion struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	SHASUMS     string        `json:"shasums,omitempty"`
	SHASUMSSig  string        `json:"shasums_signature,omitempty"`
	SHASUMSSigs []string      `json:"shasums_signatures,omitempty"`
	Builds      ProductBuilds `json:"builds"`
}

type ProductBuilds []*ProductBuild

func (pbs ProductBuilds) BuildForOsArch(os string, arch string) (*ProductBuild, bool) {
	for _, pb := range pbs {
		if pb.OS == os && pb.Arch == arch {
			return pb, true
		}
	}
	return nil, false
}

// ProductBuild is an OS/arch-specific representation of a product. This is the
// actual file that a user would download, like "consul_0.5.1_linux_amd64".
type ProductBuild struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type Releases struct {
	logger  *log.Logger
	BaseURL string
}

func NewReleases() *Releases {
	return &Releases{
		logger:  log.New(ioutil.Discard, "", 0),
		BaseURL: defaultBaseURL,
	}
}

func (r *Releases) SetLogger(logger *log.Logger) {
	r.logger = logger
}

func (r *Releases) ListProductVersions(ctx context.Context, productName string) (map[string]*ProductVersion, error) {
	client := httpclient.NewHTTPClient()

	productIndexURL := fmt.Sprintf("%s/%s/index.json", r.BaseURL, productName)
	r.logger.Printf("requesting versions from %s", productIndexURL)

	resp, err := client.Get(productIndexURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r.logger.Printf("received %s", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	p := Product{}
	err = json.Unmarshal(body, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q", string(body))
	}

	for rawVersion := range p.Versions {
		v, err := version.NewVersion(rawVersion)
		if err != nil {
			// remove unparseable version
			delete(p.Versions, rawVersion)
			continue
		}

		if ok, _ := versionIsSupported(v); !ok {
			// Remove (currently unsupported) enterprise
			// version and any other "custom" build
			delete(p.Versions, rawVersion)
		}
	}

	return p.Versions, nil
}

func (r *Releases) GetProductVersion(ctx context.Context, product string, version *version.Version) (*ProductVersion, error) {
	if ok, err := versionIsSupported(version); !ok {
		return nil, fmt.Errorf("%s: %w", product, err)
	}

	client := httpclient.NewHTTPClient()

	indexURL := fmt.Sprintf("%s/%s/%s/index.json", r.BaseURL, product, version)
	r.logger.Printf("requesting version from %s", indexURL)

	resp, err := client.Get(indexURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r.logger.Printf("received %s", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pv := &ProductVersion{}
	err = json.Unmarshal(body, pv)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q", string(body))
	}

	return pv, nil
}

func versionIsSupported(v *version.Version) (bool, error) {
	isSupported := v.Metadata() == ""
	if !isSupported {
		return false, fmt.Errorf("cannot obtain %s (enterprise versions are not supported)",
			v.String())
	}
	return true, nil
}

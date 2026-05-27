package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RegistryClient interacts with the npm Registry
type RegistryClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewRegistryClient creates a new RegistryClient with connection pooling
func NewRegistryClient() *RegistryClient {
	t := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
	}
	return &RegistryClient{
		httpClient: &http.Client{
			Transport: t,
			Timeout:   10 * time.Second,
		},
		baseURL: "https://registry.npmjs.org",
	}
}

// PackageVersionInfo contains details about a single version of a package
type PackageVersionInfo struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Dist            struct {
		Tarball   string `json:"tarball"`
		Shasum    string `json:"shasum"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
	License string `json:"license"`
}

// PackageMetadata represents the response from registry.npmjs.org/<name>
type PackageMetadata struct {
	Name     string                         `json:"name"`
	Versions map[string]PackageVersionInfo `json:"versions"`
	DistTags map[string]string              `json:"dist-tags"`
	Time     map[string]string              `json:"time"`
}

// SearchPackageInfo represents info about a package found in search
type SearchPackageInfo struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Description     string `json:"description"`
	WeeklyDownloads int64  `json:"weekly_downloads"`
}

// SearchResponse represents the response from registry.npmjs.org/-/v1/search
type SearchResponse struct {
	Objects []struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Date        string `json:"date"`
		} `json:"package"`
		Score struct {
			Detail struct {
				Popularity float64 `json:"popularity"`
			} `json:"detail"`
		} `json:"score"`
	} `json:"objects"`
}

// GetPackage fetches metadata for a specific package
func (c *RegistryClient) GetPackage(name string) (*PackageMetadata, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	var metadata PackageMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// SearchPackages searches the npm registry for packages matching the query
func (c *RegistryClient) SearchPackages(query string, limit int) ([]SearchPackageInfo, error) {
	if query == "" {
		return nil, nil
	}

	apiURL := fmt.Sprintf("%s/-/v1/search?text=%s&size=%d", c.baseURL, url.QueryEscape(query), limit)
	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	results := make([]SearchPackageInfo, len(searchResp.Objects))
	for i, obj := range searchResp.Objects {
		// Mock weekly downloads based on popularity score
		downloads := int64(obj.Score.Detail.Popularity * 12500000)
		if downloads < 500 {
			downloads = int64(100 + i*50)
		}
		results[i] = SearchPackageInfo{
			Name:            obj.Package.Name,
			Version:         obj.Package.Version,
			Description:     obj.Package.Description,
			WeeklyDownloads: downloads,
		}
	}

	return results, nil
}

// DownloadTarball downloads a package tarball and returns the response body reader
func (c *RegistryClient) DownloadTarball(tarballURL string) (io.ReadCloser, int64, error) {
	resp, err := c.httpClient.Get(tarballURL)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("failed to download tarball: %s (status %d)", tarballURL, resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

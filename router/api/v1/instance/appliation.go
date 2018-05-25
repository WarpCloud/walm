package instance

type Application struct {
	Name       string   `json:"name"  description:"name of the release"`
	Namespace  string   `json:"namespace,omitempty"  description:"namespace of the release"`
	Version    string   `json:"version,omitempty"  description:"version of the release"`
	Value      string   `json:"value,omitempty"  description:"value array of the release (json format)"`
	Links      []string `json:"links,omitempty"  description:"link array of the release"`
	Repo       string   `json:"repo,omitempty"  description:"repo where to find chart"`
	Install    bool     `json:"install,omitempty"  description:"if install when not existed"`
	ResetValue bool     `json:"reset-values,omitempty"  description:"if reset value when install"`
	ReuseValue bool     `json:"reuse-values,omitempty"  description:"if reuse value when install"`
}

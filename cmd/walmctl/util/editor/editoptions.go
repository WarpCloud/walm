package editor

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
	goruntime "runtime"

	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"WarpCloud/walm/pkg/release"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	yaml2 "github.com/ghodss/yaml"
	"github.com/go-resty/resty"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
	"io"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor/crlf"
	"os"
	"path/filepath"
	"strings"
)

// EditOptions contains all the options for running edit cli command.

type EditOptions struct {
	EditMode           EditMode
	WindowsLineEndings bool
	Namespace          string
	WalmServer         string
	SourceType         string
	SourceName         string
	ChangeCause        string
	genericclioptions.IOStreams
	editPrinterOptions *editPrinterOptions
	PrintFlags         *genericclioptions.PrintFlags
}

func NewEditOptions(editMode EditMode, ioStreams genericclioptions.IOStreams) *EditOptions {
	return &EditOptions{

		EditMode: editMode,

		PrintFlags: genericclioptions.NewPrintFlags("edited").WithTypeSetter(scheme.Scheme),

		editPrinterOptions: &editPrinterOptions{
			// create new editor-specific PrintFlags, with all
			// output flags disabled, except json / yaml
			printFlags: (&genericclioptions.PrintFlags{
				JSONYamlPrintFlags: genericclioptions.NewJSONYamlPrintFlags(),
			}).WithDefaultOutput("yaml"),
			ext:       ".yaml",
			addHeader: true,
		},

		WindowsLineEndings: goruntime.GOOS == "windows",
		IOStreams:          ioStreams,
	}
}

type editPrinterOptions struct {
	printFlags *genericclioptions.PrintFlags
	ext        string
	addHeader  bool
}

func (e *editPrinterOptions) Complete(fromPrintFlags *genericclioptions.PrintFlags) error {
	if e.printFlags == nil {
		return fmt.Errorf("missing PrintFlags in editor printer options")
	}

	// bind output format from existing printflags
	if fromPrintFlags != nil && len(*fromPrintFlags.OutputFormat) > 0 {
		e.printFlags.OutputFormat = fromPrintFlags.OutputFormat
	}

	// prevent a commented header at the top of the user's
	// default editor if presenting contents as json.
	if *e.printFlags.OutputFormat == "json" {
		e.addHeader = false
		e.ext = ".json"
		return nil
	}

	// we default to yaml if check above is false, as only json or yaml are supported
	e.addHeader = true
	e.ext = ".yaml"
	return nil
}

type EditMode string

const (
	NormalEditMode       EditMode = "normal_mode"
	EditBeforeCreateMode EditMode = "edit_before_create_mode"
	ApplyEditMode        EditMode = "edit_last_applied_mode"
)

// Validate checks the EditOptions to see if there is sufficient information to run the command.
func (o *EditOptions) Validate() error {
	return nil
}

func (o *EditOptions) Run() error {

	o.editPrinterOptions.Complete(o.PrintFlags)
	// check resource
	var resp *resty.Response
	var releaseInfo release.ReleaseInfoV2
	var err error
	client := walmctlclient.CreateNewClient(o.WalmServer)
	if err = client.ValidateHostConnect(); err != nil {
		return err
	}
	if o.SourceType == "release" {
		resp, err = client.GetRelease(o.Namespace, o.SourceName)
	} else {
		resp, err = client.GetProject(o.Namespace, o.SourceName)
	}

	if err != nil {
		return errors.Errorf("%s %s not found in redis\n", o.SourceType, o.SourceName)
	}
	err = json.Unmarshal(resp.Body(), &releaseInfo)
	if err != nil {
		return err
	}

	edit := NewDefaultEditor(editorEnvs())
	// editFn is invoked for each edit session (once with a list for normal edit, once for each individual resource in a edit-on-create invocation)
	editFn := func(releaseRequest *release.ReleaseRequestV2) error {

		var (
			results = editResults{}
			//original []byte
			edited []byte
			file   string
		)
		containsError := false

		// generate the file to edit
		buf := &bytes.Buffer{}
		var w io.Writer = buf
		if o.WindowsLineEndings {
			w = crlf.NewCRLFWriter(w)
		}
		if o.editPrinterOptions.addHeader {
			results.header.writeTo(w, o.EditMode)
		}

		if !containsError {

			if err := o.editPrinterOptions.PrintRequest(releaseRequest, w); err != nil {
				return preservedFile(err, results.file, o.ErrOut)
			}
			// []byte without name
			//original = buf.Bytes()
		} else {
			// Todo:
			// In case of an error, preserve the edited file.
			// Remove the comments (header) from it since we already
			// have included the latest header in the buffer above.
		}

		// launch the editor
		editedDiff := edited
		edited, file, err = edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), o.editPrinterOptions.ext, buf)
		if err != nil {
			return preservedFile(err, results.file, o.ErrOut)
		}

		// If we're retrying the loop because of an error, and no change was made in the file, short-circuit
		if containsError && bytes.Equal(StripComments(editedDiff), StripComments(edited)) {
			return preservedFile(fmt.Errorf("%s", "Edit cancelled, no valid changes were saved."), file, o.ErrOut)
		}

		// cleanup any file from the previous pass
		if len(results.file) > 0 {
			os.Remove(results.file)
		}
		klog.V(4).Infof("User edited:\n%s", string(edited))

		// Todo:// if not change, not send request
		// Compare content without comments
		//if bytes.Equal(StripComments(original), StripComments(edited)) {
		//	os.Remove(file)
		//	fmt.Fprintln(o.ErrOut, "Edit cancelled, no changes made.")
		//	return nil
		//}

		lines, err := hasLines(bytes.NewBuffer(edited))
		if err != nil {
			return preservedFile(err, file, o.ErrOut)
		}

		if !lines {
			os.Remove(file)
			fmt.Fprintln(o.ErrOut, "Edit cancelled, saved file was empty.")
			return nil
		}

		edited, err = yaml.ToJSON(edited)
		edited, err = sjson.SetBytes(edited, "name", o.SourceName)
		if err != nil {
			return err
		}

		resp, err = client.UpdateRelease(o.Namespace, string(StripComments(edited)), true, 0)
		if err != nil {
			return err
		}

		//Todo: cleanUp, apply validation, syntax check

		fmt.Printf("edit resource succeed.\n")
		return nil
	}

	switch o.EditMode {
	// If doing normal edit we cannot use Visit because we need to edit a list for convenience. Ref: #20519
	case NormalEditMode:
		releaseRequest := releaseInfo.BuildReleaseRequestV2()
		if releaseRequest.Name == "" {
			return errors.New("edit cancelled, no resources found.")
		}
		return editFn(releaseRequest)
	default:
		return fmt.Errorf("unsupported edit mode %q", o.EditMode)
	}
	return nil
}

func (e *editPrinterOptions) PrintRequest(releaseRequest *release.ReleaseRequestV2, out io.Writer) error {

	var releaseRequestByte []byte
	var err error

	releaseRequestByte, err = json.Marshal(releaseRequest)
	if err != nil {
		return err
	}

	// Todo: disable print name [Done]
	releaseRequestByte, err = sjson.DeleteBytes(releaseRequestByte, "name")
	if err != nil {
		return err
	}

	if *e.printFlags.OutputFormat == "yaml" {
		releaseRequestByte, err = yaml2.JSONToYAML(releaseRequestByte)
	} else {
		releaseRequestByte, err = yaml2.YAMLToJSON(releaseRequestByte)
	}
	if err != nil {
		return err
	}

	fmt.Fprint(out, string(releaseRequestByte))
	return nil
}

// StripComments will transform a YAML file into JSON, thus dropping any comments
// in it. Note that if the given file has a syntax error, the transformation will
// fail and we will manually drop all comments from the file.
func StripComments(file []byte) []byte {
	stripped := file
	stripped, err := yaml.ToJSON(stripped)
	if err != nil {
		stripped = ManualStrip(file)
	}
	return stripped
}

// ManualStrip is used for dropping comments from a YAML file
func ManualStrip(file []byte) []byte {
	var stripped []byte
	lines := bytes.Split(file, []byte("\n"))
	for i, line := range lines {
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte("#")) {
			continue
		}
		stripped = append(stripped, line...)
		if i < len(lines)-1 {
			stripped = append(stripped, '\n')
		}
	}
	return stripped
}

// editReason preserves a message about the reason this file must be edited again
type editReason struct {
	head  string
	other []string
}

// editHeader includes a list of reasons the edit must be retried
type editHeader struct {
	reasons []editReason
}

// writeTo outputs the current header information into a stream
func (h *editHeader) writeTo(w io.Writer, editMode EditMode) error {
	if editMode == ApplyEditMode {
		fmt.Fprint(w, `# Please edit the 'last-applied-configuration' annotations below.
# Lines beginning with a '#' will be ignored, and an empty file will abort the edit.
#
`)
	} else {
		fmt.Fprint(w, `# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
`)
	}

	for _, r := range h.reasons {
		if len(r.other) > 0 {
			fmt.Fprintf(w, "# %s:\n", hashOnLineBreak(r.head))
		} else {
			fmt.Fprintf(w, "# %s\n", hashOnLineBreak(r.head))
		}
		for _, o := range r.other {
			fmt.Fprintf(w, "# * %s\n", hashOnLineBreak(o))
		}
		fmt.Fprintln(w, "#")
	}
	return nil
}

func (h *editHeader) flush() {
	h.reasons = []editReason{}
}

type editResults struct {
	header    editHeader
	retryable int
	notfound  int
	edit      *release.ReleaseRequestV2
	file      string
}

// preservedFile writes out a message about the provided file if it exists to the
// provided output stream when an error happens. Used to notify the user where
// their updates were preserved.
func preservedFile(err error, path string, out io.Writer) error {
	if len(path) > 0 {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			fmt.Fprintf(out, "A copy of your changes has been stored to %q\n", path)
		}
	}
	return err
}

// hasLines returns true if any line in the provided stream is non empty - has non-whitespace
// characters, or the first non-whitespace character is a '#' indicating a comment. Returns
// any errors encountered reading the stream.
func hasLines(r io.Reader) (bool, error) {
	// TODO: if any files we read have > 64KB lines, we'll need to switch to bytes.ReadLine
	// TODO: probably going to be secrets
	s := bufio.NewScanner(r)
	for s.Scan() {
		if line := strings.TrimSpace(s.Text()); len(line) > 0 && line[0] != '#' {
			return true, nil
		}
	}
	if err := s.Err(); err != nil && err != io.EOF {
		return false, err
	}
	return false, nil
}

// hashOnLineBreak returns a string built from the provided string by inserting any necessary '#'
// characters after '\n' characters, indicating a comment.
func hashOnLineBreak(s string) string {
	r := ""
	for i, ch := range s {
		j := i + 1
		if j < len(s) && ch == '\n' && s[j] != '#' {
			r += "\n# "
		} else {
			r += string(ch)
		}
	}
	return r
}

// editorEnvs returns an ordered list of env vars to check for editor preferences.
func editorEnvs() []string {
	return []string{
		"KUBE_EDITOR",
		"EDITOR",
	}
}

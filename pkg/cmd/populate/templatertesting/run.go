package templatertesting

import (
	"fmt"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-secret/pkg/extsecrets/testsecrets"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	secretstorefake "github.com/jenkins-x-plugins/secretfacade/testing/fake"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner/fakerunner"
	v1 "github.com/jenkins-x/jx-secret/pkg/apis/external/v1"
	"github.com/jenkins-x/jx-secret/pkg/cmd/populate"
	"github.com/jenkins-x/jx-secret/pkg/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

// Run runs the test cases
func (r *Runner) Run(t *testing.T) {
	require.NotEmpty(t, r.TestCases, "no TestCases supplied")
	require.NotEmpty(t, r.SchemaFile, "no SchemaFile supplied")

	schema, err := schemas.LoadSchemaFile(r.SchemaFile)
	require.NoError(t, err, "failed to load schema file %s", r.SchemaFile)

	_, o := populate.NewCmdPopulate()
	o.NoWait = true

	o.Namespace = r.Namespace

	runner := &fakerunner.FakeRunner{}
	o.CommandRunner = runner.Run

	for _, tc := range r.TestCases {
		kubeObjects := append(r.KubeObjects, tc.KubeObjects...)
		o.KubeClient = fake.NewSimpleClientset(testsecrets.AddVaultSecrets(kubeObjects...)...)

		fakeFactory := &secretstorefake.FakeSecretManagerFactory{}
		o.SecretStoreManagerFactory = fakeFactory

		_, err = fakeFactory.NewSecretManager(tc.ExternalSecretStorageType)
		assert.NoError(t, err)

		fakeStore := fakeFactory.GetSecretStore()

		o.Options.ExternalSecrets = []*v1.ExternalSecret{}
		for _, p := range tc.ExternalSecrets {
			es := p.ExternalSecret
			o.Options.ExternalSecrets = append(o.Options.ExternalSecrets, &es)
			err := fakeStore.SetSecret(p.Location, p.Name, &p.Value)
			assert.NoError(t, err)
		}

		objName := tc.ObjectName
		if tc.Requirements != nil {
			o.Requirements = tc.Requirements
		}
		if o.Requirements == nil {
			if tc.Dir != "" {
				o.Dir = tc.Dir
			}
			if o.Dir == "" {
				o.Dir = r.Dir
			}
			require.NotEmpty(t, o.Dir, "you must either specify Requirements or a Dir on the Runner or TestCase to be able to detect the Requirements to use the the template generation")
		}
		object := schema.Spec.FindObject(objName)
		require.NotNil(t, object, "could not find schema for object name %s", objName)

		propName := tc.Property
		property := object.FindProperty(propName)
		require.NotNil(t, property, "could not find property for object %s name %s", objName, propName)

		templateText := property.Template
		require.NotEmpty(t, templateText, "no template defined for object %s property name %s", objName, propName)

		text, err := o.EvaluateTemplate(r.Namespace, objName, propName, templateText, false)
		require.NoError(t, err, "failed to evaluate template for object %s property name %s", objName, propName)

		message := fmt.Sprintf("test %s for object %s property name %s", tc.TestName, objName, propName)
		require.NotEmpty(t, text, message)

		format := tc.Format
		if format == "" {
			format = "txt"
		}
		switch format {
		case "xml":
			AssertValidXML(t, text, message)
		case "json":
			AssertValidJSON(t, text, message)
		}

		expectedDir := filepath.Join("test_data", "template", "expected", objName, propName)
		expectedFile := filepath.Join(expectedDir, tc.TestName+"."+format)

		if tc.GenerateTestOutput {
			err = os.MkdirAll(expectedDir, files.DefaultDirWritePermissions)
			require.NoError(t, err, "failed to create dir %s", expectedDir)

			err = ioutil.WriteFile(expectedFile, []byte(text), files.DefaultFileWritePermissions)
			require.NoError(t, err, "failed to save file %s", expectedFile)

			t.Logf("generated %s\n", expectedFile)

		} else {
			require.FileExists(t, expectedFile, "for expected output")

			expected, err := ioutil.ReadFile(expectedFile)
			require.NoError(t, err, "failed to load %s", expectedFile)

			if tc.VerifyFn != nil {
				tc.VerifyFn(t, text)
			} else if assert.Equal(t, string(expected), text, "generated template for %s", message) {
				t.Logf("got expected file %s for %s\n", expectedFile, message)
			} else {
				t.Logf("generated for %s:\n%s\n", message, text)
			}
		}
	}
}

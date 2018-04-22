package webapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/queued_verify"
	"bitbucket.org/digitorus/pdfsigner/queuedsign"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"bitbucket.org/digitorus/pdfsigner/version"
	"github.com/stretchr/testify/assert"
)

type filePart struct {
	fieldName string
	path      string
}

var (
	wa      *WebAPI
	qs      *queuedsign.QSign
	proto   = "http://"
	addr    = "localhost:3000"
	baseURL = proto + addr
)

// TestMain setup / tear down before tests
func TestMain(m *testing.M) {
	os.Exit(runTest(m))
}

// runTest initializes the environment
func runTest(m *testing.M) int {
	err := license.Load()
	if err != nil {
		log.Fatal(err)
	}

	// create new QSign
	qs = queuedsign.NewQSign()

	// create signer
	signData := signer.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Tim",
				Location:    "Spain",
				Reason:      "Test",
				ContactInfo: "None",
				Date:        time.Now().Local(),
			},
			CertType: 2,
			Approval: false,
		},
	}
	signData.SetPEM("../testfiles/test.crt", "../testfiles//test.pem", "")
	qs.AddSigner("simple", signData, 10)
	qs.Runner()

	qv := queued_verify.NewQVerify()
	qv.Runner()

	// create web api
	wa = NewWebAPI(addr, qs, qv, []string{
		"simple",
	}, version.Version{Version: "0.1"})

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	return m.Run()
}

func TestUploadCheckDownload(t *testing.T) {
	// test upload
	//create file parts
	fileParts := []filePart{
		{"testfile1", "../testfiles/testfile20.pdf"},
		{"testfile2", "../testfiles/testfile20.pdf"},
		{"testfile2", "../testfiles/testfile20.pdf"},
	}
	// create multipart request
	r, err := newMultipleFilesUploadRequest(
		baseURL+"/sign",
		map[string]string{
			"signer":      "simple",
			"name":        "My Name",
			"location":    "My Location",
			"reason":      "My Reason",
			"contactInfo": "My ContactInfo",
			"certType":    "1",
			"approval":    "true",
		}, fileParts)
	if err != nil {
		t.Fatal(err)
	}

	// create recorder
	w := httptest.NewRecorder()
	// make request
	wa.r.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status not ok: %v", w.Body.String())
	}

	// get job id
	var scheduleResponse handleSignScheduleResponse
	if err := json.NewDecoder(w.Body).Decode(&scheduleResponse); err != nil {
		t.Fatal(err)
	}
	if scheduleResponse.JobID == "" {
		t.Fatal("not received jobID")
	}

	if w.Header().Get("Location") != "/sign/"+scheduleResponse.JobID {
		t.Fatal("location is not set")
	}

	// wait for signing files
	time.Sleep(2 * time.Second)

	// test check
	r = httptest.NewRequest("GET", baseURL+"/sign/"+scheduleResponse.JobID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status not ok: %v", w.Body.String())
	}

	var jobStatus JobStatus
	if err := json.NewDecoder(w.Body).Decode(&jobStatus); err != nil {
		t.Fatal(err)
	}

	//assert.Equal(t, true, job.IsCompleted)
	assert.Equal(t, 3, len(jobStatus.Tasks))
	for _, task := range jobStatus.Tasks {
		assert.Equal(t, queuedsign.StatusCompleted, task.Status)

		// test get completed task
		r = httptest.NewRequest("GET", baseURL+"/sign/"+scheduleResponse.JobID+"/"+task.ID+"/download", nil)
		w = httptest.NewRecorder()
		wa.r.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status not ok: %v", w.Body.String())
		}
		assert.Equal(t, 8994, len(w.Body.Bytes()))
	}

	// test delete job
	r = httptest.NewRequest("DELETE", baseURL+"/sign/"+scheduleResponse.JobID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status not ok: %v", w.Body.String())
	}

	r = httptest.NewRequest("GET", baseURL+"/sign/"+scheduleResponse.JobID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)

	if w.Code == http.StatusOK {
		t.Fatalf("not removed")
	}

	// test get version
	r = httptest.NewRequest("GET", baseURL+"/version", nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status not ok: %v", w.Body.String())
	}
	var ver version.Version
	if err := json.NewDecoder(w.Body).Decode(&ver); err != nil {
		t.Fatal(err)
	}
	if ver.Version != "0.1" {
		t.Fatal("Version is not correct")
	}

}

// Creates a new multiple files upload http request with optional extra params
func newMultipleFilesUploadRequest(uri string, params map[string]string, fileParts []filePart) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	for _, f := range fileParts {
		file, err := os.Open(f.path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(f.fieldName, filepath.Base(f.path))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

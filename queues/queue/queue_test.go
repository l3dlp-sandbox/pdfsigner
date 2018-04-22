package queue

import (
	"testing"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/queues/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/stretchr/testify/assert"
)

func TestQSignersMap(t *testing.T) {
	license.Load()

	// create sign data
	d := signer.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Tim",
				Location:    "Spain",
				Reason:      "Test",
				ContactInfo: "None",
			},
			CertType: 2,
			Approval: false,
		},
	}
	d.SetPEM("../../testfiles/test.crt", "../../testfiles/test.pem", "")

	// create Queue
	qs := NewQueue()

	// add signer
	qs.AddUnit("simple", d)

	// create session
	var signData signer.SignData
	jobID := qs.AddJob(signData)
	job, err := qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, jobID, job.ID)

	// add job
	taskID, err := qs.AddTask(
		"simple",
		jobID,
		"../../testfiles/testfile20.pdf",
		"../../testfiles/testfile20_signed.pdf",
		priority_queue.HighPriority,
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(job.TasksMap))

	// sign job
	qs.processNextTask("simple")

	job, err = qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(job.TasksMap))
	assert.Equal(t, StatusCompleted, job.TasksMap[taskID].Status)
}
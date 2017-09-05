package signer

import (
	"crypto"
	"io"
	"io/ioutil"
)

// by default we queue 25 signing jobs
// TODO: this needs to come from the configuration
const queueSize = 25

// Priority of signing request
type Priority int

const (
	// UnknownPriority represents an Unknown Priority siging request
	UnknownPriority Priority = iota
	// LowPriority represents an Low Priority siging request
	LowPriority
	// MediumPriority represents an Medium Priority siging request
	MediumPriority
	// HighPriority represents an High Priority siging request
	HighPriority
)

// signJob is the internal job specification
type signJob struct {
	file    string // tmp file location
	options *Options
}

type priorityQueue struct {
	low    chan signJob
	medium chan signJob
	heigh  chan signJob
}

type signer struct {
	c crypto.Signer
	q priorityQueue
}

// Status contains the current signing proccess status for a specific document
type Status struct {
	Ready bool // true when the document is signed
}

// Options contains information required to schedule a document to be processed
// by one of the signers.
type Options struct {
	// must include info for sign.SignData
	Priority Priority // batch procedures should run with a low priority
}

// Signer exposes an transparent interface to the sign queue, all clients should
// implement this interface.
//
// The crypto.Signer map can contain multiple singers as defined in the config,
// a Singer implementation can be a private key or PKCS#11 device.
type Signer struct {
	s map[string]signer
}

// Sign reads a file and stores it at temporary location so that it can be
// processed later without consuming memeory. The function returns a tracking
// id or error.
func (s *Signer) Sign(file io.Reader, options *Options) (*string, error) {
	// TODO: Should we encrypt temporary files?
	tmpfile, err := ioutil.TempFile("", "pdf")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmpfile, file)
	if err != nil {
		return nil, err
	}

	// based on the request we must identify what device or private key we should
	// be using.
	//
	// TODO: signer priority queues need to be initialized with make(chan signJob, queueSize)
	signer := "s.c-id"

	job := signJob{
		file:    tmpfile.Name(),
		options: options,
	}

	// Add job to the signing queue according to it's priority
	switch options.Priority {
	case HighPriority:
		s.s[signer].q.heigh <- job
	case MediumPriority:
		s.s[signer].q.medium <- job
	default:
		s.s[signer].q.low <- job
	}

	// create a unqiue id that can be used by a client to obtain the document or
	// current state of the job
	tracker := tmpfile.Name()

	return &tracker, nil
}

// Get returns the signed document based on the tracker id.
func (s *Signer) Get(tracker string) (*io.Reader, error) {
	return nil, nil
}

// Status returns if the document has been signed already.
func (s *Signer) Status(tracker string) (*Status, error) {
	// TODO: calculate retry time based on queue lenght and documents per second
	return nil, nil
}

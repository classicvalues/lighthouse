package changesets

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nwidger/lighthouse"
)

type Service struct {
	basePath string
	s        *lighthouse.Service
}

func NewService(s *lighthouse.Service, projectID int) (*Service, error) {
	return &Service{
		basePath: s.BasePath + "/projects/" + strconv.Itoa(projectID) + "/changesets",
		s:        s,
	}, nil
}

type Change []string

func (c *Change) Operation() string {
	if len(*c) != 2 {
		return ""
	}
	return (*c)[0]
}

func (c *Change) Path() string {
	if len(*c) != 2 {
		return ""
	}
	return (*c)[1]
}

type Changes []Change

type Changeset struct {
	Body      string     `json:"body"`
	BodyHTML  string     `json:"body_html"`
	ChangedAt *time.Time `json:"changed_at"`
	Changes   Changes    `json:"changes"`
	Committer string     `json:"committer"`
	ProjectID int        `json:"project_id"`
	Revision  string     `json:"revision"`
	TicketID  int        `json:"ticket_id"`
	Title     string     `json:"title"`
	UserID    int        `json:"user_id"`
}

type Changesets []*Changeset

type ChangesetCreate struct {
	Body      string     `json:"body"`
	ChangedAt *time.Time `json:"changed_at"`
	Changes   Changes    `json:"changes"`
	Revision  string     `json:"revision"`
	Title     string     `json:"title"`
	UserID    int        `json:"user_id"`
}

type changesetRequest struct {
	Changeset interface{} `json:"changeset"`
}

func (cr *changesetRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(cr)
}

type changesetResponse struct {
	Changeset *Changeset `json:"changeset"`
}

func (cr *changesetResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(cr)
}

type changesetsResponse struct {
	ChangesetResponse []*changesetResponse `json:"changesets"`
}

func (csr *changesetsResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(csr)
}

func (csr *changesetsResponse) changesets() Changesets {
	cs := make(Changesets, 0, len(csr.ChangesetResponse))
	for _, c := range csr.ChangesetResponse {
		cs = append(cs, c.Changeset)
	}

	return cs
}

func (s *Service) List() (Changesets, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	csresp := &changesetsResponse{}
	err = csresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return csresp.changesets(), nil
}

func (s *Service) New() (*Changeset, error) {
	return s.Get("new")
}

func (s *Service) Get(revision string) (*Changeset, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+"/"+revision+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	cresp := &changesetResponse{}
	err = cresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return cresp.Changeset, nil
}

// Only the fields in ChangesetCreate can be set.
func (s *Service) Create(c *Changeset) (*Changeset, error) {
	creq := &changesetRequest{
		Changeset: &ChangesetCreate{
			Body:      c.Body,
			ChangedAt: c.ChangedAt,
			Changes:   c.Changes,
			Revision:  c.Revision,
			Title:     c.Title,
			UserID:    c.UserID,
		},
	}

	buf := &bytes.Buffer{}
	err := creq.Encode(buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.s.RoundTrip("POST", s.basePath+".json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	cresp := &changesetResponse{
		Changeset: c,
	}
	err = cresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Service) Delete(revision string) error {
	resp, err := s.s.RoundTrip("DELETE", s.basePath+"/"+revision+".json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return err
	}

	return nil
}
// Copyright 2022 European Digital Reading Lab. All rights reserved.
// Use of this source code is governed by a BSD-style license
// specified in the Github project LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/edrlab/lcp-server/pkg/stor"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// ListPublications lists all publications present in the database.
func (h *APIHandler) ListPublications(w http.ResponseWriter, r *http.Request) {
	publications, err := h.Store.Publication().ListAll()
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	if err := render.RenderList(w, r, NewPublicationListResponse(publications)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// SearchPublications searches publications corresponding to a specific criteria.
func (h *APIHandler) SearchPublications(w http.ResponseWriter, r *http.Request) {
	var publications *[]stor.Publication
	var err error

	// by format
	if format := r.URL.Query().Get("format"); format != "" {
		var contentType string
		switch format {
		case "epub":
			contentType = "application/epub+zip"
		case "lcpdf":
			contentType = "application/pdf+lcp"
		case "lcpau":
			contentType = "application/audiobook+lcp"
		case "lcpdi":
			contentType = "application/divina+lcp"
		default:
			err = errors.New("invalid content type query string parameter")
		}
		if contentType != "" {
			publications, err = h.Store.Publication().FindByType(contentType)
		}
	} else {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err := render.RenderList(w, r, NewPublicationListResponse(publications)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CreatePublication adds a new Publication to the database.
func (h *APIHandler) CreatePublication(w http.ResponseWriter, r *http.Request) {

	// get the payload
	data := &PublicationRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	publication := data.Publication

	// db create
	err := h.Store.Publication().Create(publication)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusCreated)
	if err := render.Render(w, r, NewPublicationResponse(publication)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// GetPublication returns a specific publication
func (h *APIHandler) GetPublication(w http.ResponseWriter, r *http.Request) {

	var publication *stor.Publication
	var err error

	if publicationID := chi.URLParam(r, "publicationID"); publicationID != "" {
		publication, err = h.Store.Publication().Get(publicationID)
	} else {
		render.Render(w, r, ErrInvalidRequest(errors.New("missing required publication identifier")))
		return
	}
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err := render.Render(w, r, NewPublicationResponse(publication)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// UpdatePublication updates an existing Publication in the database.
func (h *APIHandler) UpdatePublication(w http.ResponseWriter, r *http.Request) {

	// get the payload
	data := &PublicationRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	publication := data.Publication

	var currentPub *stor.Publication
	var err error

	// get the existing publication
	if publicationID := chi.URLParam(r, "publicationID"); publicationID != "" {
		currentPub, err = h.Store.Publication().Get(publicationID)
	} else {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	// set the gorm fields
	publication.ID = currentPub.ID
	publication.CreatedAt = currentPub.CreatedAt
	//publication.UpdatedAt = currentPub.UpdatedAt
	//publication.DeletedAt = currentPub.DeletedAt

	// db update
	err = h.Store.Publication().Update(publication)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	if err := render.Render(w, r, NewPublicationResponse(publication)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// DeletePublication removes an existing Publication from the database.
func (h *APIHandler) DeletePublication(w http.ResponseWriter, r *http.Request) {

	var publication *stor.Publication
	var err error

	// get the existing publication
	if publicationID := chi.URLParam(r, "publicationID"); publicationID != "" {
		publication, err = h.Store.Publication().Get(publicationID)
	} else {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	// db delete
	err = h.Store.Publication().Delete(publication)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := render.Render(w, r, NewPublicationResponse(publication)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// --
// Request and Response payloads for the REST api.
// --

type omit *struct{}

// PublicationRequest is the request publication payload.
type PublicationRequest struct {
	*stor.Publication
}

// PublicationResponse is the response publication payload.
type PublicationResponse struct {
	*stor.Publication
	ID        omit `json:"ID,omitempty"`
	CreatedAt omit `json:"CreatedAt,omitempty"`
	UpdatedAt omit `json:"UpdatedAt,omitempty"`
	DeletedAt omit `json:"DeletedAt,omitempty"`
}

// NewPublicationListResponse creates a rendered list of publications
func NewPublicationListResponse(publications *[]stor.Publication) []render.Renderer {
	list := []render.Renderer{}
	for i := 0; i < len(*publications); i++ {
		list = append(list, NewPublicationResponse(&(*publications)[i]))
	}
	return list
}

// NewPublicationResponse creates a rendered publication.
func NewPublicationResponse(pub *stor.Publication) *PublicationResponse {
	return &PublicationResponse{Publication: pub}
}

// Bind post-processes requests after unmarshalling.
func (p *PublicationRequest) Bind(r *http.Request) error {
	return p.Publication.Validate()
}

// Render processes responses before marshalling.
func (pub *PublicationResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

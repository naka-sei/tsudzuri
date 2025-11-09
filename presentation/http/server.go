package http

import (
	page "github.com/naka-sei/tsudzuri/presentation/http/page"
	prouter "github.com/naka-sei/tsudzuri/presentation/router"
)

const (
	apiPathPrefix = "/api"
	versionPathV1 = "/v1"
	apiV1BasePath = apiPathPrefix + versionPathV1

	// Page routes
	pageBasePath  = "/pages"
	pageIDPath    = "/{id}"
	pageLinksPath = "/links"

	// User routes
	userBasePath = "/users"
)

type Server struct {
	*page.CreateService
	*page.GetService
	*page.ListService
	*page.EditService
	*page.DeleteService
	*page.LinkAddService
	*page.LinkRemoveService
}

func NewServer(
	createPageService *page.CreateService,
	getPageService *page.GetService,
	listPageService *page.ListService,
	editPageService *page.EditService,
	deletePageService *page.DeleteService,
	linkAddPageService *page.LinkAddService,
	linkRemovePageService *page.LinkRemoveService,
) *Server {
	return &Server{
		CreateService:     createPageService,
		GetService:        getPageService,
		ListService:       listPageService,
		EditService:       editPageService,
		DeleteService:     deletePageService,
		LinkAddService:    linkAddPageService,
		LinkRemoveService: linkRemovePageService,
	}
}

func (s *Server) Route(r prouter.Router) {
	// /api/v1
	r.Route(apiV1BasePath, func(r prouter.Router) {
		// /api/v1/pages
		r.Route(pageBasePath, func(r prouter.Router) {
			// POST /api/v1/pages
			r.Post("/", s.Create, prouter.WithStatusCreated())
			// GET /api/v1/pages
			r.Get("/", s.List)
			// /api/v1/pages/{id}
			r.Route(pageIDPath, func(r prouter.Router) {
				// GET /api/v1/pages/{id}
				r.Get("/", s.Get)
				// PUT /api/v1/pages/{id}
				r.Put("/", s.Edit)
				// DELETE /api/v1/pages/{id}
				r.Delete("/", s.Delete, prouter.WithStatusNoContent())
				r.Route(pageLinksPath, func(r prouter.Router) {
					// POST /api/v1/pages/{id}/links
					r.Post("/", s.LinkAdd, prouter.WithStatusCreated())
					// DELETE /api/v1/pages/{id}/links
					r.Delete("/", s.LinkRemove, prouter.WithStatusNoContent())
				})
			})
		})
	})
}
